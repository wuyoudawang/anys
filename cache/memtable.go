package cache

import (
	"encoding/binary"
	"fmt"

	"github.com/liuzhiyi/anys/cache/iterator"
	"github.com/liuzhiyi/anys/pkg/comparator"
	"github.com/liuzhiyi/anys/pkg/utils"
)

func getLenPrefixedBytes(data []byte) []byte {
	l, offset := binary.Varint(data)
	return data[offset : int64(offset)+l]
}

func encodeKey(src []byte) (dst []byte) {
	dst = make([]byte, len(src)+5)
	s := binary.PutVarint(dst, int64(len(src)))
	copy(dst[s:], src)
	return
}

type MemTableIterator struct {
	*utils.SkipIterator
}

func (mti *MemTableIterator) Seek(k []byte) {
	mti.SkipIterator.Seek(encodeKey(k))
}

func (mti *MemTableIterator) Key() []byte {
	return getLenPrefixedBytes(mti.SkipIterator.Key())
}

func (mti *MemTableIterator) Value() []byte {
	raw := mti.SkipIterator.Key()
	key := getLenPrefixedBytes(raw)
	return getLenPrefixedBytes(raw[len(key):])
}

type MemTable struct {
	table       utils.Skiplist
	refs        int
	compare     comparator.Comparator
	varbuf      [10]byte
	memoryUsege int
}

func (mt *MemTable) Ref() {
	mt.refs++
}

func (mt *MemTable) Unref() error {
	mt.refs--
	if mt.refs < 0 {
		return fmt.Errorf("reference ")
	} else if mt.refs == 0 {

	}
	return nil
}

func (mt *MemTable) Iterator() iterator.Interface {
	return &MemTableIterator{mt.table.Iterator()}
}

func (mt *MemTable) Add(s uint64, valueType int, key, value []byte) {
	// Format of an entry is concatenation of:
	//  key_size     : varint32 of internal_key.size()
	//  key bytes    : char[internal_key.size()]
	//  value_size   : varint32 of value.size()
	//  value bytes  : char[value.size()]
	keyLen := len(key)
	valuelen := len(value)
	internalKeyLen := keyLen + 8
	var buf []byte

	writtenLen := binary.PutVarint(mt.varbuf[:], int64(internalKeyLen))
	buf = append(buf, mt.varbuf[:writtenLen]...)
	buf = append(buf, key...)
	binary.LittleEndian.PutUint64(mt.varbuf[:], (s<<8)|uint64(valueType))
	buf = append(buf, mt.varbuf[:8]...)
	writtenLen = binary.PutVarint(mt.varbuf[:], int64(valuelen))
	buf = append(buf, mt.varbuf[:writtenLen]...)
	buf = append(buf, value...)
	mt.memoryUsege += len(buf)

	mt.table.Insert(buf)
}

func (mt *MemTable) Get(key []byte) ([]byte, bool) {
	itor := mt.Iterator()
	itor.Seek(key)
	if itor.Valid() {
		// entry format is:
		//    klength  varint32
		//    userkey  char[klength]
		//    tag      uint64
		//    vlength  varint32
		//    value    char[vlength]
		// Check that it belongs to same user key.  We do not check the
		// sequence number since the Seek() call above should have skipped
		// all entries with overly large sequence numbers.
		entry := itor.Key()
		keyLen, offset := binary.Varint(entry)
		valueType := binary.LittleEndian.Uint64(entry[int64(offset)+keyLen-8:])
		switch valueType {
		case kTypeValue:
			v := getLenPrefixedBytes(entry[int64(offset)+keyLen:])
			return v, true
		case kTypeDeletion:
			return nil, false
		}
	}
	return nil, false
}

func (mt *MemTable) ApproximateMemoryUsage() int {
	return mt.memoryUsege
}
