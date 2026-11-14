package transfer

import (
	"context"
	"sort"
	"sync"
	"time"

	"lantransfer/internal/i18n"
)

type peerKey struct{}

// WithPeer stores the authenticated sender name in a request context.
func WithPeer(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, peerKey{}, name)
}

// PeerFromContext returns the sender name set by WithPeer.
func PeerFromContext(ctx context.Context) string {
	name, _ := ctx.Value(peerKey{}).(string)
	return name
}

// incomingStaleAfter closes an incoming job that stopped receiving data.
const incomingStaleAfter = 30 * time.Second

// IncomingJob is the receiver's view of one transfer.
type IncomingJob struct {
	ID         string    `json:"id"`
	Peer       string    `json:"peer"`
	Name       string    `json:"name"`
	TotalBytes int64     `json:"total_bytes"`
	TotalFiles int       `json:"total_files"`
	DoneBytes  int64     `json:"done_bytes"`
	DoneFiles  int       `json:"done_files"`
	Current    string    `json:"current"`
	Directory  string    `json:"directory"`
	Error      string    `json:"error,omitempty"`
	ErrorMsg   *i18n.Msg `json:"error_msg,omitempty"`
	Started    time.Time `json:"started"`
	Updated    time.Time `json:"updated"`
	Speed      float64   `json:"speed"`

	meter speedMeter
	files map[string]int64 // bytes stored per relative path
}

// IncomingTracker aggregates per-file uploads into jobs for the UI and history.
// A nil tracker is valid and ignores everything.
type IncomingTracker struct {
	mu      sync.Mutex
	jobs    map[string]*IncomingJob
	History *History
	now     func() time.Time
}

// NewIncomingTracker creates a tracker that records finished jobs in history.
func NewIncomingTracker(history *History) *IncomingTracker {
	return &IncomingTracker{jobs: map[string]*IncomingJob{}, History: history, now: time.Now}
}

func (t *IncomingTracker) job(peer string, f *fileRequest) *IncomingJob {
	job, ok := t.jobs[f.transfer]
	if !ok {
		now := t.now()
		job = &IncomingJob{ID: f.transfer, Peer: peer, Name: f.jobName, TotalBytes: f.jobSize, TotalFiles: f.jobFiles,
			Directory: f.destDir, Started: now, Updated: now, files: map[string]int64{}}
		if job.Name == "" {
			job.Name = f.rel
		}
		t.jobs[f.transfer] = job
	}
	return job
}

// start registers an upload and returns a progress callback.
func (t *IncomingTracker) start(peer string, f *fileRequest) func(int64) {
	if t == nil {
		return nil
	}
	t.mu.Lock()
	job := t.job(peer, f)
	job.Current = f.rel
	job.Error, job.ErrorMsg = "", nil
	// A resumed file already has `offset` bytes on disk; forget whatever an
	// earlier, interrupted attempt counted beyond that.
	job.DoneBytes += f.offset - job.files[f.rel]
	job.files[f.rel] = f.offset
	rel := f.rel
	t.mu.Unlock()
	return func(n int64) {
		t.mu.Lock()
		job.files[rel] += n
		job.DoneBytes += n
		job.Updated = t.now()
		t.mu.Unlock()
	}
}

// dirDone notes activity for a directory entry (not counted as a file).
func (t *IncomingTracker) dirDone(peer string, f *fileRequest) {
	if t == nil {
		return
	}
	t.mu.Lock()
	t.job(peer, f).Updated = t.now()
	t.mu.Unlock()
}

// fileDone counts a stored file and reports whether its job is complete.
func (t *IncomingTracker) fileDone(peer string, f *fileRequest) bool {
	if t == nil {
		return false
	}
	t.mu.Lock()
	job := t.job(peer, f)
	job.DoneFiles++
	job.Updated = t.now()
	finished := job.TotalFiles > 0 && job.DoneFiles >= job.TotalFiles
	if finished {
		delete(t.jobs, job.ID)
	}
	t.mu.Unlock()
	if finished {
		t.record(job, StatusDone, nil)
	}
	return finished
}

func (t *IncomingTracker) failed(peer string, f *fileRequest, msg i18n.Msg) {
	if t == nil {
		return
	}
	t.mu.Lock()
	job := t.job(peer, f)
	job.Error, job.ErrorMsg = i18n.Render(i18n.Current(), msg), &msg
	job.Updated = t.now()
	t.mu.Unlock()
}

// Snapshot returns active incoming jobs; stale ones are moved to history.
func (t *IncomingTracker) Snapshot() []IncomingJob {
	if t == nil {
		return nil
	}
	now := t.now()
	var stale []*IncomingJob
	t.mu.Lock()
	result := make([]IncomingJob, 0, len(t.jobs))
	for id, job := range t.jobs {
		if now.Sub(job.Updated) > incomingStaleAfter {
			delete(t.jobs, id)
			stale = append(stale, job)
			continue
		}
		job.Speed = job.meter.sample(now, job.DoneBytes)
		copyJob := *job
		copyJob.meter, copyJob.files = speedMeter{}, nil
		result = append(result, copyJob)
	}
	t.mu.Unlock()
	for _, job := range stale {
		msg := job.ErrorMsg
		if msg == nil {
			stopped := i18n.M("incoming.senderStopped")
			msg = &stopped
		}
		t.record(job, StatusInterrupted, msg)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Started.Before(result[j].Started) })
	return result
}

func (t *IncomingTracker) record(job *IncomingJob, status string, msg *i18n.Msg) {
	if t.History == nil {
		return
	}
	entry := HistoryEntry{
		Direction: "received", Name: job.Name, Peer: job.Peer, Bytes: job.DoneBytes, Files: job.DoneFiles,
		Status: status, Started: job.Started, Finished: t.now(), Location: job.Directory,
	}
	if msg != nil {
		entry.Error, entry.ErrorMsg = i18n.Render(i18n.Current(), *msg), msg
	}
	t.History.Add(entry)
}

// speedMeter computes a smoothed transfer rate in bytes per second.
type speedMeter struct {
	lastTime  time.Time
	lastBytes int64
	rate      float64
}

func (m *speedMeter) sample(now time.Time, bytes int64) float64 {
	if m.lastTime.IsZero() {
		m.lastTime, m.lastBytes = now, bytes
		return m.rate
	}
	elapsed := now.Sub(m.lastTime).Seconds()
	if elapsed < 0.2 {
		return m.rate
	}
	instant := float64(bytes-m.lastBytes) / elapsed
	if instant < 0 {
		instant = 0
	}
	if m.rate == 0 {
		m.rate = instant
	} else {
		m.rate = 0.6*m.rate + 0.4*instant
	}
	m.lastTime, m.lastBytes = now, bytes
	return m.rate
}
