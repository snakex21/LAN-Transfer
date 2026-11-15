// Package service runs the HTTPS endpoint that receives files and the UDP
// discovery responder.
package service

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lantransfer/internal/app"
	"lantransfer/internal/config"
	"lantransfer/internal/discovery"
	"lantransfer/internal/filebrowser"
	"lantransfer/internal/i18n"
	"lantransfer/internal/protocol"
	"lantransfer/internal/security"
	"lantransfer/internal/transfer"
)

// Start listens on the configured address and serves in the background.
func Start(state *app.State) (*http.Server, error) {
	listener, err := Listen()
	if err != nil {
		return nil, err
	}
	server := NewServer(state)
	go discovery.Responder{Info: state.Info}.Run()
	go func() {
		if err := server.ServeTLS(listener, "", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = listener.Close()
		}
	}()
	return server, nil
}

// Run serves in the foreground (CLI `serve`).
func Run(state *app.State) error {
	listener, err := Listen()
	if err != nil {
		return err
	}
	go discovery.Responder{Info: state.Info}.Run()
	return NewServer(state).ServeTLS(listener, "", "")
}

// Listen opens the TCP listener on the configured interface and port.
func Listen() (net.Listener, error) {
	return net.Listen("tcp4", net.JoinHostPort(config.BindHost(), strconv.Itoa(config.HTTPPort)))
}

// NewServer builds the HTTPS server. Only /api/info and /api/pair are public;
// browsing folders and uploading require a token obtained with the pairing code.
func NewServer(state *app.State) *http.Server {
	receiver := &transfer.Receiver{Directory: state.ReceiveDir, Incoming: state.Incoming}
	return NewServerWith(state, receiver)
}

// NewServerWith is NewServer with a custom receiver (tests).
func NewServerWith(state *app.State, receiver *transfer.Receiver) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+protocol.PathInfo, func(w http.ResponseWriter, _ *http.Request) {
		transfer.WriteJSON(w, state.Info())
	})
	mux.HandleFunc("POST "+protocol.PathPair, pairHandler(state))
	auth := requirePairing(state.Pairing)
	mux.Handle("GET "+protocol.PathBrowse, auth(http.HandlerFunc(browseHandler)))
	mux.Handle("GET "+protocol.PathOffset, auth(http.HandlerFunc(receiver.HandleOffset)))
	mux.Handle("POST "+protocol.PathUpload, auth(http.HandlerFunc(receiver.HandleUpload)))

	tlsConfig := state.Identity.ServerConfig()
	return &http.Server{
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
		// HTTP/1.1 only: one TCP stream per upload gives the best throughput.
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}
}

func pairHandler(state *app.State) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var request protocol.PairRequest
		if err := json.NewDecoder(io.LimitReader(req.Body, 4096)).Decode(&request); err != nil {
			transfer.WriteError(w, http.StatusBadRequest, "", i18n.M("err.badPairRequest"))
			return
		}
		host, _, _ := net.SplitHostPort(req.RemoteAddr)
		token, err := state.Pairing.Pair(host, request.Code, security.Peer{DeviceID: request.DeviceID, Name: limit(request.Name, 64)})
		switch {
		case errors.Is(err, security.ErrLocked):
			transfer.WriteError(w, http.StatusTooManyRequests, "locked", i18n.MsgOf(err))
		case errors.Is(err, security.ErrCodeRequired):
			transfer.WriteError(w, http.StatusForbidden, protocol.CodeCodeRequired, i18n.MsgOf(err))
		case err != nil:
			transfer.WriteError(w, http.StatusForbidden, "wrong_code", i18n.MsgOf(err))
		default:
			transfer.WriteJSON(w, protocol.PairResponse{Token: token, Name: state.Name})
		}
	}
}

func requirePairing(pairing *security.Pairing) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			token, ok := strings.CutPrefix(req.Header.Get("Authorization"), "Bearer ")
			peer, authorized := pairing.Authorize(token)
			if !ok || !authorized {
				transfer.WriteError(w, http.StatusUnauthorized, protocol.CodeUnauthorized, i18n.M("err.pairingRequired"))
				return
			}
			name := peer.Name
			if name == "" {
				name = peer.Address
			}
			next.ServeHTTP(w, req.WithContext(transfer.WithPeer(req.Context(), name)))
		})
	}
}

func browseHandler(w http.ResponseWriter, req *http.Request) {
	listing, err := filebrowser.Browse(req.URL.Query().Get("path"))
	if err != nil {
		transfer.WriteError(w, http.StatusBadRequest, "", i18n.MsgOf(err))
		return
	}
	transfer.WriteJSON(w, listing)
}

func limit(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
