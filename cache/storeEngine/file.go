package storeEngine

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"anys/pkg/utils"
)

type fileStoreEngine struct {
	dbname string

	mu      sync.Mutex
	flock   fileLock
	slock   *fileStorageLock
	logw    *os.File
	logfile *os.File
	buf     []byte
	// Opened file counter; if open < 0 means closed.
	open int
	day  int
}

func (fse *fileStoreEngine) OpenLogFile(number uint64) error {
	name := logFileName(name, number)
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, "0644")
	if err != nil {
		return err
	}
	fse.logfile = f
	return nil
}

func (d *DB) makeRoomForWrite(force bool) {
	for {

	}
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

func (f *fileStoreEngine) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if fs.open < 0 {
		return ErrClosed
	}
	// Clear the finalizer.
	runtime.SetFinalizer(fs, nil)

	if fs.open > 0 {
		fs.log(fmt.Sprintf("close: warning, %d files still open", fs.open))
	}
	fs.open = -1
	e1 := fs.logw.Close()
	err := fs.flock.release()
	if err == nil {
		err = e1
	}
	return err
}

func (f *fileStoreEngine) printDay() {
	if f.day == t.Day() {
		return
	}
	f.day = t.Day()
	f.logw.Write([]byte("=============== " + t.Format("Jan 2, 2006 (MST)") + " ===============\n"))
}

func (f *fileStoreEngine) log(t time.Time, data string) {
	f.printDay(t)
	hour, min, sec := t.Clock()
	msec := t.Nanosecond() / 1e3
	// time
	f.buf = itoa(fs.buf[:0], hour, 2)
	f.buf = append(fs.buf, ':')
	f.buf = itoa(fs.buf, min, 2)
	f.buf = append(fs.buf, ':')
	f.buf = itoa(fs.buf, sec, 2)
	f.buf = append(fs.buf, '.')
	f.buf = itoa(fs.buf, msec, 6)
	f.buf = append(fs.buf, ' ')
	// write
	f.buf = append(fs.buf, []byte(str)...)
	f.buf = append(fs.buf, '\n')
	f.logw.Write(fs.buf)
}

func (f *fileStoreEngine) Log(data string) {
	t := time.Now()
}
