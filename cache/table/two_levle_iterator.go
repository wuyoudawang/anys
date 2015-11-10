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

}

func (tli *twoLevelIterator) Last() {

}

func (tli *twoLevelIterator) Seek(target []byte) {
	tli.indexIter.Seek(target)

}

func (tli *twoLevelIterator) Next() {

}

func (tli *twoLevelIterator) Prev() {

}

func (tli *twoLevelIterator) Key() []byte {
	return nil
}

func (tli *twoLevelIterator) Value() []byte {
	return nil
}

func (tli *twoLevelIterator) saveError(err error) {
	if tli.err == nil && err != nil {
		tli.err = err
	}
}

func (tli *twoLevelIterator) Valid() bool {
	return false
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
