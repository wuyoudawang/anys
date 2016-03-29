package utils

import (
	"math/rand"
	"sync/atomic"
	"unsafe"

	"github.com/liuzhiyi/anys/pkg/comparator"
)

const (
	KMAX_HEIGHT = 12
)

type skipNode struct {
	key   []byte
	next_ []unsafe.Pointer
}

func (sn *skipNode) next(n int) *skipNode {
	if n >= 0 {
		return (*skipNode)(atomic.LoadPointer(&(sn.next_[n])))
	}
	return nil
}

func (sn *skipNode) setNext(n int, x *skipNode) {
	if n >= 0 {
		val := unsafe.Pointer(x)
		atomic.StorePointer(&(sn.next_[n]), val)
	}
}

type SkipIterator struct {
	node *skipNode
	list *Skiplist
}

func (si *SkipIterator) Key() []byte {
	if si.Valid() {
		return si.node.key
	}
	return []byte{}
}

func (si *SkipIterator) Valid() bool {
	return si.node != nil
}

func (si *SkipIterator) Next() {
	if si.Valid() {
		si.node = si.node.next(0)
	}
}

func (si *SkipIterator) Prev() {
	if si.Valid() {
		si.node = si.list.FindLT(si.node.key)
	}
}

func (si *SkipIterator) Seek(key []byte) {
	si.node = si.list.FindGTEQ(key, nil)
}

func (si *SkipIterator) First() {
	si.node = si.list.head.next(0)
}

func (si *SkipIterator) Last() {
	node := si.list.Last()
	if node == si.list.head {
		node = nil
	}
	si.node = node
}

type Skiplist struct {
	maxHeight int32
	head      *skipNode
	rnd       *rand.Rand
	compare   comparator.Comparator
}

func (sl *Skiplist) getMaxHeight() int {
	return int(atomic.LoadInt32(&sl.maxHeight))
}

func (sl *Skiplist) equal(a, b []byte) bool {
	return sl.compare.Compare(a, b) == 0
}

func (sl *Skiplist) newNode(key []byte, height int) *skipNode {
	node := new(skipNode)
	node.key = append(node.key, key...)
	node.next_ = make([]unsafe.Pointer, height)
	return node
}

func (sl *Skiplist) Iterator() *SkipIterator {
	i := new(SkipIterator)
	i.list = sl
	i.node = nil
	return i
}

func (sl *Skiplist) randomHeight() int {
	branching := 4
	height := 1

	for height < KMAX_HEIGHT && (sl.rnd.Int()%branching) == 0 {
		height++
	}
	return height
}

func (sl *Skiplist) KeyIsAfterNode(key []byte, n *skipNode) bool {
	return (n != nil) && (sl.compare.Compare(n.key, key) < 0)
}

func (sl *Skiplist) FindGTEQ(key []byte, prev []*skipNode) *skipNode {
	x := sl.head
	level := sl.getMaxHeight() - 1
	for {
		next := x.next(level)
		if sl.KeyIsAfterNode(key, next) {
			x = next
		} else {
			if prev != nil {
				prev[level] = x
			}

			if level == 0 {
				return next
			} else {
				level--
			}
		}
	}
}

func (sl *Skiplist) FindLT(key []byte) *skipNode {
	x := sl.head
	level := sl.getMaxHeight() - 1
	for {
		next := x.next(level)
		if next == nil || sl.compare.Compare(next.key, key) >= 0 {
			if level == 0 {
				return x
			} else {
				level--
			}
		} else {
			x = next
		}
	}
}

func (sl *Skiplist) Last() *skipNode {
	x := sl.head
	level := sl.getMaxHeight() - 1
	for {
		next := x.next(level)
		if next == nil {
			if level == 0 {
				return x
			} else {
				level--
			}
		} else {
			x = next
		}
	}
}

func (sl *Skiplist) Insert(key []byte) {
	prev := make([]*skipNode, KMAX_HEIGHT)
	x := sl.FindGTEQ(key, prev)
	if x == nil || sl.equal(x.key, key) {
		return
	}

	height := sl.randomHeight()
	if height > sl.getMaxHeight() {
		for i := sl.getMaxHeight(); i < height; i++ {
			prev[i] = sl.head
		}
		atomic.StoreInt32(&sl.maxHeight, int32(height))
	}

	x = sl.newNode(key, height)
	for i := 0; i < height; i++ {
		x.setNext(i, prev[i].next(i))
		prev[i].setNext(i, x)
	}
}

func (sl *Skiplist) Contains(key []byte) bool {
	x := sl.FindGTEQ(key, nil)
	if x != nil && sl.equal(x.key, key) {
		return true
	} else {
		return false
	}
}
