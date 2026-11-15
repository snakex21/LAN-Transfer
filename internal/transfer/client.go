package transfer

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"lantransfer/internal/filebrowser"
	"lantransfer/internal/i18n"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
)

// ErrNotPaired is returned when the receiver does not accept our token.
var ErrNotPaired = i18n.New("err.notPaired")

// ErrOldPeer means the peer still runs the unencrypted version 1 protocol.
var ErrOldPeer = i18n.New("err.oldPeer")

// RemoteError is a JSON error returned by the receiver.
type RemoteError struct {
	Status int
	Code   string
	Msg    string      // text sent by the receiver (English since 2.1)
	Key    string      // translation key (2.1+ receivers)
	Params i18n.Params // its parameters
}

func (e *RemoteError) Error() string { return i18n.Render(i18n.Current(), e.Message()) }

// Message is the translatable form: the receiver's key when it sent one,
// otherwise a key derived from the error code, otherwise its text.
func (e *RemoteError) Message() i18n.Msg {
	switch {
	case e.Key != "":
		return i18n.Msg{Key: e.Key, Params: e.Params}
	case e.Code == "wrong_code":
		return i18n.M("err.wrongCode")
	case e.Code == "locked":
		return i18n.M("err.locked")
	case e.Code == protocol.CodeCodeRequired:
		return i18n.M("err.codeRequired")
	case e.Code == protocol.CodeChecksum:
		return i18n.M("err.checksum")
	case e.Msg == "":
		return i18n.M("err.remoteHTTP", "status", fmt.Sprint(e.Status))
	}
	return i18n.Text(e.Msg)
}

// Session is an authenticated connection to one receiver.
type Session struct {
	Address     string // host:port
	Token       string
	Fingerprint string
	PeerName    string
	client      *http.Client
}

