package transfer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"lantransfer/internal/protocol"
)

// Job is one queued item (file or folder) in the sending queue.
type Job struct {
	ID         string
	Name       string
	Source     string
	Address    string
	PeerName   string
	Dest       string
	Status     string
	Error      string
	Created    time.Time
	Started    time.Time
	Finished   time.Time
	TotalBytes int64
	TotalFiles int

	plan     *Plan
	progress *Progress
	cancel   context.CancelFunc
	meter    speedMeter
}

// JobView is the JSON snapshot shown by the UI.
type JobView struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Source     string         `json:"source"`
	Peer       string         `json:"peer"`
	Address    string         `json:"address"`
	Dest       string         `json:"dest"`
	Status     string         `json:"status"`
	Error      string         `json:"error,omitempty"`
	TotalBytes int64          `json:"total_bytes"`
	TotalFiles int            `json:"total_files"`
	DoneBytes  int64          `json:"done_bytes"`
	DoneFiles  int64          `json:"done_files"`
	Resumed    int64          `json:"resumed_bytes"`
	Speed      float64        `json:"speed"`
	ETA        float64        `json:"eta"`
	Elapsed    float64        `json:"elapsed"`
	Active     []FileProgress `json:"active"`
}

// Manager owns paired sessions and runs the sending queue one job at a time.
type Manager struct {
	Self     protocol.Info
	History  *History
	Parallel int

	mu       sync.Mutex
	sessions map[string]*Session
	jobs     map[string]*Job
	order    []string
	queue    chan *Job
	now      func() time.Time
}

// NewManager starts the queue worker.
func NewManager(self protocol.Info, history *History) *Manager {
	m := &Manager{
		Self: self, History: history, Parallel: DefaultParallel,
		sessions: map[string]*Session{}, jobs: map[string]*Job{},
		queue: make(chan *Job, 1024), now: time.Now,
	}
	go m.worker()
	return m
}

// Pair pairs with the receiver at address and remembers the session.
func (m *Manager) Pair(ctx context.Context, address, code string) (*Session, error) {
	session, err := Pair(ctx, address, code, m.Self)
	if err != nil {
		return nil, err
	}
	m.mu.Lock()
	if old := m.sessions[address]; old != nil {
		old.Close()
	}
	m.sessions[address] = session
	m.mu.Unlock()
	return session, nil
}

// Session returns the paired session for address, if any.
func (m *Manager) Session(address string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sessions[address]
}

// Forget drops a session (e.g. after the receiver rejected our token).
func (m *Manager) Forget(address string) {
	m.mu.Lock()
	if s := m.sessions[address]; s != nil {
		s.Close()
		delete(m.sessions, address)
	}
	m.mu.Unlock()
}

// Paired lists addresses with a session.
func (m *Manager) Paired() map[string]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make(map[string]string, len(m.sessions))
	for address, s := range m.sessions {
		result[address] = s.Fingerprint
	}
	return result
}

// Enqueue plans each path and queues it for sending.
func (m *Manager) Enqueue(paths []string, address, dest string) ([]string, error) {
	session := m.Session(address)
	if session == nil {
		return nil, ErrNotPaired
	}
	ids := make([]string, 0, len(paths))
	for _, path := range paths {
		plan, err := PlanPath(path, address, dest)
		if err != nil {
			return ids, fmt.Errorf("%s: %w", path, err)
		}
		job := &Job{
			ID: plan.ID, Name: plan.Name, Source: plan.Source, Address: address, PeerName: session.PeerName,
			Dest: dest, Status: StatusQueued, Created: m.now(), TotalBytes: plan.TotalBytes,
			TotalFiles: plan.TotalFiles, plan: plan, progress: &Progress{},
		}
		m.mu.Lock()
		if existing := m.jobs[job.ID]; existing != nil && (existing.Status == StatusRunning || existing.Status == StatusQueued) {
			m.mu.Unlock()
			continue // already on its way
		}
		if m.jobs[job.ID] == nil {
			m.order = append(m.order, job.ID)
		}
		m.jobs[job.ID] = job
		m.mu.Unlock()
		m.queue <- job
		ids = append(ids, job.ID)
	}
	return ids, nil
}

