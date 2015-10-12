package table

import (
	"bytes"

	"anys/cache/iterator"
	"anys/cache/option"
)

type twoLevelIterator struct {
	indexIter       iterator.Interface
	blockIter       iterator.Interface
	err             error
	readOpt         *option.Options
	dataBlockHandle []byte
}

func NewTwoLevelIterator(indexIter iterator.Interface, opt *option.ReadOptions) iterator.Interface {
	return &twoLevelIterator{
		indexIter: indexIter,
		readOpt:   opt,
	}
}

func (tli *twoLevelIterator) Seek(target []byte) {
	tli.indexIter.Seek(target)

}

func (tli *twoLevelIterator) saveError(err error) {
	if tli.err == nil && err != nil {
		tli.err = err
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

		} else {

		}
	}
}
