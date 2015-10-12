package cache

import (
	"encoding/binary"
	"fmt"
)

const (
	kMaxSeq uint64 = (uint64(1) << 56) - 1
	// Maximum value possible for packed sequence number and type.
	kMaxNum uint64 = (kMaxSeq << 8) | uint64(kValueTypeForSeek)
)

type internalKey []byte

func newInternalKey(key []byte, seq uint64, kt int) internalKey {
	if seq > kMaxSeq {
		panic("invalid sequence number")
	} else if kt > kTypeValue {
		panic("invalid type")
	}

	ik := make(internalKey, len(key)+8)
	copy(ik, key)
	binary.LittleEndian.PutUint64(ik, (seq<<8)|uint64(kt))
	return ik
}

func (ik internalKey) assert() {
	if ik == nil {
		panic("nil iKey")
	}
	if len(ik) < 8 {
		panic(fmt.Sprintf("iKey %q, len=%d: invalid length", []byte(ik), len(ik)))
	}
}

func (ik internalKey) userKey() []byte {
	ik.assert()
	return ik[:len(ik)-8]
}

func (ik internalKey) encode() []byte {
	ik.assert()
	return ik
}
