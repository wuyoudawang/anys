package storeEngine

import (
	"errors"
	"os"
	"sync"

	// "anys/pkg/utils"
)

var (
	ErrClosed = errors.New("leveldb/storage: closed")
)

type fileLock struct{}

type fileStorageLock struct{}

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
	name := logFileName(fse.dbname, number)
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	fse.logfile = f
	return nil
}

func (fse *fileStoreEngine) Flush() error {
	return nil
}

// Status BuildTable(const std::string& dbname,
//                   Env* env,
//                   const Options& options,
//                   TableCache* table_cache,
//                   Iterator* iter,
//                   FileMetaData* meta) {
//   Status s;
//   meta->file_size = 0;
//   iter->SeekToFirst();

//   std::string fname = TableFileName(dbname, meta->number);
//   if (iter->Valid()) {
//     WritableFile* file;
//     s = env->NewWritableFile(fname, &file);
//     if (!s.ok()) {
//       return s;
//     }

//     TableBuilder* builder = new TableBuilder(options, file);
//     meta->smallest.DecodeFrom(iter->key());
//     for (; iter->Valid(); iter->Next()) {
//       Slice key = iter->key();
//       meta->largest.DecodeFrom(key);
//       builder->Add(key, iter->value());
//     }

//     // Finish and check for builder errors
//     if (s.ok()) {
//       s = builder->Finish();
//       if (s.ok()) {
//         meta->file_size = builder->FileSize();
//         assert(meta->file_size > 0);
//       }
//     } else {
//       builder->Abandon();
//     }
//     delete builder;

//     // Finish and check for file errors
//     if (s.ok()) {
//       s = file->Sync();
//     }
//     if (s.ok()) {
//       s = file->Close();
//     }
//     delete file;
//     file = NULL;

//     if (s.ok()) {
//       // Verify that the table is usable
//       Iterator* it = table_cache->NewIterator(ReadOptions(),
//                                               meta->number,
//                                               meta->file_size);
//       s = it->status();
//       delete it;
//     }
//   }

//   // Check for input iterator errors
//   if (!iter->status().ok()) {
//     s = iter->status();
//   }

//   if (s.ok() && meta->file_size > 0) {
//     // Keep it
//   } else {
//     env->DeleteFile(fname);
//   }
//   return s;
// }
