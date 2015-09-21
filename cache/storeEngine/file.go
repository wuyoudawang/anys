package storeEngine

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"anys/pkg/utils"
)

const (
	logFileName   = "LOG"
	logFileSuffix = ".OLD"
)

type fileStoreEngine struct {
	path string

	mu    sync.Mutex
	flock fileLock
	slock *fileStorageLock
	logw  *os.File
	buf   []byte
	// Opened file counter; if open < 0 means closed.
	open int
	day  int
}

func OpenFile(path string) (Interface, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, err
	}

	os.Rename(
		filepath.Join(path, logFileName),
		filepath.Join(path, fmt.Sprintf("%s.%s", logFileName, logFileSuffix)),
	)
	logw, err := os.OpenFile(filepath.Join(path, logFileName), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	f := &fileStoreEngine{path: path, logw: logw}
	runtime.SetFinalizer(f, (*fileStoreEngine).Close)

	return f, nil
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
