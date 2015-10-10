package cache

import (
	"encoding/binary"

	"anys/pkg/utils"
)

const (
	kComparator     = 1
	kLogNumber      = 2
	kNextFileNumber = 3
	kLastSequence   = 4
	kCompactPointer = 5
	kDeletedFile    = 6
	kNewFile        = 7
	// 8 was used for large value refs
	kPrevLogNumber = 9
)

type fileMetaData struct {
	refs         int
	allowedSeeks int
	number       uint64
	fileSize     uint64
	smallest     internalKey
	largest      internalKey
	level        int
}

type dFile struct {
	level  int
	number uint64
}

type compactPtr struct {
	level int
	key   internalKey
}

type versionEdit struct {
	comparatorName    string
	logNumber         uint64
	prevLogNumber     uint64
	nextFileNumber    uint64
	lastSequence      uint64
	hasComparator     bool
	hasLogNumber      bool
	hasPrevLogNumber  bool
	hasNextFileNumber bool
	hasLastSequence   bool

	compactPointers []compactPtr
	deletFiles      []dFile
	newFiles        []fileMetaData
}

func (ve *versionEdit) SetLogNumber(num uint64) {
	ve.hasLogNumber = true
	ve.logNumber = num
}

func (ve *versionEdit) SetPrevLogNumber(num uint64) {
	ve.hasPrevLogNumber = true
	ve.prevLogNumber = num
}

func (ve *versionEdit) SetNextFileNumber(num uint64) {
	ve.hasNextFileNumber = true
	ve.nextFileNumber = num
}

func (ve *versionEdit) SetSequence(seq uint64) {
	ve.hasLastSequence = true
	ve.lastSequence = seq
}

func (ve *versionEdit) SetComparatorName(name string) {
	ve.hasComparator = true
	ve.comparatorName = name
}

func (ve *versionEdit) SetCompactPointer(ptr unsafe.Pointer) {
	ve.compactPointers = append(ve.compactPointers, ptr)
}

func (ve *versionEdit) Clear() {
	ve.comparatorName = ""
	ve.logNumber = 0
	ve.prevLogNumber = 0
	ve.nextFileNumber = 0
	ve.lastSequence = 0
	ve.hasComparator = false
	ve.hasLogNumber = false
	ve.hasNextFileNumber = false
	ve.hasLastSequence = false
	ve.deletFiles = ve.deletFiles[:0]
	ve.newFiles = ve.newFiles[:0]
}

func (ve *versionEdit) encodeTo(dst *[]byte) {
	if ve.hasComparator {
		binary.PutVarint(dst, kComparator)
		utils.PutLenPrefixedBytes(&dst, &ve.comparatorName)
	}

	if ve.hasLogNumber {
		binary.PutVarint(dst, kLogNumber)
		binary.PutVarint(dst, ve.logNumber)
	}

	if ve.hasPrevLogNumber {
		binary.PutVarint(dst, kPrevLogNumber)
		binary.PutVarint(dst, ve.prevLogNumber)
	}

	if ve.hasNextFileNumber {
		binary.PutVarint(dst, kNextFileNumber)
		binary.PutVarint(dst, ve.nextFileNumber)
	}

	if ve.hasLastSequence {
		binary.PutVarint(dst, kLastSequence)
		binary.PutVarint(dst, ve.lastSequence)
	}

	for _, cp := range ve.compactPointers {
		binary.PutVarint(dst, kCompactPointer)
		binary.PutVarint(dst, cp.level)
		utils.PutLenPrefixedBytes(&dst, cp.key.encode())
	}

	for _, ditem := range ve.deletFiles {
		binary.PutVarint(dst, kDeletedFile)
		binary.PutVarint(dst, ditem.level)
		binary.PutVarint(dst, ditem.number)
	}

	for _, metaData := range ve.newFiles {
		binary.PutVarint(dst, kNewFile)
		binary.PutVarint(dst, metaData.level) // level
		binary.PutVarint(dst, metaData.number)
		binary.PutVarint(dst, metaData.fileSize)
		utils.PutLenPrefixedBytes(dst, f.smallest.Encode())
		utils.PutLenPrefixedBytes(dst, f.largest.Encode())
	}
}

func (ve *versionEdit) getInternalKey(input *[]byte) (internalKey, bool) {
	var dst []byte
	if utils.GetLenPrefixedBytes(&dst, &src) {
		return internalKey(dst), true
	} else {
		return nil, false
	}
}

func (ve *versionEdit) getLevel(input []byte) (int, bool) {
	v, _ := binary.Varint(input)
	if v < kNumLevels {
		return v, bool
	} else {
		return -1, false
	}
}

func (ve *versionEdit) decodeFrom(src []byte) error {
	ve.Clear()

	return nil
}
