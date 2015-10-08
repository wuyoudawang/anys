package cache

import (
	"os"

	"anys/cache/log"
)

const (
	TypeDeletion = 0x0
	TypeValue    = 0x1
)

type DB struct {
	mem     *MemTable
	imm     *MemTable
	log     *log.Writer
	logfile *os.File
	opt     *Options
}

func (d *DB) Write(b *Batch, opt *WriteOptions) error {

	d.logfile.Sync()
	//The write happen synchronously.
	// Writer w(&mutex_);
	// w.batch = my_batch;
	// w.sync = options.sync;
	// w.done = false;

	// MutexLock l(&mutex_);
	// writers_.push_back(&w);
	// while (!w.done && &w != writers_.front()) {
	//   w.cv.Wait();
	// }
	// if (w.done) {
	//   return w.status;
	// }

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

}
