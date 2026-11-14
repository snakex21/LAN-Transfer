package transfer

import (
	"errors"
	"hash"
	"io"
	"sync"
	"sync/atomic"
)

// SHA-256 (~2 GB/s on one core), TLS and disk I/O used to run one after
// another in a single goroutine on each side, which capped a transfer at
// roughly 40% of what the disk can do. The helpers below overlap them:
// a chunk is hashed while it is being written (receiver) or encrypted and sent
// (sender). Order is preserved because each consumer is a single goroutine.

const (
	chunkSize     = 1 << 20
	pipelineDepth = 8
)

type chunk struct {
	buf  []byte
	n    int
	refs atomic.Int32
}

// chunkPool bounds how many chunks one transfer has in flight; the buffers
// themselves are recycled through a process-wide sync.Pool so thousands of
// small files do not allocate megabytes each.
type chunkPool chan *chunk

var chunkBuffers = sync.Pool{New: func() any { return &chunk{buf: make([]byte, chunkSize)} }}

func newChunkPool(depth int) chunkPool {
	pool := make(chunkPool, depth)
	for i := 0; i < depth; i++ {
		pool <- nil // allocated lazily by get
	}
	return pool
}

func (p chunkPool) get() *chunk {
	c := <-p
	if c == nil {
		c = chunkBuffers.Get().(*chunk)
	}
	return c
}

func (p chunkPool) release(c *chunk) {
	if c.refs.Add(-1) == 0 {
		p <- c
	}
}

// drain returns every buffer to the shared pool once the pipeline is idle.
func (p chunkPool) drain() {
	for {
		select {
		case c := <-p:
			if c != nil {
				chunkBuffers.Put(c)
			}
		default:
			return
		}
	}
}

// copyHashing copies src to dst and feeds h, writing and hashing concurrently.
// progress is called after each chunk is handed over. It returns the number of
// bytes read from src; a write error stops the copy.
func copyHashing(dst io.Writer, h hash.Hash, src io.Reader, progress func(int64)) (int64, error) {
	pool := newChunkPool(pipelineDepth)
	writeCh := make(chan *chunk, pipelineDepth)
	hashCh := make(chan *chunk, pipelineDepth)
	var writeErr atomic.Pointer[error]
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for c := range writeCh {
			if writeErr.Load() == nil {
				if _, err := dst.Write(c.buf[:c.n]); err != nil {
					writeErr.Store(&err)
				}
			}
			pool.release(c)
		}
	}()
	go func() {
		defer wg.Done()
		for c := range hashCh {
			h.Write(c.buf[:c.n])
			pool.release(c)
		}
	}()

	var total int64
	var readErr error
	for {
		if p := writeErr.Load(); p != nil {
			break
		}
		c := pool.get()
		n, err := io.ReadFull(src, c.buf)
		if n > 0 {
			c.n = n
			c.refs.Store(2)
			writeCh <- c
			hashCh <- c
			total += int64(n)
			if progress != nil {
				progress(int64(n))
			}
		} else {
			pool <- c
		}
		if err != nil {
			if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
				readErr = err
			}
			break
		}
	}
	close(writeCh)
	close(hashCh)
	wg.Wait()
	pool.drain()
	if p := writeErr.Load(); p != nil {
		return total, *p
	}
	return total, readErr
}

// prefetchReader reads a file ahead in a goroutine and hashes it in another,
// so the HTTP transport only has to encrypt and send.
type prefetchReader struct {
	out     chan *chunk
	pool    chunkPool
	cur     *chunk
	off     int
	err     error
	hashed  chan struct{}
	stop    chan struct{}
	stopped sync.Once
}

func newPrefetchReader(src io.Reader, h hash.Hash, onRead func(int64)) *prefetchReader {
	r := &prefetchReader{
		out: make(chan *chunk, pipelineDepth), pool: newChunkPool(pipelineDepth),
		hashed: make(chan struct{}), stop: make(chan struct{}),
	}
	hashCh := make(chan *chunk, pipelineDepth)
	go func() {
		defer close(r.hashed)
		for c := range hashCh {
			h.Write(c.buf[:c.n])
			r.pool.release(c)
		}
	}()
	go func() {
		defer close(hashCh)
		defer close(r.out)
		for {
			var c *chunk
			select {
			case c = <-r.pool:
			case <-r.stop:
				return
			}
			if c == nil {
				c = chunkBuffers.Get().(*chunk)
			}
			n, err := io.ReadFull(src, c.buf)
			if n > 0 {
				c.n = n
				c.refs.Store(2)
				hashCh <- c
				select {
				case r.out <- c:
				case <-r.stop:
					return
				}
				if onRead != nil {
					onRead(int64(n))
				}
			}
			if err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
					select {
					case r.out <- &chunk{n: -1, buf: []byte(err.Error())}:
					case <-r.stop:
					}
				}
				return
			}
		}
	}()
	return r
}

func (r *prefetchReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	if r.cur == nil {
		c, ok := <-r.out
		if !ok {
			<-r.hashed // the digest must be complete before EOF (HTTP trailer)
			r.pool.drain()
			r.err = io.EOF
			return 0, io.EOF
		}
		if c.n < 0 {
			r.err = errors.New(string(c.buf))
			return 0, r.err
		}
		r.cur, r.off = c, 0
	}
	n := copy(p, r.cur.buf[r.off:r.cur.n])
	r.off += n
	if r.off == r.cur.n {
		r.pool.release(r.cur)
		r.cur = nil
	}
	return n, nil
}

// Close stops the read-ahead goroutine (e.g. after a cancelled upload).
func (r *prefetchReader) Close() {
	r.stopped.Do(func() { close(r.stop) })
}
