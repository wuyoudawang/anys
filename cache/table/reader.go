package table

import (
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/golang/snappy"

	"anys/cache/option"
	"anys/pkg/utils"
)

var (
	ErrNotFound       = errors.ErrNotFound
	ErrReaderReleased = errors.New("leveldb/table: reader released")
	ErrIterReleased   = errors.New("leveldb/table: iterator released")
)

type ErrCorrupted struct {
	Pos    int64
	Size   int64
	Kind   string
	Reason string
}

func (e *ErrCorrupted) Error() string {
	return fmt.Sprintf("leveldb/table: corruption on %s (pos=%d): %s", e.Kind, e.Pos, e.Reason)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

type block struct {
	bpool          *utils.BufferPool
	bh             blockHandle
	data           []byte
	restartsLen    int
	restartsOffset int
}

// Iterator* Block::NewIterator(const Comparator* cmp) {
//   if (size_ < sizeof(uint32_t)) {
//     return NewErrorIterator(Status::Corruption("bad block contents"));
//   }
//   const uint32_t num_restarts = NumRestarts();
//   if (num_restarts == 0) {
//     return NewEmptyIterator();
//   } else {
//     return new Iter(cmp, data_, restart_offset_, num_restarts);
//   }
// }

func decodeEntry(data []byte) (key, value []byte, nShared, n int, err error) {
	if len(data) < 3 {
		return
	}
	v0, n0 := binary.Uvarint(data)
	v1, n1 := binary.Uvarint(data[n0:])
	v2, n2 := binary.Uvarint(data[n0+n1:])
	m := n0 + n1 + n2
	n = m + int(v1) + int(v2)
	if n0 <= 0 || n1 <= 0 || n2 <= 0 || len(data[m:]) > v0+v2 {
		err = &ErrCorrupted{Reason: "entries corrupted"}
		return
	}
	key = data[m : m+v1]
	value = data[m+v1 : n]
}

type blockIter struct {
	cmp         comparer.Comparer
	data        []byte
	restarts    uint32
	numRestarts uint32

	current      uint32
	restartIndex uint32
	key, value   []byte
	err          error
}

func (i *blockIter) nextEntryOffset() uint32 {
	return i.current + len(i.value)
}

func (i *blockIter) getRestartOffset(index uint32) uint32 {
	return binary.LittleEndian.Uint32(i.data[i.restarts+4*index:])
}

func (i *blockIter) seekToRestartOffset(index uint32) {
	i.key = i.key[:0]
	i.restartIndex = index
	offset := i.getRestartOffset(index)
	i.value = i.data[offset:0]
	// restart_index_ = index;
	// // current_ will be fixed by ParseNextKey();

	// // ParseNextKey() starts at the end of value_, so set value_ accordingly
	// uint32_t offset = GetRestartPoint(index);
	// value_ = Slice(data_ + offset, 0);
}

func (i *blockIter) Valid() bool {
	return i.current < i.restarts
}

func (i *blockIter) Error() error {
	return i.err
}

func (i *blockIter) Value() []byte {
	if !i.Valid() {
		return nil
	}
	return i.value
}

func (i *blockIter) Key() []byte {
	return i.key
}

func (i *blockIter) Next() {
	i.parseNextKey()
}

func (i *blockIter) Prev() {
	original := i.current
	for i.getRestartOffset(i.restartIndex) >= original {
		if i.restartIndex == 0 {
			i.current = i.restarts
			i.restartIndex = i.numRestarts
			return
		}
		i.restartIndex--
	}

	i.seekToRestartOffset(i.restartIndex)
	for i.parseNextKey() && i.nextEntryOffset() < original {
	}
}

func (i *blockIter) Seek(key []byte) {
	// Binary search in restart array to find the last restart point
	// with a key < target
	left := 0
	right := i.numRestarts - 1
	for left < right {
		mid := (left + right + 1) / 2
		region_offset := i.getRestartOffset(mid)
		midkey, value, shared, n, err := decodeEntry(i.data[region_offset:i.restarts])
		if err != nil {
			i.err = err
			return
		}
		if i.cmp.Compare(midkey, key) < 0 {
			// key at 'mid' is smaller than 'target'. therefore all
			// blocks before 'mid' are uninteresting.
			left = mid
		} else {
			right = mid - 1
		}
	}
	//    // with a key < target
	//    uint32_t left = 0;
	//    uint32_t right = num_restarts_ - 1;
	//    while (left < right) {
	//      uint32_t mid = (left + right + 1) / 2;
	//      uint32_t region_offset = GetRestartPoint(mid);
	//      uint32_t shared, non_shared, value_length;
	//      const char* key_ptr = DecodeEntry(data_ + region_offset,
	//                                        data_ + restarts_,
	//                                        &shared, &non_shared, &value_length);
	//      if (key_ptr == NULL || (shared != 0)) {
	//        CorruptionError();
	//        return;
	//      }
	//      Slice mid_key(key_ptr, non_shared);
	//      if (Compare(mid_key, target) < 0) {
	//        // Key at "mid" is smaller than "target".  Therefore all
	//        // blocks before "mid" are uninteresting.
	//        left = mid;
	//      } else {
	//        // Key at "mid" is >= "target".  Therefore all blocks at or
	//        // after "mid" are uninteresting.
	//        right = mid - 1;
	//      }

	// Linear search (within restart block) for first key >= target
	i.seekToRestartOffset(left)
	for {
		if !i.parseNextKey() {
			return
		}
		if i.cmp.Compare(i.key, key) >= 0 {
			return
		}

	}
}

func (i *blockIter) First() {
	i.seekToRestartOffset(0)
	i.parseNextKey()
}

func (i *blockIter) Last() {
	i.seekToRestartOffset(i.numRestarts - 1)
	for i.parseNextKey() && i.nextEntryOffset() < i.restarts {
	}
}

func (i *blockIter) parseNextKey() bool {
	i.current = i.nextEntryOffset()
	limit := i.restarts
	if i.current >= limit {
		i.current = i.restarts
		i.restartIndex = i.numRestarts
		return false
	}

	if key, value, shared, n, err := decodeEntry(i.data[i.current:limit]); err != nil {
		i.err = err
		return false
	} else {
		i.key = i.key[:shared]
		i.key = append(i.key, key...)
		i.value = value
		for i.restartIndex+1 < i.numRestarts && i.getRestartOffset(i.restartIndex+1) < i.current {
			i.restartIndex++
		}
		return true
	}
}

type filterBlock struct {
	bpool      *util.BufferPool
	data       []byte
	oOffset    int
	baseLg     uint
	filtersNum int
}

func (b *filterBlock) contains(filter filter.Filter, offset uint64, key []byte) bool {
	i := int(offset >> b.baseLg)
	if i < b.filtersNum {
		o := b.data[b.oOffset+i*4:]
		n := int(binary.LittleEndian.Uint32(o))
		m := int(binary.LittleEndian.Uint32(o[4:]))
		if n < m && m <= b.oOffset {
			return filter.Contains(b.data[n:m], key)
		} else if n == m {
			return false
		}
	}
	return true
}

func (b *filterBlock) Release() {
	b.bpool.Put(b.data)
	b.bpool = nil
	b.data = nil
}

type Reader struct {
	opt             *option.Options
	err             error
	rd              io.ReaderAt
	cacheId         uint64
	filter          *filterBlock
	filterData      []byte
	indexBlock      *blockIter
	metaindexHandle blockHandle
}

func NewReader(opt *option.Options, rd io.ReaderAt, size uint64) (*Reader, error) {
	r := new(Reader)
	if size < footerLen {
		r.err = r.newErrCorrupted(0, size, "table", "too small")
		return r, nil
	}

	footerPos := size - footerLen
	var footer [footerLen]byte
	if _, err := rd.ReadAt(footer, footerPos); err != nil {
		return nil, err
	}
	if string(footer[footerLen-len(magic):footerLen]) != magic {
		r.err = r.newErrCorrupted(footerPos, footerLen, "table-footer", "bad magic number")
		return r, nil
	}

	var n int
	r.metaindexHandle, n = decodeBlockHandle(footer)
	if n == 0 {
		r.err = r.newErrCorrupted(footerPos, footerLen, "table-footer", "bad metaindex block handle")
		return r, nil
	}

	// Decode the index block handle.
	r.indexBH, n = decodeBlockHandle(footer[n:])
	if n == 0 {
		r.err = r.newErrCorrupted(footerPos, footerLen, "table-footer", "bad index block handle")
		return r, nil
	}

	// Read metaindex block.
	metaBlock, err := r.readBlock(r.metaBH, true)
	if err != nil {
		if errors.IsCorrupted(err) {
			r.err = err
			return r, nil
		} else {
			return nil, err
		}
	}
}

type blockContent struct {
	data     []byte
	cachable bool
}

func ReadBlock(rd io.ReaderAt, rdOpt *option.ReadOptions, handle *blockHandle, result *blockContent) *block {
	result.data = make([]byte, 0)
	result.cachable = false
	// result->data = Slice();
	// result->cachable = false;
	// result->heap_allocated = false;

	// // Read the block contents as well as the type/crc footer.
	// // See table_builder.cc for the code that built this structure.
	// size_t n = static_cast<size_t>(handle.size());
	// char* buf = new char[n + kBlockTrailerSize];
	// Slice contents;
	// Status s = file->Read(handle.offset(), n + kBlockTrailerSize, &contents, buf);
	// if (!s.ok()) {
	//   delete[] buf;
	//   return s;
	// }
	// if (contents.size() != n + kBlockTrailerSize) {
	//   delete[] buf;
	//   return Status::Corruption("truncated block read");
	// }

	// // Check the crc of the type and the block contents
	// const char* data = contents.data();    // Pointer to where Read put the data
	// if (options.verify_checksums) {
	//   const uint32_t crc = crc32c::Unmask(DecodeFixed32(data + n + 1));
	//   const uint32_t actual = crc32c::Value(data, n + 1);
	//   if (actual != crc) {
	//     delete[] buf;
	//     s = Status::Corruption("block checksum mismatch");
	//     return s;
	//   }
	// }

	// switch (data[n]) {
	//   case kNoCompression:
	//     if (data != buf) {
	//       // File implementation gave us pointer to some other data.
	//       // Use it directly under the assumption that it will be live
	//       // while the file is open.
	//       delete[] buf;
	//       result->data = Slice(data, n);
	//       result->heap_allocated = false;
	//       result->cachable = false;  // Do not double-cache
	//     } else {
	//       result->data = Slice(buf, n);
	//       result->heap_allocated = true;
	//       result->cachable = true;
	//     }

	//     // Ok
	//     break;
	//   case kSnappyCompression: {
	//     size_t ulength = 0;
	//     if (!port::Snappy_GetUncompressedLength(data, n, &ulength)) {
	//       delete[] buf;
	//       return Status::Corruption("corrupted compressed block contents");
	//     }
	//     char* ubuf = new char[ulength];
	//     if (!port::Snappy_Uncompress(data, n, ubuf)) {
	//       delete[] buf;
	//       delete[] ubuf;
	//       return Status::Corruption("corrupted compressed block contents");
	//     }
	//     delete[] buf;
	//     result->data = Slice(ubuf, ulength);
	//     result->heap_allocated = true;
	//     result->cachable = true;
	//     break;
	//   }
	//   default:
	//     delete[] buf;
	//     return Status::Corruption("bad block type");
	// }

	// return Status::OK();
}
