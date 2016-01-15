package cache

import (
	"sync"
	"time"
	"unsafe"

	"anys/cache/log"
	"anys/pkg/utils"
)

const (
	kTypeDeletion     = 0x0
	kTypeValue        = 0x1
	kValueTypeForSeek = kTypeValue

	kNumLevels int = 7

	// Level-0 compaction is started when we hit this many files.
	kL0_CompactionTrigger int = 4

	// Soft limit on number of level-0 files.  We slow down writes at this point.
	kL0_SlowdownWritesTrigger int = 8

	// Maximum number of level-0 files.  We stop writes at this point.
	kL0_StopWritesTrigger int = 12

	// Maximum level to which a new compacted memtable is pushed if it
	// does not create overlap.  We try to push to level 2 to avoid the
	// relatively expensive level 0=>1 compactions and to avoid some
	// expensive manifest file operations.  We do not push all the way to
	// the largest level since that can generate a lot of wasted disk
	// space if the same key space is being repeatedly overwritten.
	kMaxMemCompactLevel int = 2

	// Approximate gap in bytes between samples of data read during iteration.
	kReadBytesPeriod int = 1048576
)

type compactionState struct {
	cpact             *compaction
	smallest_snaphost int64
	outputs           []struct {
		number            int64
		file_size         int64
		smallest, largest internalKey
	}
	bld         *builder
	outfile     os.File
	total_bytes uint64
}

type BatchWriter struct {
	err   error
	batch *Batch
	sync  bool
	done  bool
	mc    *sync.Cond
}

type manualCompaction struct {
	level       int
	done        bool
	begin       *internalKey
	end         *internalKey
	tmp_storage internalKey
}

type DB struct {
	mem            *MemTable
	imm            *MemTable
	log            *log.Writer
	mu             sync.Mutex
	tmpBatch       *Batch
	opt            *Options
	writes         *utils.Queue
	version        *VersionSet
	pending_output []uint64
	dbname         string
	table_cache    *tableCache
}

func (d *DB) AddStorage() {

}

func (d *DB) MaybeScheduleStorage() {

}

func (d *DB) MergeBatch(lastWriter **BatchWriter) *Batch {
	if d.writes.Empty() {
		return nil
	}

	first := (*BatchWriter)(d.writes.Head())
	result := first.batch
	size := first.batch.Size()

	// Allow the group to grow up to a maximum size, buf if the
	// original write is small, limit the growth so we do not slow
	// down the small write too much
	maxSize := 1 << 20
	if size <= (128 << 10) {
		maxSize = size + (128 << 10)
	}

	*lastWriter = first
	for i := 0; int64(i) < d.writes.Size(); i++ {
		w := (*BatchWriter)(d.writes.Index(i))
		if w.sync && !first.sync {
			break
		}

		if w.batch != nil {
			size += w.batch.Size()
			if size > maxSize {
				break
			}

			if result == first.batch {

			}
			result.Append(w.batch)
		}
		*lastWriter = w
	}

	return result

	//  *last_writer = first;
	//  std::deque<Writer*>::iterator iter = writers_.begin();
	//  ++iter;  // Advance past "first"
	//  for (; iter != writers_.end(); ++iter) {
	//    Writer* w = *iter;
	//    if (w->sync && !first->sync) {
	//      // Do not include a sync write into a batch handled by a non-sync write.
	//      break;
	//    }

	//    if (w->batch != NULL) {
	//      size += WriteBatchInternal::ByteSize(w->batch);
	//      if (size > max_size) {
	//        // Do not make batch too big
	//        break;
	//      }

	//      // Append to *result
	//      if (result == first->batch) {
	//        // Switch to temporary batch instead of disturbing caller's batch
	//        result = tmp_batch_;
	//        assert(WriteBatchInternal::Count(result) == 0);
	//        WriteBatchInternal::Append(result, first->batch);
	//      }
	//      WriteBatchInternal::Append(result, w->batch);
	//    }
	//    *last_writer = w;
	//  }
	//  return result;
}

func (d *DB) Write(b *Batch, opt *WriteOptions) error {
	var err error
	w := new(BatchWriter)
	w.batch = b
	w.sync = opt.Sync
	w.done = false
	w.mc = sync.NewCond(&d.mu)

	d.mu.Lock()
	d.writes.Push(unsafe.Pointer(w))
	for !w.done && unsafe.Pointer(w) != d.writes.Head() {
		w.mc.Wait()
	}

	if w.done {
		return w.err
	}

	lastWriter := w
	if b != nil {
		updates := d.MergeBatch(&lastWriter)
		// updates.setSequence(lastSequence+1)
		d.mu.Unlock()
		err = d.log.AddRecord(updates.Contents())
		if err == nil && opt.Sync {
			//d.log.Sync()
		}

		if err == nil {
			updates.InsertInto(d.mem)
		}
		d.mu.Lock()
		if updates == d.tmpBatch {
			d.tmpBatch.Clear()
		}
	}

	for {
		ready := (*BatchWriter)(d.writes.Pop())
		if ready != w {
			ready.err = err
			ready.done = true
			ready.mc.Signal()
		}

		if ready == lastWriter {
			break
		}
	}

	if !d.writes.Empty() {
		(*BatchWriter)(d.writes.Head()).mc.Signal()
	}

	return err

	// // May temporarily unlock and wait.
	// Status status = MakeRoomForWrite(my_batch == NULL);
	// uint64_t last_sequence = versions_->LastSequence();
	// Writer* last_writer = &w;
	// if (status.ok() && my_batch != NULL) {  // NULL batch is for compactions
	//   WriteBatch* updates = BuildBatchGroup(&last_writer);
	//   WriteBatchInternal::SetSequence(updates, last_sequence + 1);
	//   last_sequence += WriteBatchInternal::Count(updates);

	//   // Add to log and apply to memtable.  We can release the lock
	//   // during this phase since &w is currently responsible for logging
	//   // and protects against concurrent loggers and concurrent writes
	//   // into mem_.
	//   {
	//     mutex_.Unlock();
	//     status = log_->AddRecord(WriteBatchInternal::Contents(updates));
	//     bool sync_error = false;
	//     if (status.ok() && options.sync) {
	//       status = logfile_->Sync();
	//       if (!status.ok()) {
	//         sync_error = true;
	//       }
	//     }
	//     if (status.ok()) {
	//       status = WriteBatchInternal::InsertInto(updates, mem_);
	//     }
	//     mutex_.Lock();
	//     if (sync_error) {
	//       // The state of the log file is indeterminate: the log record we
	//       // just added may or may not show up when the DB is re-opened.
	//       // So we force the DB into a mode where all future writes fail.
	//       RecordBackgroundError(status);
	//     }
	//   }
	//   if (updates == tmp_batch_) tmp_batch_->Clear();

	//   versions_->SetLastSequence(last_sequence);
	// }

	// while (true) {
	//   Writer* ready = writers_.front();
	//   writers_.pop_front();
	//   if (ready != &w) {
	//     ready->status = status;
	//     ready->done = true;
	//     ready->cv.Signal();
	//   }
	//   if (ready == last_writer) break;
	// }

	// // Notify new head of write queue
	// if (!writers_.empty()) {
	//   writers_.front()->cv.Signal();
	// }

	// return status;
}

