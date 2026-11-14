package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"lantransfer/internal/fsafe"
	"lantransfer/internal/i18n"
	"lantransfer/internal/protocol"
)

// partialDirName holds unfinished uploads inside the destination folder, so a
// cancelled transfer never leaves a truncated file under its real name.
const partialDirName = ".lantransfer-partial"

// spaceMargin is kept free on the receiving volume.
const spaceMargin = 64 << 20

// Receiver stores files uploaded by paired peers.
type Receiver struct {
	Directory   string                           // default destination
	MaxFileSize int64                            // 0 = unlimited
	FreeSpace   func(dir string) (uint64, error) // injectable for tests
	Incoming    *IncomingTracker                 // optional progress/history sink

	once     sync.Once
	mu       sync.Mutex
	busy     map[string]bool   // part files currently being written
	done     map[string]string // finished uploads (for resume skip)
	topNames map[string]string // folder name chosen for a transfer
}

// fileRequest holds validated upload/offset parameters.
type fileRequest struct {
	transfer string
	rel      string // slash form, as sent
	osRel    string // validated OS form
	size     int64
	offset   int64
	destDir  string
	isDir    bool
	jobName  string
	jobSize  int64
	jobFiles int
}

// httpError is sent to the peer as protocol.ErrorResponse: English text plus
// the translation key and parameters.
type httpError struct {
	status int
	code   string
	msg    i18n.Msg
}

func (e *httpError) Error() string { return i18n.Render(i18n.Current(), e.msg) }

func newHTTPError(status int, code, key string, kv ...any) *httpError {
	return &httpError{status: status, code: code, msg: i18n.M(key, kv...)}
}

func badRequest(code, key string, kv ...any) *httpError {
	return newHTTPError(http.StatusBadRequest, code, key, kv...)
}

// internalError reports an OS error (its text is not translated).
func internalError(err error) *httpError {
	return &httpError{status: http.StatusInternalServerError, msg: i18n.MsgOf(err)}
}

func (r *Receiver) parseRequest(req *http.Request) (*fileRequest, *httpError) {
	r.once.Do(func() {
		r.busy, r.done, r.topNames = map[string]bool{}, map[string]string{}, map[string]string{}
	})
	q := req.URL.Query()
	f := &fileRequest{transfer: q.Get("transfer"), rel: q.Get("path"), isDir: q.Get("dir") == "1", jobName: q.Get("job_name")}
	if len(f.transfer) < 8 || len(f.transfer) > 64 || strings.Trim(f.transfer, "0123456789abcdef") != "" {
		return nil, badRequest("", "err.badTransferID")
	}
	osRel, err := fsafe.CleanRelPath(f.rel)
	if err != nil {
		return nil, &httpError{status: http.StatusBadRequest, code: protocol.CodeBadName, msg: i18n.MsgOf(err)}
	}
	f.osRel = osRel
	if f.size, err = strconv.ParseInt(q.Get("size"), 10, 64); err != nil || f.size < 0 {
		if !f.isDir {
			return nil, badRequest("", "err.badFileSize")
		}
	}
	if v := q.Get("offset"); v != "" {
		if f.offset, err = strconv.ParseInt(v, 10, 64); err != nil || f.offset < 0 || f.offset > f.size {
			return nil, badRequest(protocol.CodeOffset, "err.badOffset")
		}
	}
	f.jobSize, _ = strconv.ParseInt(q.Get("job_size"), 10, 64)
	f.jobFiles, _ = strconv.Atoi(q.Get("job_files"))

	f.destDir = r.Directory
	if dest := q.Get("dest"); dest != "" {
		clean := filepath.Clean(dest)
		info, err := os.Stat(clean)
		if !filepath.IsAbs(clean) || err != nil || !info.IsDir() {
			return nil, badRequest("", "err.badDestination")
		}
		f.destDir = clean
	}
	if r.MaxFileSize > 0 && f.size > r.MaxFileSize {
		return nil, newHTTPError(http.StatusRequestEntityTooLarge, protocol.CodeTooLarge, "err.fileTooLarge", "max", r.MaxFileSize)
	}
	return f, nil
}

func (f *fileRequest) key() string {
	return f.destDir + "\x00" + f.transfer + "\x00" + f.rel
}

func (f *fileRequest) partPath() string {
	sum := sha256.Sum256([]byte(f.key()))
	return filepath.Join(f.destDir, partialDirName, hex.EncodeToString(sum[:12])+".part")
}

// HandleOffset reports how much of a file is already stored, for resuming.
func (r *Receiver) HandleOffset(w http.ResponseWriter, req *http.Request) {
	f, herr := r.parseRequest(req)
	if herr != nil {
		writeError(w, herr)
		return
	}
	_, done := r.completed(f)
	var offset int64
	if info, err := os.Stat(f.partPath()); err == nil && !done {
		offset = info.Size()
	}
	writeJSON(w, http.StatusOK, protocol.OffsetResponse{Offset: offset, Done: done})
}