// Cancel stops a queued or running job. Received parts stay on the receiver,
// so Retry continues where the transfer stopped.
func (m *Manager) Cancel(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	job := m.jobs[id]
	if job == nil {
		return false
	}
	switch job.Status {
	case StatusQueued:
		job.Status = StatusCancelled
		job.Finished = m.now()
	case StatusRunning:
		if job.cancel != nil {
			job.cancel()
		}
	default:
		return false
	}
	return true
}

// Retry re-queues a failed or cancelled job; it resumes from what the receiver has.
func (m *Manager) Retry(id string) error {
	m.mu.Lock()
	job := m.jobs[id]
	if job == nil {
		m.mu.Unlock()
		return errors.New("nie ma takiego transferu")
	}
	source, address, dest := job.Source, job.Address, job.Dest
	m.mu.Unlock()
	_, err := m.Enqueue([]string{source}, address, dest)
	return err
}

// ClearFinished removes finished jobs from the list (history keeps them).
func (m *Manager) ClearFinished() {
	m.mu.Lock()
	defer m.mu.Unlock()
	kept := m.order[:0]
	for _, id := range m.order {
		if job := m.jobs[id]; job.Status == StatusQueued || job.Status == StatusRunning {
			kept = append(kept, id)
		} else {
			delete(m.jobs, id)
		}
	}
	m.order = kept
}

// Snapshot returns all jobs, oldest first.
func (m *Manager) Snapshot() []JobView {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := m.now()
	views := make([]JobView, 0, len(m.order))
	for _, id := range m.order {
		job := m.jobs[id]
		v := JobView{
			ID: job.ID, Name: job.Name, Source: job.Source, Peer: job.PeerName, Address: job.Address, Dest: job.Dest,
			Status: job.Status, Error: job.Error, TotalBytes: job.TotalBytes, TotalFiles: job.TotalFiles,
			DoneBytes: job.progress.DoneBytes.Load(), DoneFiles: job.progress.DoneFiles.Load(),
			Resumed: job.progress.Resumed.Load(), Active: job.progress.Active(),
		}
		if !job.Started.IsZero() {
			end := now
			if !job.Finished.IsZero() {
				end = job.Finished
			}
			v.Elapsed = end.Sub(job.Started).Seconds()
		}
		if job.Status == StatusRunning {
			v.Speed = job.meter.sample(now, v.DoneBytes-v.Resumed)
			if v.Speed > 0 {
				v.ETA = float64(v.TotalBytes-v.DoneBytes) / v.Speed
			}
		} else if v.Elapsed > 0 {
			v.Speed = float64(v.DoneBytes-v.Resumed) / v.Elapsed
		}
		views = append(views, v)
	}
	return views
}

func (m *Manager) worker() {
	for job := range m.queue {
		m.run(job)
	}
}

func (m *Manager) run(job *Job) {
	m.mu.Lock()
	if job.Status != StatusQueued || m.jobs[job.ID] != job {
		m.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	job.cancel = cancel
	job.Status = StatusRunning
	job.Started = m.now()
	session := m.sessions[job.Address]
	parallel := m.Parallel
	m.mu.Unlock()

	var err error
	if session == nil {
		err = ErrNotPaired
	} else {
		err = session.Send(ctx, job.plan, job.Dest, parallel, job.progress)
	}

	m.mu.Lock()
	job.Finished = m.now()
	job.cancel = nil
	switch {
	case err == nil:
		job.Status = StatusDone
	case errors.Is(err, context.Canceled):
		job.Status = StatusCancelled
		job.Error = "anulowano – wyślij ponownie, aby wznowić"
	default:
		job.Status = StatusFailed
		job.Error = err.Error()
		if errors.Is(err, ErrNotPaired) {
			delete(m.sessions, job.Address)
		}
	}
	entry := HistoryEntry{
		Direction: "sent", Name: job.Name, Peer: job.PeerName, Bytes: job.progress.DoneBytes.Load(),
		Files: int(job.progress.DoneFiles.Load()), Status: job.Status, Error: job.Error,
		Started: job.Started, Finished: job.Finished, Location: job.Source, JobID: job.ID,
	}
	m.mu.Unlock()
	m.History.Add(entry)
}
