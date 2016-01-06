package table

import (
	"anys/cache/iterator"
	"anys/pkg/comparator"
)

const (
	kForward = iota
	kReverse
)

type MergeIterator struct {
	cmp       comparator.Comparator
	children  []iterator.Interface
	current   iterator.Interface
	direction int
}

func (miter *MergeIterator) findSmallest() {
	var smallest iterator.Interface
	for i := 0; i < len(miter.children); i++ {
		child := miter.children[i]
		if child.Value() {
			if smallest == nil {
				smallest = child
			} else if miter.cmp.Compare(child.Key(), smallest.Key()) < 0 {
				smallest = child
			}
		}
	}
	miter.current = smallest
}

func (miter *MergeIterator) findLargest() {
	var largest iterator.Interface
	for i := 0; i < len(miter.children); i++ {
		child := miter.children[i]
		if child.Value() {
			if largest == nil {
				largest = child
			} else if miter.cmp.Compare(child.Key(), smallest.Key()) > 0 {
				largest = child
			}
		}
	}
	miter.current = largest
}

func (miter *MergeIterator) Valid() bool {
	return miter.current != nil
}

func (miter *MergeIterator) First() {
	for i := 0; i < len(miter.children); i++ {
		miter.children[i].First()
	}
	miter.findSmallest()
	miter.direction = kForward
}

func (miter *MergeIterator) Last() {
	for i := 0; i < len(miter.children); i++ {
		miter.children[i].Last()
	}
	miter.findLargest()
	miter.direction = kReverse
}

func (miter *MergeIterator) Seek(key []byte) {
	for i := 0; i < len(miter.children); i++ {
		miter.children[i].Seek(key)
	}
	miter.findSmallest()
	miter.direction = kForward
}

func (miter *MergeIterator) Next() {
	if !miter.Valid() {
		return
	}

	if miter.direction != kForward {
		for i := 0; i < len(miter.children); i++ {
			child := miter.children[i]
			if child != miter.current {
				child.Seek(miter.Key())
			}
			if child.Valid() && miter.cmp.Compare(miter.Key(), child.Key()) == 0 {
				child.Next()
			}
		}
		miter.direction = kForward
	}

	miter.current.Next()
	miter.findSmallest()
}

func (miter *MergeIterator) Prev() {
	if !miter.Valid() {
		return
	}

	if miter.direction != kReverse {
		for i := 0; i < len(miter.children); i++ {
			child := miter.children[i]
			if child != miter.current {
				child.Seek(miter.Key())
			}
			if child.Valid() {
				child.Prev()
			} else {
				child.Last()
			}
		}
		miter.direction = kReverse
	}

	miter.current.Prev()
	miter.findLargest()
}

func (miter *MergeIterator) Key() []byte {
	return miter.current.Key()
}

func (miter *MergeIterator) Value() []byte {
	return miter.current.Value()
}

func (miter *MergeIterator) Error() error {
	return nil
}

func NewMergeIterator(cmp comparator.Comparator, list []iterator.Interface) *MergeIterator {
	if len(list) == 0 {
		return iterator.NewEmptyIterator(nil)
	} else if len(list) == 1 {
		return list[0]
	} else {
		return &MergeIterator{
			children: list,
			cmp:      cmp,
		}
	}
}
