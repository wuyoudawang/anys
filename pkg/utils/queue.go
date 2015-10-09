package utils

import (
	"unsafe"
)

type Queue struct {
	elems []unsafe.Pointer
	size  int64
}

func NewQueue(n int64) *Queue {
	q := new(Queue)
	q.elems = make([]unsafe.Pointer, n)
	q.size = 0

	return q
}

func (q *Queue) Tail() unsafe.Pointer {
	if q.Empty() {
		return nil
	}

	return q.elems[q.size]
}

func (q *Queue) Index(i int) unsafe.Pointer {
	if q.Empty() {
		return nil
	}

	return q.elems[i]
}

func (q *Queue) Size() int64 {
	return q.size
}

func (q *Queue) Head() unsafe.Pointer {
	if q.Empty() {
		return nil
	}

	return q.elems[0]
}

func (q *Queue) Push(elem unsafe.Pointer) {
	if int64(len(q.elems)) == q.size {
		q.elems = append(q.elems, elem)
	} else {
		q.elems[q.size] = elem
	}

	q.size++
}

func (q *Queue) Pop() unsafe.Pointer {
	if q.Empty() {
		return nil
	}

	elem := q.elems[q.size-1]
	q.size--

	return elem
}

func (q *Queue) Empty() bool {
	return q.size == 0
}