// HandleUpload stores one file (or creates one directory).
func (r *Receiver) HandleUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		writeError(w, &httpError{status: http.StatusMethodNotAllowed, msg: i18n.Text("POST only")})
		return
	}
	f, herr := r.parseRequest(req)
	if herr != nil {
		writeError(w, herr)
		return
	}
	peer := PeerFromContext(req.Context())
	if f.isDir {
		final, err := r.finalPath(f)
		if err == nil {
			err = os.MkdirAll(final, 0o755)
		}
		if err != nil {
			writeError(w, internalError(err))
			return
		}
		r.Incoming.dirDone(peer, f)
		writeJSON(w, http.StatusOK, protocol.UploadResponse{SavedAs: final})
		return
	}

	if !r.acquire(f.key()) {
		writeError(w, newHTTPError(http.StatusConflict, protocol.CodeOffset, "err.fileBusy"))
		return
	}
	defer r.release(f.key())

	if saved, ok := r.completed(f); ok {
		// Already received (e.g. a resumed folder job re-sends small files
		// without asking for the offset first): keep the existing copy.
		_, _ = io.Copy(io.Discard, io.LimitReader(req.Body, f.size-f.offset+1))
		writeJSON(w, http.StatusOK, protocol.UploadResponse{SavedAs: saved})
		return
	}
	saved, herr := r.receive(req, f, peer)
	if herr != nil {
		r.Incoming.failed(peer, f, herr.msg)
		writeError(w, herr)
		return
	}
	writeJSON(w, http.StatusOK, protocol.UploadResponse{SavedAs: saved})
}

func (r *Receiver) receive(req *http.Request, f *fileRequest, peer string) (string, *httpError) {
	part := f.partPath()
	if err := os.MkdirAll(filepath.Dir(part), 0o755); err != nil {
		return "", internalError(err)
	}
	hideFile(filepath.Dir(part))

	var current int64
	if info, err := os.Stat(part); err == nil {
		current = info.Size()
	}
	if f.offset > current {
		return "", newHTTPError(http.StatusConflict, protocol.CodeOffset, "err.receiverHasLess", "have", current)
	}
	if err := r.ensureSpace(f.destDir, f.size-f.offset); err != nil {
		return "", err
	}

	file, err := os.OpenFile(part, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", internalError(err)
	}
	closed := false
	defer func() {
		if !closed {
			file.Close()
		}
	}()
	if err := file.Truncate(f.offset); err != nil {
		return "", internalError(err)
	}
	hasher := sha256.New()
	if f.offset > 0 {
		if _, err := io.CopyN(hasher, file, f.offset); err != nil {
			return "", newHTTPError(http.StatusInternalServerError, "", "err.readPart", "error", err)
		}
	}
	if _, err := file.Seek(f.offset, io.SeekStart); err != nil {
		return "", internalError(err)
	}

	remaining := f.size - f.offset
	progress := r.Incoming.start(peer, f)
	body := io.LimitReader(req.Body, remaining+1)
	var written int64
	var copyErr error
	if remaining > chunkSize {
		written, copyErr = copyHashing(file, hasher, body, progress)
	} else {
		written, copyErr = copySmall(io.MultiWriter(file, hasher), body, progress)
	}
	switch {
	case copyErr != nil && isNoSpace(copyErr):
		return "", newHTTPError(http.StatusInsufficientStorage, protocol.CodeNoSpace, "err.diskFull")
	case copyErr != nil:
		// Connection dropped or sender cancelled: keep the part for resuming.
		return "", badRequest("", "err.uploadInterrupted", "error", copyErr)
	case written > remaining:
		_ = file.Truncate(f.offset)
		return "", badRequest("", "err.tooMuchData")
	case written < remaining:
		return "", badRequest("", "err.incomplete", "done", f.offset+written, "total", f.size)
	}

	expected := strings.ToLower(req.Trailer.Get(protocol.TrailerSHA256))
	actual := hex.EncodeToString(hasher.Sum(nil))
	if expected != actual {
		file.Close()
		closed = true
		_ = os.Remove(part)
		return "", newHTTPError(http.StatusUnprocessableEntity, protocol.CodeChecksum, "err.checksum")
	}
	if err := file.Close(); err != nil {
		closed = true
		return "", internalError(err)
	}
	closed = true

	final, err := r.finalPath(f)
	if err != nil {
		return "", internalError(err)
	}
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		return "", internalError(err)
	}
	r.mu.Lock()
	final = fsafe.UniquePath(final)
	err = os.Rename(part, final)
	if err == nil {
		r.done[f.key()] = final
	}
	r.mu.Unlock()
	if err != nil {
		return "", internalError(err)
	}
	if r.Incoming.fileDone(peer, f) || r.Incoming == nil {
		r.removePartialDir(f.destDir, f.key())
	}
	return final, nil
}

