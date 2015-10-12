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
