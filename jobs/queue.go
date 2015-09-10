package jobs

import (
	"sync"
)

const (
	minQueueSize     = 4
	defaultQueueSize = 10
)

type CycleQueue struct {
	buf     []*Job
	r       int
	w       int
	lastJob *Job
	m       sync.Mutex
	locked  bool
	mLock   sync.Mutex
}

func NewCycleQueue() *CycleQueue {
	return NewCycleQueueSize(defaultQueueSize)
}

func NewCycleQueueSize(size int) *CycleQueue {
	if size < minQueueSize {
		size = minQueueSize
	}

	q := new(CycleQueue)
	buf := make([]*Job, 1<<uint(size))
	q.reset(buf)

	return q
}

func (q *CycleQueue) Reset(buf []*Job) {
	q.reset(buf)
}

func (q *CycleQueue) reset(buf []*Job) {
	q.buf = buf
	q.r = 0
	q.w = 0
	q.lastJob = nil
}

func (q *CycleQueue) Full() bool {
	return q.w == (q.r ^ len(q.buf))
}

func (q *CycleQueue) Empty() bool {
	return q.w == q.r
}

func (q *CycleQueue) Incr(val int) int {
	return (val + 1) & (2*len(q.buf) - 1)
}

func (q *CycleQueue) Size() int {
	return 0
}

func (q *CycleQueue) Push(job *Job) bool {

	if q.Full() {
		return false
	}

	if q.isLocked() {
		q.setLocked(false)
		defer q.unlock()
	}

	q.buf[q.index(q.w)] = job
	q.w = q.Incr(q.w)
	return true
}

func (q *CycleQueue) index(val int) int {
	return val & (len(q.buf) - 1)
}

func (q *CycleQueue) lock() {
	q.m.Lock()
}

func (q *CycleQueue) unlock() {
	q.m.Unlock()
}

func (q *CycleQueue) setLocked(val bool) {
	q.mLock.Lock()
	defer q.mLock.Unlock()

	q.locked = val
}

func (q *CycleQueue) isLocked() bool {
	q.mLock.Lock()
	defer q.mLock.Unlock()

	return q.locked
}

func (q *CycleQueue) Pop() *Job {
	q.lock()
	defer q.unlock()

	if q.Empty() {

		q.setLocked(true)
		q.lock()
	}

	job := q.buf[q.index(q.r)]
	q.r = q.Incr(q.r)
	return job
}

type Queue struct {
	elems []int
	size  int
}

func NewQueue(n int) *Queue {
	q := new(Queue)
	q.elems = make([]int, n)
	q.size = 0

	return q
}

func (q *Queue) Push(elem int) {
	if len(q.elems) == q.size {
		q.elems = append(q.elems, elem)
	} else {
		q.elems[q.size] = elem
	}

	q.size++
}

func (q *Queue) Pop() int {
	if q.Empty() {
		return -1
	}

	elem := q.elems[q.size-1]
	q.size--

	return elem
}

func (q *Queue) Empty() bool {
	return q.size == 0
}