func (d *DB) makeRoomForWrite(force bool) {
	// mutex_.AssertHeld();
	// assert(!writers_.empty());
	// bool allow_delay = !force;
	// Status s;
	// while (true) {
	//   if (!bg_error_.ok()) {
	//     // Yield previous error
	//     s = bg_error_;
	//     break;
	//   } else if (
	//       allow_delay &&
	//       versions_->NumLevelFiles(0) >= config::kL0_SlowdownWritesTrigger) {
	//     // We are getting close to hitting a hard limit on the number of
	//     // L0 files.  Rather than delaying a single write by several
	//     // seconds when we hit the hard limit, start delaying each
	//     // individual write by 1ms to reduce latency variance.  Also,
	//     // this delay hands over some CPU to the compaction thread in
	//     // case it is sharing the same core as the writer.
	//     mutex_.Unlock();
	//     env_->SleepForMicroseconds(1000);
	//     allow_delay = false;  // Do not delay a single write more than once
	//     mutex_.Lock();
	//   } else if (!force &&
	//              (mem_->ApproximateMemoryUsage() <= options_.write_buffer_size)) {
	//     // There is room in current memtable
	//     break;
	//   } else if (imm_ != NULL) {
	//     // We have filled up the current memtable, but the previous
	//     // one is still being compacted, so we wait.
	//     Log(options_.info_log, "Current memtable full; waiting...\n");
	//     bg_cv_.Wait();
	//   } else if (versions_->NumLevelFiles(0) >= config::kL0_StopWritesTrigger) {
	//     // There are too many level-0 files.
	//     Log(options_.info_log, "Too many L0 files; waiting...\n");
	//     bg_cv_.Wait();
	//   } else {
	//     // Attempt to switch to a new memtable and trigger compaction of old
	//     assert(versions_->PrevLogNumber() == 0);
	//     uint64_t new_log_number = versions_->NewFileNumber();
	//     WritableFile* lfile = NULL;
	//     s = env_->NewWritableFile(LogFileName(dbname_, new_log_number), &lfile);
	//     if (!s.ok()) {
	//       // Avoid chewing through file number space in a tight loop.
	//       versions_->ReuseFileNumber(new_log_number);
	//       break;
	//     }
	//     delete log_;
	//     delete logfile_;
	//     logfile_ = lfile;
	//     logfile_number_ = new_log_number;
	//     log_ = new log::Writer(lfile);
	//     imm_ = mem_;
	//     has_imm_.Release_Store(imm_);
	//     mem_ = new MemTable(internal_comparator_);
	//     mem_->Ref();
	//     force = false;   // Do not force another compaction if have room
	//     MaybeScheduleCompaction();
	//   }
	// }
	// return s;
}

func (d *DB) WriteLevel0Table(mem *MemTable, edit *versionEdit, base *Version) {
	d.mu.Lock()
	defer d.mu.Unlock()

	startMicros := time.Now().Nanosecond() / 1000
	var meta fileMetaData
	meta.number = d.version.NewFileNumber()
	d.pending_output = append(meta.number)
	iter := mem.Iterator()
	//log(options_.info_log, "Level-0 table #%llu: started",(unsigned long long) meta.number)
	d.mu.Unlock()
	err := buildTable(d.dbname, d.opt, d.table_cache, iter, meta)
	if err != nil {
		return err
	}
	d.mu.Lock()

	d.pending_output[meta.number] //delete this meta pending

	level := 0
	if meta.fileSize > 0 {
		min_user_key := meta.smallest.userKey()
		max_user_key := meta.largest.userKey()
		if base != nil {
			level = base.fileToStrageLevel
		}
		edit.AddFile(level, meta.number, meta.fileSize, meta.smallest, meta.largest)
	}
}

func (d *DB) CompactMemTable() {
	var edit versionEdit
	base := d.version.current
	base.Ref()
	d.WriteLevel0Table(d.mem, &edit, base)
	base.Unref()

	if err == nil {
		edit.SetPrevLogNumber(0)
		edit.SetLogNumber(d.logfile_number)
		err = d.version.LogAndApply(edit, &d.mu)
	}
}