// Transport tuning: large buffers matter for multi-gigabyte files, HTTP/2 is
// avoided because its flow-control windows cap single-stream throughput.
func newHTTPClient(fingerprint string, seen func(string), timeout time.Duration) *http.Client {
	transport := &http.Transport{
		TLSClientConfig:       security.ClientConfig(fingerprint, seen),
		MaxIdleConnsPerHost:   16,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 0,
		WriteBufferSize:       256 << 10,
		ReadBufferSize:        64 << 10,
		DisableCompression:    true,
		TLSNextProto:          map[string]func(string, *tls.Conn) http.RoundTripper{},
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

// FetchInfo reads the public information of a peer and its certificate
// fingerprint.
func FetchInfo(ctx context.Context, address string) (protocol.Info, error) {
	var seen string
	client := newHTTPClient("", func(fp string) { seen = fp }, 5*time.Second)
	defer client.CloseIdleConnections()
	var info protocol.Info
	if err := doJSON(ctx, client, http.MethodGet, "https://"+address+protocol.PathInfo, "", nil, &info); err != nil {
		return info, wrapConnError(err)
	}
	info.Fingerprint = seen
	return info, nil
}

// Pair proves knowledge of the receiver's code and pins its certificate.
func Pair(ctx context.Context, address, code string, self protocol.Info) (*Session, error) {
	var seen string
	pairClient := newHTTPClient("", func(fp string) { seen = fp }, 10*time.Second)
	defer pairClient.CloseIdleConnections()
	var response protocol.PairResponse
	request := protocol.PairRequest{Code: code, DeviceID: self.ID, Name: self.Name}
	if err := doJSON(ctx, pairClient, http.MethodPost, "https://"+address+protocol.PathPair, "", request, &response); err != nil {
		return nil, wrapConnError(err)
	}
	if response.Token == "" || seen == "" {
		return nil, i18n.New("err.noToken")
	}
	return &Session{
		Address: address, Token: response.Token, Fingerprint: seen, PeerName: response.Name,
		client: newHTTPClient(seen, nil, 0),
	}, nil
}

// Browse lists folders on the receiver.
func (s *Session) Browse(ctx context.Context, path string) (filebrowser.Listing, error) {
	var listing filebrowser.Listing
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := doJSON(ctx, s.client, http.MethodGet, s.url(protocol.PathBrowse, url.Values{"path": {path}}), s.Token, nil, &listing)
	return listing, wrapConnError(err)
}

// Close releases idle connections.
func (s *Session) Close() { s.client.CloseIdleConnections() }

func (s *Session) url(path string, query url.Values) string {
	return "https://" + s.Address + path + "?" + query.Encode()
}

func (s *Session) offset(ctx context.Context, query url.Values) (protocol.OffsetResponse, error) {
	var response protocol.OffsetResponse
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	err := doJSON(ctx, s.client, http.MethodGet, s.url(protocol.PathOffset, query), s.Token, nil, &response)
	return response, err
}

// upload streams body (the file from `offset` on) and sends the SHA-256 of
// the complete file as an HTTP trailer once the body has been read.
func (s *Session) upload(ctx context.Context, query url.Values, body io.Reader, digest func() string) (protocol.UploadResponse, error) {
	var response protocol.UploadResponse
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url(protocol.PathUpload, query), nil)
	if err != nil {
		return response, err
	}
	request.Trailer = http.Header{protocol.TrailerSHA256: nil}
	request.Body = &trailerBody{Reader: body, done: func() { request.Trailer.Set(protocol.TrailerSHA256, digest()) }}
	request.ContentLength = -1
	request.Header.Set("Authorization", "Bearer "+s.Token)
	request.Header.Set("Content-Type", "application/octet-stream")
	resp, err := s.client.Do(request)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	if err := decodeResponse(resp, &response); err != nil {
		return response, err
	}
	return response, nil
}

// trailerBody calls done exactly once, when the wrapped reader hits EOF.
type trailerBody struct {
	io.Reader
	once sync.Once
	done func()
}

func (b *trailerBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	if err == io.EOF {
		b.once.Do(b.done)
	}
	return n, err
}

func (b *trailerBody) Close() error { return nil }

func doJSON(ctx context.Context, client *http.Client, method, endpoint, token string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	if in != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return decodeResponse(resp, out)
}

func decodeResponse(resp *http.Response, out any) error {
	if resp.StatusCode/100 != 2 {
		var e protocol.ErrorResponse
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		_ = json.Unmarshal(data, &e)
		if resp.StatusCode == http.StatusUnauthorized {
			return ErrNotPaired
		}
		if e.Error == "" {
			e.Error = strings.TrimSpace(string(data))
		}
		return &RemoteError{Status: resp.StatusCode, Code: e.Code, Msg: e.Error, Key: e.Key, Params: e.Params}
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(out)
}

// IsCodeRequired reports a receiver that requires the pairing code (the
// sender tried without one).
func IsCodeRequired(err error) bool {
	var remote *RemoteError
	return errors.As(err, &remote) && remote.Code == protocol.CodeCodeRequired
}

// wrapConnError turns low-level failures into messages people understand.
func wrapConnError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case errors.Is(err, ErrNotPaired):
		return err
	case errors.Is(err, security.ErrFingerprintMismatch):
		// Drop the "Post https://…: tls: …" wrapping, keep the localized message.
		var loc i18n.Localized
		if i18n.AsLocalized(err, &loc) {
			return loc
		}
		return err
	case strings.Contains(msg, "server gave HTTP response to HTTPS client"),
		strings.Contains(msg, "first record does not look like a TLS handshake"):
		return ErrOldPeer
	case strings.Contains(msg, "connection refused"), strings.Contains(msg, "actively refused"),
		strings.Contains(msg, "No connection could be made"):
		return i18n.Wrap(err, "err.peerNotResponding", "error", err.Error())
	}
	var remote *RemoteError
	if errors.As(err, &remote) {
		return err
	}
	return err
}
