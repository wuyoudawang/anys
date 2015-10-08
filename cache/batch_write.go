package cache

import (
	"encoding/binary"
	"fmt"

	"anys/pkg/utils"
)

const (
	header = 12
)

type Inserter interface {
	Put(k, v []byte)
	Delete(k []byte)
}

type Batch struct {
	rep    []byte
	scarch []byte
}

func (b *Batch) count() uint32 {
	return binary.LittleEndian.Uint32(b.rep[8:])
}

func (b *Batch) increCount() {
	binary.LittleEndian.PutUint32(b.rep[8:], b.count()+1)
}

func (b *Batch) sequence() uint64 {
	return binary.LittleEndian.Uint64(b.rep[:8])
}

func (b *Batch) setSequence(v uint64) {
	binary.LittleEndian.PutUint64(b.rep[:8], v)
}

func (b *Batch) Put(key, value []byte) {
	b.increCount()
	b.rep = append(b.rep, byte(TypeValue))
	l := utils.PutLenPrefixedBytes(&(b.scarch), &key)
	b.rep = append(b.rep, b.scarch[:l]...)
	l = utils.PutLenPrefixedBytes(&b.scarch, &value)
	b.rep = append(b.rep, b.scarch[:l]...)
}

func (b *Batch) Delete(key []byte) {
	b.increCount()
	b.rep = append(b.rep, byte(TypeDeletion))
	l := utils.PutLenPrefixedBytes(&b.scarch, &key)
	b.rep = append(b.rep, b.scarch[:l]...)
}

func (b *Batch) Clear() {
	b.rep = b.rep[:header]
	for i := 0; i < header; i++ {
		b.rep[i] = 0
	}
}

func (b *Batch) Iterate(iser Inserter) error {
	if len(b.rep) < header {
		return fmt.Errorf("malformed WriteBatch (too small)")
	}

	input := b.rep[header:]
	found := 0
	var key, value []byte
	for len(input) > 0 {
		found++
		tag := input[0]
		input = input[1:]

		switch tag {
		case TypeValue:
			if utils.GetLenPrefixedBytes(&key, &input) &&
				utils.GetLenPrefixedBytes(&value, &input) {
				iser.Put(key, value)
			} else {
				return fmt.Errorf("bad WriteBatch Put")
			}
		case TypeDeletion:
			if utils.GetLenPrefixedBytes(&key, &input) {
				iser.Delete(key)
			} else {
				return fmt.Errorf("bad WriteBatch Delete")
			}
		default:
			return fmt.Errorf("unknown WriteBatch tag")
		}
	}

	if uint32(found) != b.count() {
		return fmt.Errorf("WriteBatch has wrong count")
	} else {
		return nil
	}
}

func (b *Batch) InsertInto(memt *MemTable) {
	inser := new(batchInserter)
	inser.sequence_ = b.sequence()
	inser.mem = memt
	b.Iterate(inser)
}

type batchInserter struct {
	sequence_ uint64
	mem       *MemTable
}

func (bi *batchInserter) Put(key, val []byte) {
	bi.mem.Add(bi.sequence_, TypeValue, key, val)
	bi.sequence_++
}

func (bi *batchInserter) Delete(key []byte) {
	bi.mem.Add(bi.sequence_, TypeDeletion, key, nil)
	bi.sequence_++
}
