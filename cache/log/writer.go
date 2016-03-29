package log

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/liuzhiyi/anys/pkg/utils"
)

type flusher interface {
	Flush() error
}

var (
	ErrOverMaxRecordSize = errors.New("size is greater than the max-record-size")
	ErrWriterBufOverflow = errors.New("the log buffer has overflow")
)

type Writer struct {
	w           io.Writer
	blockOffset int
	buf         [blockSize]byte
	f           flusher
}

func NewWriter(w io.Writer) *Writer {
	lw := new(Writer)
	lw.w = w
	if f, ok := w.(flusher); ok {
		lw.f = f
	}

	return lw
}

func (w *Writer) write() error {
	written := 0
	for {
		n, err := w.w.Write(w.buf[written:w.blockOffset])
		if err != nil {
			return err
		}
		written += n
		if written == w.blockOffset {
			w.blockOffset = 0
			break
		}
	}

	return nil
}

func (w *Writer) Flush() error {
	var err error
	if w.blockOffset > 0 {
		err = w.write()
		w.blockOffset = 0
	}
	if w.f != nil {
		return w.f.Flush()
	}

	return err
}

func (w *Writer) AddRecord(b []byte) error {
	left := len(b)
	off := 0
	var err error
	begin := true
	for left > 0 && err == nil {
		leftOver := blockSize - w.blockOffset
		if leftOver < 0 {
			return ErrWriterBufOverflow
		}

		if leftOver < headerSize {
			if leftOver > 0 {
				for w.blockOffset < blockSize {
					w.buf[w.blockOffset] = 0x00
					w.blockOffset++
				}
			}
			err = w.write()
			w.blockOffset = 0
		}

		avail := blockSize - w.blockOffset - headerSize
		fragmentLen := avail
		if left < avail {
			fragmentLen = left
		}

		end := (left == fragmentLen)
		typ := KZeroType
		if begin && end {
			typ = KFullType
		} else if begin {
			typ = KFirstType
		} else if end {
			typ = KLastType
		} else {
			typ = KMiddleType
		}
		if string(b) == "99141." {
			fmt.Println(typ, fragmentLen, string(b[off:off+fragmentLen]))
		}
		err = w.addPhysicalRecord(typ, b[off:off+fragmentLen])
		off += fragmentLen
		left -= fragmentLen
		begin = false
	}

	return err
}

func (w *Writer) addPhysicalRecord(typ int, b []byte) error {
	if len(b) > 0xffff {
		return ErrOverMaxRecordSize
	}

	if len(b)+headerSize+w.blockOffset > blockSize {
		return ErrWriterBufOverflow
	}

	w.buf[w.blockOffset+6] = byte(typ)
	binary.LittleEndian.PutUint32(w.buf[w.blockOffset+0:w.blockOffset+4], utils.NewCRC(b).Value())
	binary.LittleEndian.PutUint16(w.buf[w.blockOffset+4:w.blockOffset+6], uint16(len(b)))

	copy(w.buf[w.blockOffset+headerSize:w.blockOffset+headerSize+len(b)], b)
	w.blockOffset += headerSize + len(b)

	return nil
}