// finalPath maps the relative path onto the destination folder. The first path
// component of a folder transfer is renamed once ("Folder (1)") when a folder of
// that name already exists, so two transfers never get mixed together.
func (r *Receiver) finalPath(f *fileRequest) (string, error) {
	parts := strings.SplitN(f.osRel, string(filepath.Separator), 2)
	top := parts[0]
	if len(parts) == 2 || f.isDir {
		key := f.destDir + "\x00" + f.transfer + "\x00" + top
		r.mu.Lock()
		resolved, ok := r.topNames[key]
		if !ok {
			resolved = filepath.Base(fsafe.UniquePath(filepath.Join(f.destDir, top)))
			if err := os.MkdirAll(filepath.Join(f.destDir, resolved), 0o755); err != nil {
				r.mu.Unlock()
				return "", err
			}
			r.topNames[key] = resolved
		}
		r.mu.Unlock()
		top = resolved
	}
	final := filepath.Join(f.destDir, top)
	if len(parts) == 2 {
		final = filepath.Join(final, parts[1])
	}
	if !fsafe.Within(f.destDir, final) {
		return "", i18n.New("err.outsideDestination")
	}
	return final, nil
}

func (r *Receiver) ensureSpace(dir string, need int64) *httpError {
	freeSpace := r.FreeSpace
	if freeSpace == nil {
		freeSpace = fsafe.FreeSpace
	}
	free, err := freeSpace(dir)
	if err != nil {
		return nil // unknown: let the write itself fail if needed
	}
	if uint64(need)+spaceMargin > free {
		return newHTTPError(http.StatusInsufficientStorage, protocol.CodeNoSpace, "err.notEnoughSpace", "need", need, "free", int64(free))
	}
	return nil
}

// completed returns where an upload was stored if it finished earlier and the
// file is still there with the expected size.
func (r *Receiver) completed(f *fileRequest) (string, bool) {
	r.mu.Lock()
	saved, ok := r.done[f.key()]
	r.mu.Unlock()
	if !ok {
		return "", false
	}
	if info, err := os.Stat(saved); err != nil || info.Size() != f.size {
		r.mu.Lock()
		delete(r.done, f.key())
		r.mu.Unlock()
		return "", false
	}
	return saved, true
}

// removePartialDir deletes the (empty) partial folder once no upload into
// destDir is running. Removing it after every file raced with parallel uploads.
func (r *Receiver) removePartialDir(destDir, self string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	prefix := destDir + "\x00"
	for key := range r.busy {
		if key != self && strings.HasPrefix(key, prefix) {
			return
		}
	}
	_ = os.Remove(filepath.Join(destDir, partialDirName)) // fails while other parts remain
}

func (r *Receiver) acquire(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.busy[key] {
		return false
	}
	r.busy[key] = true
	return true
}

func (r *Receiver) release(key string) {
	r.mu.Lock()
	delete(r.busy, key)
	r.mu.Unlock()
}

func isNoSpace(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		// ERROR_DISK_FULL (112), ERROR_HANDLE_DISK_FULL (39), ENOSPC
		if runtime.GOOS == "windows" && (errno == 112 || errno == 39) {
			return true
		}
		return errno == syscall.ENOSPC
	}
	return false
}

// copySmall handles files up to one chunk without starting goroutines.
func copySmall(dst io.Writer, src io.Reader, progress func(int64)) (int64, error) {
	c := chunkBuffers.Get().(*chunk)
	defer chunkBuffers.Put(c)
	n, err := io.ReadFull(src, c.buf)
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	if n > 0 {
		if _, werr := dst.Write(c.buf[:n]); werr != nil {
			return int64(n), werr
		}
		if progress != nil {
			progress(int64(n))
		}
	}
	if err == nil && n == len(c.buf) {
		// More data than one chunk: read the rest so the caller can detect it.
		extra, rerr := io.Copy(io.Discard, io.LimitReader(src, 1))
		return int64(n) + extra, rerr
	}
	return int64(n), err
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err *httpError) {
	writeJSON(w, err.status, protocol.ErrorResponse{
		Error: i18n.Render(i18n.Fallback, err.msg), Code: err.code, Key: err.msg.Key, Params: err.msg.Params,
	})
}

// WriteError sends a JSON error (used by the service package).
func WriteError(w http.ResponseWriter, status int, code string, msg i18n.Msg) {
	writeError(w, &httpError{status: status, code: code, msg: msg})
}

// WriteJSON sends a JSON value with status 200.
func WriteJSON(w http.ResponseWriter, value any) { writeJSON(w, http.StatusOK, value) }
