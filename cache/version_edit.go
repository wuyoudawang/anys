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

func (ve *versionEdit) AddFile(level int, fileNumber, fileSize uint64, smallest, largest []byte) {
	// FileMetaData f;
	// f.number = file;
	// f.file_size = file_size;
	// f.smallest = smallest;
	// f.largest = largest;
	// new_files_.push_back(std::make_pair(level, f));
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
	// Clear();
	// Slice input = src;
	// const char* msg = NULL;
	// uint32_t tag;

	// // Temporary storage for parsing
	// int level;
	// uint64_t number;
	// FileMetaData f;
	// Slice str;
	// InternalKey key;

	// while (msg == NULL && GetVarint32(&input, &tag)) {
	//   switch (tag) {
	//     case kComparator:
	//       if (GetLengthPrefixedSlice(&input, &str)) {
	//         comparator_ = str.ToString();
	//         has_comparator_ = true;
	//       } else {
	//         msg = "comparator name";
	//       }
	//       break;

	//     case kLogNumber:
	//       if (GetVarint64(&input, &log_number_)) {
	//         has_log_number_ = true;
	//       } else {
	//         msg = "log number";
	//       }
	//       break;

	//     case kPrevLogNumber:
	//       if (GetVarint64(&input, &prev_log_number_)) {
	//         has_prev_log_number_ = true;
	//       } else {
	//         msg = "previous log number";
	//       }
	//       break;

	//     case kNextFileNumber:
	//       if (GetVarint64(&input, &next_file_number_)) {
	//         has_next_file_number_ = true;
	//       } else {
	//         msg = "next file number";
	//       }
	//       break;

	//     case kLastSequence:
	//       if (GetVarint64(&input, &last_sequence_)) {
	//         has_last_sequence_ = true;
	//       } else {
	//         msg = "last sequence number";
	//       }
	//       break;

	//     case kCompactPointer:
	//       if (GetLevel(&input, &level) &&
	//           GetInternalKey(&input, &key)) {
	//         compact_pointers_.push_back(std::make_pair(level, key));
	//       } else {
	//         msg = "compaction pointer";
	//       }
	//       break;

	//     case kDeletedFile:
	//       if (GetLevel(&input, &level) &&
	//           GetVarint64(&input, &number)) {
	//         deleted_files_.insert(std::make_pair(level, number));
	//       } else {
	//         msg = "deleted file";
	//       }
	//       break;

	//     case kNewFile:
	//       if (GetLevel(&input, &level) &&
	//           GetVarint64(&input, &f.number) &&
	//           GetVarint64(&input, &f.file_size) &&
	//           GetInternalKey(&input, &f.smallest) &&
	//           GetInternalKey(&input, &f.largest)) {
	//         new_files_.push_back(std::make_pair(level, f));
	//       } else {
	//         msg = "new-file entry";
	//       }
	//       break;

	//     default:
	//       msg = "unknown tag";
	//       break;
	//   }
	// }

	// if (msg == NULL && !input.empty()) {
	//   msg = "invalid tag";
	// }

	// Status result;
	// if (msg != NULL) {
	//   result = Status::Corruption("VersionEdit", msg);
	// }
	// return result;
}
