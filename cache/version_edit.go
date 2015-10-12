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

type compactKey struct {
	level int
	key   internalKey
}

type versionEdit struct {
	comparatorName    []byte
	logNumber         uint64
	prevLogNumber     uint64
	nextFileNumber    uint64
	lastSequence      uint64
	hasComparator     bool
	hasLogNumber      bool
	hasPrevLogNumber  bool
	hasNextFileNumber bool
	hasLastSequence   bool

	compactPointers []*compactKey
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
	ve.comparatorName = []byte(name)
}

func (ve *versionEdit) SetCompactPointer(ptr *compactKey) {
	ve.compactPointers = append(ve.compactPointers, ptr)
}

func (ve *versionEdit) Clear() {
	ve.comparatorName = ve.comparatorName[:0]
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
		binary.PutVarint(*dst, kComparator)
		utils.PutLenPrefixedBytes(dst, &ve.comparatorName)
	}

	if ve.hasLogNumber {
		binary.PutVarint(*dst, kLogNumber)
		binary.PutVarint(*dst, int64(ve.logNumber))
	}

	if ve.hasPrevLogNumber {
		binary.PutVarint(*dst, kPrevLogNumber)
		binary.PutVarint(*dst, int64(ve.prevLogNumber))
	}

	if ve.hasNextFileNumber {
		binary.PutVarint(*dst, kNextFileNumber)
		binary.PutVarint(*dst, int64(ve.nextFileNumber))
	}

	if ve.hasLastSequence {
		binary.PutVarint(*dst, kLastSequence)
		binary.PutVarint(*dst, int64(ve.lastSequence))
	}

	for _, cp := range ve.compactPointers {
		binary.PutVarint(*dst, kCompactPointer)
		binary.PutVarint(*dst, int64(cp.level))
		src := cp.key.encode()
		utils.PutLenPrefixedBytes(dst, &src)
	}

	for _, ditem := range ve.deletFiles {
		binary.PutVarint(*dst, kDeletedFile)
		binary.PutVarint(*dst, int64(ditem.level))
		binary.PutVarint(*dst, int64(ditem.number))
	}

	for _, metaData := range ve.newFiles {
		binary.PutVarint(*dst, kNewFile)
		binary.PutVarint(*dst, int64(metaData.level)) // level
		binary.PutVarint(*dst, int64(metaData.number))
		binary.PutVarint(*dst, int64(metaData.fileSize))
		src := metaData.smallest.encode()
		utils.PutLenPrefixedBytes(dst, &src)
		src = metaData.largest.encode()
		utils.PutLenPrefixedBytes(dst, &src)
	}
}

func (ve *versionEdit) getInternalKey(input *[]byte) (internalKey, bool) {
	var dst []byte
	if utils.GetLenPrefixedBytes(&dst, input) {
		return internalKey(dst), true
	} else {
		return nil, false
	}
}

func (ve *versionEdit) getLevel(input []byte) (int, bool) {
	v, _ := binary.Varint(input)
	if v < int64(kNumLevels) {
		return int(v), true
	} else {
		return -1, false
	}
}

func (ve *versionEdit) decodeFrom(src []byte) error {
	ve.Clear()

	return nil
}
