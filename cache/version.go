package cache

import (
	"encoding/binary"
	"sync"

	"anys/cache/iterator"
	"anys/cache/log"
	"anys/cache/option"
	"anys/cache/table"
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

type versionFileNumIterator struct {
	icmp   *iComparator
	index  int
	flist  []*fileMetaData
	valBuf [16]byte
}

func newVersionFileNumIter(icmp *iComparator, flist []*fileMetaData) *versionFileNumIterator {
	return *versionFileNumIterator{
		icmp:  icmp,
		index: len(flist),
		flist: flist,
	}
}

func (vi *versionFileNumIterator) Valid() bool {
	return vi.index < len(vi.flist)
}

func (vi *versionFileNumIterator) Seek(target []byte) {
	vi.index = FindFile(vi.cmp, vi.flist, target)
}

func (vi *versionFileNumIterator) First() {
	vi.index = 0
}

func (vi *versionFileNumIterator) Last() {
	if len(vi.flist) == 0 {
		vi.index = 0
	} else {
		vi.index = len(vi.flist) - 1
	}
}

func (vi *versionFileNumIterator) Next() {
	if vi.Valid() {
		vi.index++
	}
}

func (vi *versionFileNumIterator) Prev() {
	if vi.Valid() {
		if vi.index == 0 {
			vi.index = len(vi.flist)
		} else {
			vi.index--
		}
	}
}

func (vi *versionFileNumIterator) Key() []byte {
	if !vi.Valid() {
		return nil
	}
	return vi.flist[vi.index].largest.encode()
}

func (vi *versionFileNumIterator) Value() []byte {
	if !vi.Valid() {
		return nil
	}

	binary.LittleEndian.PutUint64(vi.valBuf[:], vi.flist[vi.index].number)
	binary.LittleEndian.PutUint64(vi.valBuf[8:], vi.flist[vi.index].fileSize)
	return vi.valBuf[:]
}

func (vi *versionFileNumIterator) Error() error {
	return nil
}

func getFileIterator(args interface{}, opt *option.ReadOptions, value []byte) iterator.Interface {
	cache, ok := args.(*tableCache)
	if !ok {
		panic("args error")
	}
	if len(value) != 16 {
		return iterator.NewEmptyIterator("FileReader invoked with unexpected value")
	} else {
		return cache.NewIterator()
	}
}

type Version struct {
	vset               *VesionSet
	prev               *Version
	next               *Version
	refs               int
	files              [kNumLevels]*fileMetaData
	fileToCompact      *fileMetaData
	fileToCompactLevel int
	compactionScore    int64
	compactionLevel    int
}

func newVersion() *Version {

}

func (v *Version) NewConcatenatingIterator(readOpt *option.ReadOptions, level int) {
	return table.NewTwoLevelIterator(
		newVersionFileNumIter(v.vset.table_cache, &files_[level]),
		getFileIterator, v.vset.table_cache, readOpt)
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

type VersionSet struct {
	dbname         string
	opt            *option.Options
	nextFileNumber uint64
	manifestNumber uint64
	lastSequence   uint64
	logNumber      uint64
	prevLogNumber  uint64

	current         *Version
	dummy_version   *Version
	descriptor_file *os.File
	descriptor_log  *log.Writer
	table_cache     *tableCache
}

func NewVersionSet() {
	return &VersionSet{
		lastSequence: 0,
	}
}

func (vs *VersionSet) NewFileNumber() uint64 {
	cur := vs.nextFileNumber
	vs.nextFileNumber++
	return cur
}

func (vs *VersionSet) PrevLogNumber() uint64 {
	return vs.prevLogNumber
}

func (vs *VersionSet) LogNUmber() uint64 {
	return vs.logNumber
}

func (vs *VersionSet) LastSequence() uint64 {
	return vs.lastSequence
}

func (vs *VersionSet) SetLastSequence(val uint64) {
	if val > vs.lastSequence {
		vs.lastSequence = val
	}
}

func (vs *VersionSet) NumLevelFiles(level int) int {
	if level >= 0 && level <= kNumLevels {
		return vs.current.files[level].fileSize
	} else {
		return 0
	}
}

func (vs *VersionSet) AppendVersion(v *Version) {
	if v.refs != 0 {
		return
	}
	if v == vs.current {
		return
	}

	if vs.current != nil {
		vs.current.Unref()
	}
	vs.current = v
	v.Ref()

	v.prev = vs.dummy_version.prev
	v.next = vs.dummy_version
	v.prer.next = v
	v.next.prev = v
}

func (vs *VersionSet) LogAndApply(edit *versionEdit, mu *sync.Mutex) {
	if edit.hasLogNumber {
		if edit.logNumber < vs.logNumber {
			return
		}
		if edit.logNumber >= vs.nextFileNumber {
			return
		}
	} else {
		edit.SetLogNumber(vs.logNumber)
	}

	if !edit.hasPrevLogNumber {
		edit.SetPrevLogNumber(vs.prevLogNumber)
	}

	edit.SetNextFileNumber(vs.nextFileNumber)
	edit.SetSequence(vs.lastSequence)

	// v := new

	//  if (!edit->has_prev_log_number_) {
	//    edit->SetPrevLogNumber(prev_log_number_);
	//  }

	//  edit->SetNextFile(next_file_number_);
	//  edit->SetLastSequence(last_sequence_);

	//  Version* v = new Version(this);
	//  {
	//    Builder builder(this, current_);
	//    builder.Apply(edit);
	//    builder.SaveTo(v);
	//  }
	//  Finalize(v);

	//  // Initialize new descriptor log file if necessary by creating
	//  // a temporary file that contains a snapshot of the current version.
	//  std::string new_manifest_file;
	//  Status s;
	//  if (descriptor_log_ == NULL) {
	//    // No reason to unlock *mu here since we only hit this path in the
	//    // first call to LogAndApply (when opening the database).
	//    assert(descriptor_file_ == NULL);
	//    new_manifest_file = DescriptorFileName(dbname_, manifest_file_number_);
	//    edit->SetNextFile(next_file_number_);
	//    s = env_->NewWritableFile(new_manifest_file, &descriptor_file_);
	//    if (s.ok()) {
	//      descriptor_log_ = new log::Writer(descriptor_file_);
	//      s = WriteSnapshot(descriptor_log_);
	//    }
	//  }

	//  // Unlock during expensive MANIFEST log write
	//  {
	//    mu->Unlock();

	//    // Write new record to MANIFEST log
	//    if (s.ok()) {
	//      std::string record;
	//      edit->EncodeTo(&record);
	//      s = descriptor_log_->AddRecord(record);
	//      if (s.ok()) {
	//        s = descriptor_file_->Sync();
	//      }
	//      if (!s.ok()) {
	//        Log(options_->info_log, "MANIFEST write: %s\n", s.ToString().c_str());
	//      }
	//    }

	//    // If we just created a new descriptor file, install it by writing a
	//    // new CURRENT file that points to it.
	//    if (s.ok() && !new_manifest_file.empty()) {
	//      s = SetCurrentFile(env_, dbname_, manifest_file_number_);
	//    }

	//    mu->Lock();
	//  }

	//  // Install the new version
	//  if (s.ok()) {
	//    AppendVersion(v);
	//    log_number_ = edit->log_number_;
	//    prev_log_number_ = edit->prev_log_number_;
	//  } else {
	//    delete v;
	//    if (!new_manifest_file.empty()) {
	//      delete descriptor_log_;
	//      delete descriptor_file_;
	//      descriptor_log_ = NULL;
	//      descriptor_file_ = NULL;
	//      env_->DeleteFile(new_manifest_file);
	//    }
	//  }

	//  return s;
}

type levelState struct {
	delete_files []int64
	added_files  *fileSet
}

type builder struct {
	vset   *VersionSet
	base   *Version
	levels [kNumLevels]levelState
}

func (b *builder) apply(edit *versionEdit) {
	// for (size_t i = 0; i < edit->compact_pointers_.size(); i++) {
	//     const int level = edit->compact_pointers_[i].first;
	//     vset_->compact_pointer_[level] =
	//         edit->compact_pointers_[i].second.Encode().ToString();
	//   }

	//   // Delete files
	//   const VersionEdit::DeletedFileSet& del = edit->deleted_files_;
	//   for (VersionEdit::DeletedFileSet::const_iterator iter = del.begin();
	//        iter != del.end();
	//        ++iter) {
	//     const int level = iter->first;
	//     const uint64_t number = iter->second;
	//     levels_[level].deleted_files.insert(number);
	//   }

	//   // Add new files
	//   for (size_t i = 0; i < edit->new_files_.size(); i++) {
	//     const int level = edit->new_files_[i].first;
	//     FileMetaData* f = new FileMetaData(edit->new_files_[i].second);
	//     f->refs = 1;

	//     // We arrange to automatically compact this file after
	//     // a certain number of seeks.  Let's assume:
	//     //   (1) One seek costs 10ms
	//     //   (2) Writing or reading 1MB costs 10ms (100MB/s)
	//     //   (3) A compaction of 1MB does 25MB of IO:
	//     //         1MB read from this level
	//     //         10-12MB read from next level (boundaries may be misaligned)
	//     //         10-12MB written to next level
	//     // This implies that 25 seeks cost the same as the compaction
	//     // of 1MB of data.  I.e., one seek costs approximately the
	//     // same as the compaction of 40KB of data.  We are a little
	//     // conservative and allow approximately one seek for every 16KB
	//     // of data before triggering a compaction.
	//     f->allowed_seeks = (f->file_size / 16384);
	//     if (f->allowed_seeks < 100) f->allowed_seeks = 100;

	//     levels_[level].deleted_files.erase(f->number);
	//     levels_[level].added_files->insert(f);
	//   }
}

func (b *builder) saveTo(v *Version) {
	// 	BySmallestKey cmp;
	//     cmp.internal_comparator = &vset_->icmp_;
	//     for (int level = 0; level < config::kNumLevels; level++) {
	//       // Merge the set of added files with the set of pre-existing files.
	//       // Drop any deleted files.  Store the result in *v.
	//       const std::vector<FileMetaData*>& base_files = base_->files_[level];
	//       std::vector<FileMetaData*>::const_iterator base_iter = base_files.begin();
	//       std::vector<FileMetaData*>::const_iterator base_end = base_files.end();
	//       const FileSet* added = levels_[level].added_files;
	//       v->files_[level].reserve(base_files.size() + added->size());
	//       for (FileSet::const_iterator added_iter = added->begin();
	//            added_iter != added->end();
	//            ++added_iter) {
	//         // Add all smaller files listed in base_
	//         for (std::vector<FileMetaData*>::const_iterator bpos
	//                  = std::upper_bound(base_iter, base_end, *added_iter, cmp);
	//              base_iter != bpos;
	//              ++base_iter) {
	//           MaybeAddFile(v, level, *base_iter);
	//         }

	//         MaybeAddFile(v, level, *added_iter);
	//       }

	//       // Add remaining base files
	//       for (; base_iter != base_end; ++base_iter) {
	//         MaybeAddFile(v, level, *base_iter);
	//       }

	// #ifndef NDEBUG
	//       // Make sure there is no overlap in levels > 0
	//       if (level > 0) {
	//         for (uint32_t i = 1; i < v->files_[level].size(); i++) {
	//           const InternalKey& prev_end = v->files_[level][i-1]->largest;
	//           const InternalKey& this_begin = v->files_[level][i]->smallest;
	//           if (vset_->icmp_.Compare(prev_end, this_begin) >= 0) {
	//             fprintf(stderr, "overlapping ranges in same level %s vs. %s\n",
	//                     prev_end.DebugString().c_str(),
	//                     this_begin.DebugString().c_str());
	//             abort();
	//           }
	//         }
	//       }
	// #endif
	//     }
}

func (b *builder) maybeAddFile(v *Version, level int, f *fileMetaData) {
	// if (levels_[level].deleted_files.count(f->number) > 0) {
	//     // File is deleted: do nothing
	//   } else {
	//     std::vector<FileMetaData*>* files = &v->files_[level];
	//     if (level > 0 && !files->empty()) {
	//       // Must not overlap
	//       assert(vset_->icmp_.Compare((*files)[files->size()-1]->largest,
	//                                   f->smallest) < 0);
	//     }
	//     f->refs++;
	//     files->push_back(f);
	//   }
}

type compaction struct {
	level                int
	max_output_file_size uint64
	input_version        *Version
	edit                 versionEdit
	input                [2]*fileMetaData
	grandparents         *fileMetaData
	grandparent_index    int64
	seen_key             bool
	overlapped_bytes     int64
	level_ptrs           [kNumLevels]int64
}

func newCompaction(level int) *compaction {
	c := &compaction{
		level:                level,
		max_output_file_size: MaxFileSizeWithLevel(level),
		input_version:        nil,
		grandparent_index:    0,
		seen_key:             false,
	}
	for i := 0; i < kNumLevels; i++ {
		c.level_ptrs[i] = 0
	}
	return c
}

func (c *compaction) numInputFiles(which int) int {
	return len(c.input[which])
}

func (c *compaction) isTrivialMove() bool {
	return (c.numInputFiles(0) == 1 && c.numInputFiles(1) == 0 &&
		totalFileSize(c.grandparents) <= kMaxGrandParentOverlapBytes)
}

func (c *compaction) addInputDeletions(edit *versionEdit) {
	for which := 0; which < 2; which++ {
		for i := 0; i < c.numInputFiles(which); i++ {
			edit.deleteFile(level+which, c.input[which][i].number)
		}
	}
}

func (c *compaction) isBaseLevelForKey(user_key []byte) bool {
	// Maybe use binary search to find right entry instead of linear search?
	// const Comparator* user_cmp = input_version_->vset_->icmp_.user_comparator();
	// for (int lvl = level_ + 2; lvl < config::kNumLevels; lvl++) {
	//   const std::vector<FileMetaData*>& files = input_version_->files_[lvl];
	//   for (; level_ptrs_[lvl] < files.size(); ) {
	//     FileMetaData* f = files[level_ptrs_[lvl]];
	//     if (user_cmp->Compare(user_key, f->largest.user_key()) <= 0) {
	//       // We've advanced far enough
	//       if (user_cmp->Compare(user_key, f->smallest.user_key()) >= 0) {
	//         // Key falls in this file's range, so definitely not base level
	//         return false;
	//       }
	//       break;
	//     }
	//     level_ptrs_[lvl]++;
	//   }
	// }
	// return true;
}

// bool Compaction::ShouldStopBefore(const Slice& internal_key) {
//   // Scan to find earliest grandparent file that contains key.
//   const InternalKeyComparator* icmp = &input_version_->vset_->icmp_;
//   while (grandparent_index_ < grandparents_.size() &&
//       icmp->Compare(internal_key,
//                     grandparents_[grandparent_index_]->largest.Encode()) > 0) {
//     if (seen_key_) {
//       overlapped_bytes_ += grandparents_[grandparent_index_]->file_size;
//     }
//     grandparent_index_++;
//   }
//   seen_key_ = true;

//   if (overlapped_bytes_ > kMaxGrandParentOverlapBytes) {
//     // Too much overlap for current output; start new output
//     overlapped_bytes_ = 0;
//     return true;
//   } else {
//     return false;
//   }
// }
