package transfer

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"lantransfer/internal/protocol"
)

// DefaultParallel is the number of files uploaded at once. A single large file
// always uses one stream; parallelism helps folders with many small files.
const DefaultParallel = 4

// Entry is one file or directory of a planned transfer.
type Entry struct {
	Source string // absolute local path
	Rel    string // slash-separated path sent to the receiver
	Size   int64
	IsDir  bool
}

// Plan is everything needed to send one selected item (file or folder).
type Plan struct {
	ID         string
	Name       string
	Source     string
	Entries    []Entry
	TotalBytes int64
	TotalFiles int
	Skipped    []string // symlinks and special files that are not sent
}

// PlanPath walks path. The job ID is derived from the source, its size and
// modification times and the destination, so re-sending the same unchanged
// item resumes an interrupted transfer instead of starting over.
func PlanPath(path, address, dest string) (*Plan, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	plan := &Plan{Name: filepath.Base(abs), Source: abs}
	fingerprint := sha256.New()
	fmt.Fprintf(fingerprint, "%s|%s|%s|", abs, address, dest)

	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%s nie jest zwykłym plikiem", abs)
		}
		plan.Entries = []Entry{{Source: abs, Rel: plan.Name, Size: info.Size()}}
		plan.TotalBytes, plan.TotalFiles = info.Size(), 1
		fmt.Fprintf(fingerprint, "%d|%d", info.Size(), info.ModTime().UnixNano())
	} else {
		root := filepath.Dir(abs)
		err = filepath.WalkDir(abs, func(p string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if d.IsDir() {
				plan.Entries = append(plan.Entries, Entry{Source: p, Rel: rel, IsDir: true})
				return nil
			}
			if !d.Type().IsRegular() {
				plan.Skipped = append(plan.Skipped, p)
				return nil
			}
			fi, err := d.Info()
			if err != nil {
				return err
			}
			plan.Entries = append(plan.Entries, Entry{Source: p, Rel: rel, Size: fi.Size()})
			plan.TotalBytes += fi.Size()
			plan.TotalFiles++
			fmt.Fprintf(fingerprint, "%s|%d|%d|", rel, fi.Size(), fi.ModTime().UnixNano())
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	plan.ID = hex.EncodeToString(fingerprint.Sum(nil)[:16])
	return plan, nil
}

// FileProgress is the state of one file currently being uploaded.
type FileProgress struct {
	Rel  string `json:"rel"`
	Size int64  `json:"size"`
	Done int64  `json:"done"`
}

// Progress is updated concurrently by the upload workers.
type Progress struct {
	DoneBytes atomic.Int64
	DoneFiles atomic.Int64
	Resumed   atomic.Int64 // bytes skipped because the receiver already had them

	mu     sync.Mutex
	active map[string]*FileProgress
}

func (p *Progress) begin(e Entry, offset int64) *FileProgress {
	fp := &FileProgress{Rel: e.Rel, Size: e.Size, Done: offset}
	p.mu.Lock()
	if p.active == nil {
		p.active = map[string]*FileProgress{}
	}
	p.active[e.Rel] = fp
	p.mu.Unlock()
	return fp
}

func (p *Progress) add(fp *FileProgress, n int64) {
	p.DoneBytes.Add(n)
	p.mu.Lock()
	fp.Done += n
	p.mu.Unlock()
}

func (p *Progress) end(fp *FileProgress) {
	p.mu.Lock()
	delete(p.active, fp.Rel)
	p.mu.Unlock()
}

// Active returns the files in flight, sorted by name.
func (p *Progress) Active() []FileProgress {
	p.mu.Lock()
	defer p.mu.Unlock()
	result := make([]FileProgress, 0, len(p.active))
	for _, fp := range p.active {
		result = append(result, *fp)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Rel < result[j].Rel })
	return result
}

// Send uploads every entry of plan. Directories go first, then files with up
// to `parallel` concurrent streams. The first error cancels the remaining work;
// the receiver keeps partial files so calling Send again resumes.
func (s *Session) Send(ctx context.Context, plan *Plan, dest string, parallel int, progress *Progress) error {
	if progress == nil {
		progress = &Progress{}
	}
	if parallel < 1 {
		parallel = 1
	}
	base := url.Values{
		"transfer":  {plan.ID},
		"job_name":  {plan.Name},
		"job_size":  {strconv.FormatInt(plan.TotalBytes, 10)},
		"job_files": {strconv.Itoa(plan.TotalFiles)},
	}
	if dest != "" {
		base.Set("dest", dest)
	}

	var files []Entry
	for _, e := range plan.Entries {
		if !e.IsDir {
			files = append(files, e)
			continue
		}
		q := cloneValues(base)
		q.Set("path", e.Rel)
		q.Set("dir", "1")
		if _, err := s.upload(ctx, q, eofReader{}, func() string { return "" }); err != nil {
			return fmt.Errorf("%s: %w", e.Rel, wrapConnError(err))
		}
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	jobs := make(chan Entry)
	var wg sync.WaitGroup
	for i := 0; i < min(parallel, max(len(files), 1)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for e := range jobs {
				if err := s.sendFileWithRetry(ctx, base, e, progress); err != nil {
					cancel(fmt.Errorf("%s: %w", e.Rel, err))
					return
				}
				progress.DoneFiles.Add(1)
			}
		}()
	}
feed:
	for _, e := range files {
		select {
		case jobs <- e:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	if err := context.Cause(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return ctx.Err()
}

const maxAttempts = 3

// smallFileLimit: files up to this size skip the resume-offset query.
const smallFileLimit = 1 << 20

func (s *Session) sendFileWithRetry(ctx context.Context, base url.Values, e Entry, progress *Progress) error {
	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = s.sendFile(ctx, base, e, progress)
		if err == nil || ctx.Err() != nil || !retryable(err) {
			break
		}
		time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return wrapConnError(err)
}

func retryable(err error) bool {
	var remote *RemoteError
	if errors.As(err, &remote) {
		return remote.Status == 409 || remote.Status >= 500 && remote.Status != 507
	}
	var netErr net.Error
	return errors.As(err, &netErr) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
}

func (s *Session) sendFile(ctx context.Context, base url.Values, e Entry, progress *Progress) error {
	q := cloneValues(base)
	q.Set("path", e.Rel)
	q.Set("size", strconv.FormatInt(e.Size, 10))

	// Small files are simply re-sent: one request instead of two matters for
	// folders with thousands of files, and the receiver ignores duplicates of
	// files it already stored for this job.
	var state protocol.OffsetResponse
	if e.Size > smallFileLimit {
		var err error
		if state, err = s.offset(ctx, q); err != nil {
			return err
		}
	}
	if state.Done {
		progress.DoneBytes.Add(e.Size)
		progress.Resumed.Add(e.Size)
		return nil
	}
	offset := state.Offset
	if offset > e.Size {
		offset = 0
	}

	file, err := os.Open(e.Source)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() != e.Size {
		return fmt.Errorf("plik zmienił się podczas wysyłania (%s → %s)", FormatBytes(e.Size), FormatBytes(info.Size()))
	}

	hasher := sha256.New()
	if offset > 0 {
		if _, err := io.CopyN(hasher, file, offset); err != nil {
			return err
		}
		progress.Resumed.Add(offset)
	}
	fp := progress.begin(e, offset)
	defer progress.end(fp)
	progress.DoneBytes.Add(offset)
	q.Set("offset", strconv.FormatInt(offset, 10))

	body := &countingReader{src: file, hash: hasher, onRead: func(n int64) { progress.add(fp, n) }}
	_, err = s.upload(ctx, q, body, func() string { return hex.EncodeToString(hasher.Sum(nil)) })
	if err != nil && ctx.Err() == nil {
		// Roll back this attempt's bytes; a retry asks the receiver again.
		// After a cancellation the counters stay: the receiver keeps the part.
		progress.DoneBytes.Add(-(body.n + offset))
		progress.Resumed.Add(-offset)
	}
	return err
}

// countingReader feeds the hash and the progress counter.
type countingReader struct {
	src    io.Reader
	hash   hash.Hash
	onRead func(int64)
	n      int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.src.Read(p)
	if n > 0 {
		r.hash.Write(p[:n])
		r.n += int64(n)
		r.onRead(int64(n))
	}
	return n, err
}

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) { return 0, io.EOF }

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v)+4)
	for k, vals := range v {
		out[k] = append([]string(nil), vals...)
	}
	return out
}
