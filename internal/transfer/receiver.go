package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"lantransfer/internal/fsafe"
	"lantransfer/internal/protocol"
)

// partialDirName holds unfinished uploads inside the destination folder, so a
// cancelled transfer never leaves a truncated file under its real name.
const partialDirName = ".lantransfer-partial"

// spaceMargin is kept free on the receiving volume.
const spaceMargin = 64 << 20

const copyBufferSize = 1 << 20

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

type httpError struct {
	status int
	code   string
	msg    string
}

func (e *httpError) Error() string { return e.msg }

func badRequest(code, format string, args ...any) *httpError {
	return &httpError{status: http.StatusBadRequest, code: code, msg: fmt.Sprintf(format, args...)}
}

func (r *Receiver) parseRequest(req *http.Request) (*fileRequest, *httpError) {
	r.once.Do(func() {
		r.busy, r.done, r.topNames = map[string]bool{}, map[string]string{}, map[string]string{}
	})
	q := req.URL.Query()
	f := &fileRequest{transfer: q.Get("transfer"), rel: q.Get("path"), isDir: q.Get("dir") == "1", jobName: q.Get("job_name")}
	if len(f.transfer) < 8 || len(f.transfer) > 64 || strings.Trim(f.transfer, "0123456789abcdef") != "" {
		return nil, badRequest("", "nieprawidłowy identyfikator transferu")
	}
	osRel, err := fsafe.CleanRelPath(f.rel)
	if err != nil {
		return nil, badRequest(protocol.CodeBadName, "%v", err)
	}
	f.osRel = osRel
	if f.size, err = strconv.ParseInt(q.Get("size"), 10, 64); err != nil || f.size < 0 {
		if !f.isDir {
			return nil, badRequest("", "nieprawidłowy rozmiar pliku")
		}
	}
	if v := q.Get("offset"); v != "" {
		if f.offset, err = strconv.ParseInt(v, 10, 64); err != nil || f.offset < 0 || f.offset > f.size {
			return nil, badRequest(protocol.CodeOffset, "nieprawidłowe przesunięcie")
		}
	}
	f.jobSize, _ = strconv.ParseInt(q.Get("job_size"), 10, 64)
	f.jobFiles, _ = strconv.Atoi(q.Get("job_files"))

	f.destDir = r.Directory
	if dest := q.Get("dest"); dest != "" {
		clean := filepath.Clean(dest)
		info, err := os.Stat(clean)
		if !filepath.IsAbs(clean) || err != nil || !info.IsDir() {
			return nil, badRequest("", "nieprawidłowy folder docelowy")
		}
		f.destDir = clean
	}
	if r.MaxFileSize > 0 && f.size > r.MaxFileSize {
		return nil, &httpError{status: http.StatusRequestEntityTooLarge, code: protocol.CodeTooLarge,
			msg: fmt.Sprintf("plik przekracza limit %d bajtów", r.MaxFileSize)}
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
		writeError(w, &httpError{status: http.StatusMethodNotAllowed, msg: "POST only"})
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
			writeError(w, &httpError{status: http.StatusInternalServerError, msg: err.Error()})
			return
		}
		r.Incoming.dirDone(peer, f)
		writeJSON(w, http.StatusOK, protocol.UploadResponse{SavedAs: final})
		return
	}

	if !r.acquire(f.key()) {
		writeError(w, &httpError{status: http.StatusConflict, code: protocol.CodeOffset, msg: "ten plik jest właśnie odbierany"})
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
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	hideFile(filepath.Dir(part))

	var current int64
	if info, err := os.Stat(part); err == nil {
		current = info.Size()
	}
	if f.offset > current {
		return "", &httpError{status: http.StatusConflict, code: protocol.CodeOffset,
			msg: fmt.Sprintf("odbiorca ma tylko %d bajtów tego pliku", current)}
	}
	if err := r.ensureSpace(f.destDir, f.size-f.offset); err != nil {
		return "", err
	}

	file, err := os.OpenFile(part, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	closed := false
	defer func() {
		if !closed {
			file.Close()
		}
	}()
	if err := file.Truncate(f.offset); err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	hasher := sha256.New()
	if f.offset > 0 {
		if _, err := io.CopyN(hasher, file, f.offset); err != nil {
			return "", &httpError{status: http.StatusInternalServerError, msg: "nie można odczytać części pliku: " + err.Error()}
		}
	}
	if _, err := file.Seek(f.offset, io.SeekStart); err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}

	remaining := f.size - f.offset
	progress := r.Incoming.start(peer, f)
	body := io.LimitReader(req.Body, remaining+1)
	written, copyErr := copyWithProgress(io.MultiWriter(file, hasher), body, progress)
	switch {
	case copyErr != nil && isNoSpace(copyErr):
		return "", &httpError{status: http.StatusInsufficientStorage, code: protocol.CodeNoSpace, msg: "zabrakło miejsca na dysku odbiorcy"}
	case copyErr != nil:
		// Connection dropped or sender cancelled: keep the part for resuming.
		return "", &httpError{status: http.StatusBadRequest, msg: "przerwano przesyłanie: " + copyErr.Error()}
	case written > remaining:
		_ = file.Truncate(f.offset)
		return "", badRequest("", "nadawca wysłał więcej danych niż zapowiedział")
	case written < remaining:
		return "", badRequest("", "przesyłanie niekompletne (%d z %d bajtów)", f.offset+written, f.size)
	}

	expected := strings.ToLower(req.Trailer.Get(protocol.TrailerSHA256))
	actual := hex.EncodeToString(hasher.Sum(nil))
	if expected != actual {
		file.Close()
		closed = true
		_ = os.Remove(part)
		return "", &httpError{status: http.StatusUnprocessableEntity, code: protocol.CodeChecksum,
			msg: "suma kontrolna SHA-256 się nie zgadza, plik odrzucono"}
	}
	if err := file.Close(); err != nil {
		closed = true
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	closed = true

	final, err := r.finalPath(f)
	if err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	if err := os.MkdirAll(filepath.Dir(final), 0o755); err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
	}
	r.mu.Lock()
	final = fsafe.UniquePath(final)
	err = os.Rename(part, final)
	if err == nil {
		r.done[f.key()] = final
	}
	r.mu.Unlock()
	if err != nil {
		return "", &httpError{status: http.StatusInternalServerError, msg: err.Error()}
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
		return "", errors.New("ścieżka poza folderem docelowym")
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
		return &httpError{status: http.StatusInsufficientStorage, code: protocol.CodeNoSpace,
			msg: fmt.Sprintf("za mało miejsca na dysku odbiorcy (potrzeba %s, wolne %s)", FormatBytes(need), FormatBytes(int64(free)))}
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
		return errno == 112 || errno == 39 || errno == syscall.ENOSPC
	}
	return false
}

func copyWithProgress(dst io.Writer, src io.Reader, progress func(int64)) (int64, error) {
	buf := make([]byte, copyBufferSize)
	var total int64
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			written, writeErr := dst.Write(buf[:n])
			total += int64(written)
			if progress != nil {
				progress(int64(written))
			}
			if writeErr != nil {
				return total, writeErr
			}
		}
		if readErr == io.EOF {
			return total, nil
		}
		if readErr != nil {
			return total, readErr
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, err *httpError) {
	writeJSON(w, err.status, protocol.ErrorResponse{Error: err.msg, Code: err.code})
}

// WriteError sends a JSON error (used by the service package).
func WriteError(w http.ResponseWriter, status int, code, msg string) {
	writeError(w, &httpError{status: status, code: code, msg: msg})
}

// WriteJSON sends a JSON value with status 200.
func WriteJSON(w http.ResponseWriter, value any) { writeJSON(w, http.StatusOK, value) }
