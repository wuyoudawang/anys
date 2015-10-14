package cache

import (
	"encoding/binary"

	"anys/cache/iterator"
	"anys/cache/option"
	"anys/pkg/comparator"
)

const (
	kTargetFileSize int = 2 * 1048576 // 2M

	// Maximum bytes of overlaps in grandparent (i.e., level+2) before we
	// stop building a single file in a level->level+1 compaction.
	kMaxGrandParentOverlapBytes int64 = 10 * int64(kTargetFileSize) // 20M

	// Maximum number of bytes in all compacted files.  We avoid expanding
	// the lower level file set of a compaction if it would make the
	// total compaction cover more than this many bytes.
	kExpandedCompactionByteSizeLimit int64 = 25 * int64(kTargetFileSize) // 50M
)

func MaxBytesWithLevel(level int) uint64 {
	var rel uint64 = 10 * 1048576
	for level > 1 {
		rel *= 10
		level--
	}
	return rel
}

func MaxFileSizeWithLevel(level int) int {
	return kTargetFileSize
}

func totalFileSize(files []*fileMetaData) uint64 {
	var sum uint64 = 0
	for _, f := range files {
		sum += f.fileSize
	}
	return sum
}

func FindFile(cmp comparator.Comparator, files []*fileMetaData, key []byte) int {
	left := 0
	right := len(files)
	for left < right {
		mid := (left + right) / 2
		if cmp.Compare(files[mid].largest.encode(), key) < 0 {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return right
}

func afterFile(cmp comparator.Comparator, userKey []byte, file *fileMetaData) bool {
	return len(userKey) > 0 && file != nil && cmp.Compare(userKey, file.largest.userKey()) > 0
}

func beforeFile(cmp comparator.Comparator, userKey []byte, file *fileMetaData) bool {
	return len(userKey) > 0 && file != nil && cmp.Compare(userKey, file.largest.userKey()) < 0
}

func SomeFileOverlapsRange(cmp comparator.Comparator,
	disjointSortedFile bool,
	files []*fileMetaData,
	smallestUserKey []byte,
	largestUserKey []byte) bool {
	if disjointSortedFile {
		for _, file := range files {
			if afterFile(cmp, smallestUserKey, file) || beforeFile(cmp, largestUserKey, file) {

			} else {
				return true
			}
		}
		return false
	}

	index := 0
	if len(smallestUserKey) > 0 {
		small := newInternalKey(smallestUserKey, kMaxSeq, kValueTypeForSeek)
		index = FindFile(cmp, files, small)
	}

	if index >= len(files) {
		return false
	}

	return !beforeFile(cmp, largestUserKey, files[index])
}

type versionIterator struct {
	cmp    comparator.Comparator
	index  int
	flist  []*fileMetaData
	valBuf [16]byte
}

func (vi *versionIterator) Valid() bool {
	return vi.index < len(vi.flist)
}

func (vi *versionIterator) Seek(target []byte) {
	vi.index = FindFile(vi.cmp, vi.flist, target)
}

func (vi *versionIterator) First() {
	vi.index = 0
}

func (vi *versionIterator) Last() {
	if len(vi.flist) == 0 {
		vi.index = 0
	} else {
		vi.index = len(vi.flist) - 1
	}
}

func (vi *versionIterator) Next() {
	if vi.Valid() {
		vi.index++
	}
}

func (vi *versionIterator) Prev() {
	if vi.Valid() {
		if vi.index == 0 {
			vi.index = len(vi.flist)
		} else {
			vi.index--
		}
	}
}

func (vi *versionIterator) Key() []byte {
	if !vi.Valid() {
		return nil
	}
	return vi.flist[vi.index].largest.encode()
}

func (vi *versionIterator) Value() []byte {
	if !vi.Valid() {
		return nil
	}

	binary.LittleEndian.PutUint64(vi.valBuf[:], vi.flist[vi.index].number)
	binary.LittleEndian.PutUint64(vi.valBuf[8:], vi.flist[vi.index].fileSize)
	return vi.valBuf[:]
}

func (vi *versionIterator) Error() error {
	return nil
}

type Version struct {
	vset              *VesionSet
	prev              *Version
	next              *Version
	refs              int
	files             [kNumLevels]*fileMetaData
	fileToStrage      *fileMetaData
	fileToStrageLevel int
}

func (v *Version) NewConcatenatingIterator(readOpt *option.ReadOptions, level int) {

}

func (v *Version) AddIterators(readOpt *option.ReadOptions, iters *iterator.Interface) {

}

func (v *Version) GetOverlapsInput() {}

func (v *Version) Ref() {
	v.refs++
}

func (v *Version) Unref() {
	v.refs--
}

type VesionSet struct {
	dbname         string
	opt            *option.Options
	nextFileNumber uint64
	manifestNumber uint64
	lastSequence   uint64
	logNumber      uint64
	prevLogNumber  uint64
}
