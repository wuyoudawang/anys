package cache

import (
	"anys/pkg/comparator"
)

const (
	kTargetFileSize int = 2 * 1048576 // 2M

	// Maximum bytes of overlaps in grandparent (i.e., level+2) before we
	// stop building a single file in a level->level+1 compaction.
	kMaxGrandParentOverlapBytes int64 = 10 * kTargetFileSize // 20M

	// Maximum number of bytes in all compacted files.  We avoid expanding
	// the lower level file set of a compaction if it would make the
	// total compaction cover more than this many bytes.
	kExpandedCompactionByteSizeLimit int64 = 25 * kTargetFileSize // 50M
)

func MaxBytesWithLevel(level int) uint64 {
	var rel uint64 = 10 * 1048576
	for level > 1 {
		rel *= 10
		level--
	}
	return rel
}

func MaxFileSizeWithLevel(level int) uint64 {
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

}

type versionIterator struct {
	cmp   comparator.Comparator
	index int
	flist []*fileMetaData
}

func (vi *versionIterator) Valid() bool {
	return vi.index < len(vi.flist)
}

func (vi *versionIterator) Seek(target []byte) {

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

type VesionSet struct {
	dbname         string
	opt            *Options
	nextFileNumber uint64
	manifestNumber uint64
	lastSequence   uint64
	logNumber      uint64
	prevLogNumber  uint64
}
