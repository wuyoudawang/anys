package table

import (
	"bytes"

	"anys/cache/iterator"
	"anys/cache/option"
)

type BlockFunction func(interface{}, *option.ReadOptions, []byte) iterator.Interface

type twoLevelIterator struct {
	block_function  BlockFunction
	args            interface{}
	indexIter       iterator.Interface
	blockIter       iterator.Interface
	err             error
	readOpt         *option.ReadOptions
	dataBlockHandle []byte
}

func NewTwoLevelIterator(indexIter iterator.Interface, opt *option.ReadOptions) iterator.Interface {
	return &twoLevelIterator{
		indexIter: indexIter,
		readOpt:   opt,
	}
}

func (tli *twoLevelIterator) Error() error {
	return tli.err
}

func (tli *twoLevelIterator) First() {
	tli.indexIter.First()
	tli.InitDataBlock()
	if tli.blockIter != nil {
		tli.blockIter.First()
	}
	tli.skipEmptyDataBlocksForward()
}

func (tli *twoLevelIterator) Last() {
	tli.indexIter.Last()
	tli.InitDataBlock()
	if tli.blockIter != nil {
		tli.blockIter.Last()
	}
	tli.skipEmptyDataBlocksBackward()
}

func (tli *twoLevelIterator) Seek(target []byte) {
	tli.indexIter.Seek(target)
	tli.InitDataBlock()
	if tli.blockIter != nil {
		tli.blockIter.Seek(target)
	}
	tli.skipEmptyDataBlocksBackward()
}

func (tli *twoLevelIterator) Next() {
	// assert(Valid());
	tli.blockIter.Next()
	tli.skipEmptyDataBlocksForward()
}

func (tli *twoLevelIterator) Prev() {
	// assert(Valid());
	tli.blockIter.Prev()
	tli.skipEmptyDataBlocksBackward()
}

func (tli *twoLevelIterator) Key() []byte {
	if !tli.Valid() {
		return []byte{}
	}
	return tli.blockIter.Key()
}

func (tli *twoLevelIterator) Value() []byte {
	if !tli.Valid() {
		return []byte{}
	}
	return tli.blockIter.Value()
}

func (tli *twoLevelIterator) saveError(err error) {
	if tli.err == nil && err != nil {
		tli.err = err
	}
}

func (tli *twoLevelIterator) Valid() bool {
	return tli.blockIter.Valid()
}

func (tli *twoLevelIterator) skipEmptyDataBlocksForward() {
	for tli.blockIter == nil || !tli.blockIter.Valid() {
		if !tli.indexIter.Valid() {
			tli.setBlockIter(nil)
			return
		}
		tli.indexIter.Next()
		tli.InitDataBlock()
		if tli.blockIter != nil {
			tli.blockIter.First()
		}
	}
}

func (tli *twoLevelIterator) skipEmptyDataBlocksBackward() {
	for tli.blockIter == nil || !tli.blockIter.Valid() {
		if tli.indexIter.Valid() {
			tli.setBlockIter(nil)
			return
		}
		tli.indexIter.Prev()
		tli.InitDataBlock()
		if tli.blockIter != nil {
			tli.blockIter.Last()
		}
	}
}

func (tli *twoLevelIterator) setBlockIter(iter iterator.Interface) {
	if tli.blockIter != nil {
		tli.saveError(tli.blockIter.Error())
	}
	tli.blockIter = iter
}

func (tli *twoLevelIterator) InitDataBlock() {
	if !tli.indexIter.Valid() {
		tli.setBlockIter(nil)
	} else {
		handle := tli.indexIter.Value()
		if tli.blockIter != nil && bytes.Compare(tli.dataBlockHandle, handle) == 0 {
			// blockIter is already constructed with this iterator, so
			// no need to change anything
		} else {
			iter := tli.block_function(tli.args, tli.readOpt, handle)
			tli.dataBlockHandle = handle
			tli.setBlockIter(iter)
		}
	}
}
