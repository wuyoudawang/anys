package log

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/liuzhiyi/anys/pkg/utils"
)

type Dropper interface {
	Drop(int, string)
}

type Reader struct {
	r                io.Reader
	checkSum         bool
	rOffset, wOffset int
	eof              bool
	buf              [blockSize]byte
	dropper          Dropper
}

func NewReader(r io.Reader, dropper Dropper, checksum bool) *Reader {
	return &Reader{
		r:        r,
		dropper:  dropper,
		checkSum: checksum,
		eof:      false,
	}
}

func (r *Reader) corrupt(n int, reason string, skip bool) error {
	if r.dropper != nil {
		r.dropper.Drop(n, reason)
	}
	return fmt.Errorf("leveldb-distribution/log: block/chunk corrupted: %s (%d bytes)", reason, n)
}

func (r *Reader) skipTail() {

}

func (r *Reader) ReadRecord(b *[]byte) error {
	in_fragmented_record := false
	for {
		var fragment []byte
		typ := r.readPhysicalRecord(&fragment)
		switch typ {
		case KFullType:
			if in_fragmented_record {
				if len(*b) == 0 {
					in_fragmented_record = false
				} else {
					r.corrupt(len(*b), "partial record without end(1)", false)
				}
			}
			*b = fragment
			return nil
		case KFirstType:
			if in_fragmented_record {
				if len(*b) == 0 {
					in_fragmented_record = false
				} else {
					r.corrupt(len(*b), "partial record without end(2)", false)
				}
			}
			*b = append(*b, fragment...)
			in_fragmented_record = true
		case KMiddleType:
			if !in_fragmented_record {
				r.corrupt(len(fragment), "missing start of fragmented record(1)", false)
			} else {
				*b = append(*b, fragment...)
			}
		case KLastType:
			if !in_fragmented_record {
				r.corrupt(len(fragment), "missing start of fragmented record(2)", false)
			} else {
				*b = append(*b, fragment...)
				return nil
			}
		case KEof:
			return nil
		case KBadType:
			if in_fragmented_record {
				r.corrupt(len(*b), "error in middle of record", false)
				in_fragmented_record = false
				*b = (*b)[:0]
			}
		default:
			reason := fmt.Sprintf("unknown record type %u", typ)
			size := len(fragment)
			if in_fragmented_record {
				size += len(*b)
			}
			r.corrupt(size, reason, false)
			in_fragmented_record = false
			*b = (*b)[:0]
		}
	}
}

func (r *Reader) bufSize() int {
	return r.wOffset - r.rOffset
}

func (r *Reader) readPhysicalRecord(fragment *[]byte) int {
	for {
		if r.bufSize() < headerSize {
			if !r.eof {
				r.rOffset = 0
				r.wOffset = 0
				n, err := io.ReadFull(r.r, r.buf[:])
				if err != nil && err != io.ErrUnexpectedEOF {
					r.corrupt(n, err.Error(), false)
					r.eof = true
					return KEof
				}

				if n < blockSize {
					r.eof = true
				}
				r.wOffset = n
				continue
			} else {
				r.rOffset = 0
				r.wOffset = 0
				return KEof
			}
		}

		// parse the header
		checksum := binary.LittleEndian.Uint32(r.buf[r.rOffset : r.rOffset+4])
		length := binary.LittleEndian.Uint16(r.buf[r.rOffset+4 : r.rOffset+6])
		chunkType := r.buf[r.rOffset+6]

		if headerSize+int(length) > r.bufSize() {
			n := r.bufSize()
			r.rOffset = 0
			r.wOffset = 0
			if !r.eof {
				r.corrupt(n, "bad record length", false)
				return KBadType
			}
			return KEof
		}

		if chunkType == KZeroType && length == 0 {
			r.rOffset = 0
			r.wOffset = 0
			return KBadType
		}

		data := r.buf[r.rOffset+headerSize : r.rOffset+headerSize+int(length)]
		if r.checkSum {
			if utils.NewCRC(data).Value() != checksum {
				n := r.bufSize()
				r.wOffset = 0
				r.rOffset = 0
				r.corrupt(n, "checksum mismatch", false)
				return KBadType
			}
		}

		r.rOffset += headerSize + int(length)

		*fragment = append(*fragment, data...)
		return int(chunkType)

	}
}
