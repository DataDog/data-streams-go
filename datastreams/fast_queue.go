package datastreams

import "sync/atomic"

const (
	queueSize = 10000
)

// there are many writers, there is only 1 reader.
// each value will be read at most once.
// reader will stop if it catches up with writer
// if reader is too slow, there is no guarantee in which order values will be dropped.
type fastQueue struct {
	elements [queueSize]atomic.Pointer[statsPoint]
	writePos int64
	readPos  int64
}

func newFastQueue() *fastQueue {
	return &fastQueue{}
}

func (q *fastQueue) push(p *statsPoint) {
	ind := atomic.AddInt64(&q.writePos, 1)
	p.queuePos = ind - 1
	q.elements[(ind-1)%queueSize].Store(p)
}

func (q *fastQueue) pop() *statsPoint {
	writePos := atomic.LoadInt64(&q.writePos)
	if writePos <= q.readPos {
		return nil
	}
	loaded := q.elements[q.readPos%queueSize].Load()
	if loaded == nil || loaded.queuePos < q.readPos {
		// the write started, but hasn't finished yet, the element we read
		// is the one from the previous cycle.
		return nil
	}
	q.readPos++
	return loaded
}
