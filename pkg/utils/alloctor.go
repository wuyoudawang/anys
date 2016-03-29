package utils

// This a copy of Go std bytes.Buffer with some modification
// and some features stripped.

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Pool struct {
	mem []byte
}

type Alloctor struct {
	buf []byte
}

func (a *Alloctor) New() unsafe.Pointer {
	return nil
}

// A Buffer is a variable-sized buffer of bytes with Read and Write methods.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf       []byte   // contents are the bytes buf[off : len(buf)]
	off       int      // read at &buf[off], write at &buf[len(buf)]
	bootstrap [64]byte // memory to hold first slice; helps small buffers (Printf) avoid allocation.
}

// Bytes returns a slice of the contents of the unread portion of the buffer;
// len(b.Bytes()) == b.Len().  If the caller changes the contents of the
// returned slice, the contents of the buffer will change provided there
// are no intervening method calls on the Buffer.
func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// String returns the contents of the unread portion of the buffer
// as a string.  If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *Buffer) Len() int { return len(b.buf) - b.off }

// Truncate discards all but the first n unread bytes from the buffer.
// It panics if n is negative or greater than the length of the buffer.
func (b *Buffer) Truncate(n int) {
	switch {
	case n < 0 || n > b.Len():
		panic("leveldb/util.Buffer: truncation out of range")
	case n == 0:
		// Reuse buffer space.
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+n]
}

// Reset resets the buffer so it has no content.
// b.Reset() is the same as b.Truncate(0).
func (b *Buffer) Reset() { b.Truncate(0) }

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with bytes.ErrTooLarge.
func (b *Buffer) grow(n int) int {
	m := b.Len()
	// If buffer is empty, reset to recover space.
	if m == 0 && b.off != 0 {
		b.Truncate(0)
	}
	if len(b.buf)+n > cap(b.buf) {
		var buf []byte
		if b.buf == nil && n <= len(b.bootstrap) {
			buf = b.bootstrap[0:]
		} else if m+n <= cap(b.buf)/2 {
			// We can slide things down instead of allocating a new
			// slice. We only need m+n <= cap(b.buf) to slide, but
			// we instead let capacity get twice as large so we
			// don't spend all our time copying.
			copy(b.buf[:], b.buf[b.off:])
			buf = b.buf[:m]
		} else {
			// not enough space anywhere
			buf = makeSlice(2*cap(b.buf) + n)
			copy(buf, b.buf[b.off:])
		}
		b.buf = buf
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+m+n]
	return b.off + m
}

// Alloc allocs n bytes of slice from the buffer, growing the buffer as
// needed. If n is negative, Alloc will panic.
// If the buffer can't grow it will panic with bytes.ErrTooLarge.
func (b *Buffer) Alloc(n int) []byte {
	if n < 0 {
		panic("leveldb/util.Buffer.Alloc: negative count")
	}
	m := b.grow(n)
	return b.buf[m:]
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocation.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with bytes.ErrTooLarge.
func (b *Buffer) Grow(n int) {
	if n < 0 {
		panic("leveldb/util.Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[0:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with bytes.ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	m := b.grow(len(p))
	return copy(b.buf[m:], p), nil
}

// MinRead is the minimum slice size passed to a Read call by
// Buffer.ReadFrom.  As long as the Buffer has at least MinRead bytes beyond
// what is required to hold the contents of r, ReadFrom will not grow the
// underlying buffer.
const MinRead = 512

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with bytes.ErrTooLarge.
func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error) {
	// If buffer is empty, reset to recover space.
	if b.off >= len(b.buf) {
		b.Truncate(0)
	}
	for {
		if free := cap(b.buf) - len(b.buf); free < MinRead {
			// not enough space at end
			newBuf := b.buf
			if b.off+free < MinRead {
				// not enough space using beginning of buffer;
				// double buffer capacity
				newBuf = makeSlice(2*cap(b.buf) + MinRead)
			}
			copy(newBuf, b.buf[b.off:])
			b.buf = newBuf[:len(b.buf)-b.off]
			b.off = 0
		}
		m, e := r.Read(b.buf[len(b.buf):cap(b.buf)])
		b.buf = b.buf[0 : len(b.buf)+m]
		n += int64(m)
		if e == io.EOF {
			break
		}
		if e != nil {
			return n, e
		}
	}
	return n, nil // err is EOF, so return nil explicitly
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with bytes.ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(bytes.ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// WriteTo writes data to w until the buffer is drained or an error occurs.
// The return value n is the number of bytes written; it always fits into an
// int, but it is int64 to match the io.WriterTo interface. Any error
// encountered during the write is also returned.
func (b *Buffer) WriteTo(w io.Writer) (n int64, err error) {
	if b.off < len(b.buf) {
		nBytes := b.Len()
		m, e := w.Write(b.buf[b.off:])
		if m > nBytes {
			panic("leveldb/util.Buffer.WriteTo: invalid Write count")
		}
		b.off += m
		n = int64(m)
		if e != nil {
			return n, e
		}
		// all bytes should have been written, by definition of
		// Write method in io.Writer
		if m != nBytes {
			return n, io.ErrShortWrite
		}
	}
	// Buffer is now empty; reset.
	b.Truncate(0)
	return
}

// WriteByte appends the byte c to the buffer, growing the buffer as needed.
// The returned error is always nil, but is included to match bufio.Writer's
// WriteByte. If the buffer becomes too large, WriteByte will panic with
// bytes.ErrTooLarge.
func (b *Buffer) WriteByte(c byte) error {
	m := b.grow(1)
	b.buf[m] = c
	return nil
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained.  The return value n is the number of bytes read.  If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (n int, err error) {
	if b.off >= len(b.buf) {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}
	n = copy(p, b.buf[b.off:])
	b.off += n
	return
}

// Next returns a slice containing the next n bytes from the buffer,
// advancing the buffer as if the bytes had been returned by Read.
// If there are fewer than n bytes in the buffer, Next returns the entire buffer.
// The slice is only valid until the next call to a read or write method.
func (b *Buffer) Next(n int) []byte {
	m := b.Len()
	if n > m {
		n = m
	}
	data := b.buf[b.off : b.off+n]
	b.off += n
	return data
}

// ReadByte reads and returns the next byte from the buffer.
// If no byte is available, it returns error io.EOF.
func (b *Buffer) ReadByte() (c byte, err error) {
	if b.off >= len(b.buf) {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		return 0, io.EOF
	}
	c = b.buf[b.off]
	b.off++
	return c, nil
}

// ReadBytes reads until the first occurrence of delim in the input,
// returning a slice containing the data up to and including the delimiter.
// If ReadBytes encounters an error before finding a delimiter,
// it returns the data read before the error and the error itself (often io.EOF).
// ReadBytes returns err != nil if and only if the returned data does not end in
// delim.
func (b *Buffer) ReadBytes(delim byte) (line []byte, err error) {
	slice, err := b.readSlice(delim)
	// return a copy of slice. The buffer's backing array may
	// be overwritten by later calls.
	line = append(line, slice...)
	return
}

// readSlice is like ReadBytes but returns a reference to internal buffer data.
func (b *Buffer) readSlice(delim byte) (line []byte, err error) {
	i := bytes.IndexByte(b.buf[b.off:], delim)
	end := b.off + i + 1
	if i < 0 {
		end = len(b.buf)
		err = io.EOF
	}
	line = b.buf[b.off:end]
	b.off = end
	return line, err
}

// NewBuffer creates and initializes a new Buffer using buf as its initial
// contents.  It is intended to prepare a Buffer to read existing data.  It
// can also be used to size the internal buffer for writing. To do that,
// buf should have the desired capacity but a length of zero.
//
// In most cases, new(Buffer) (or just declaring a Buffer variable) is
// sufficient to initialize a Buffer.
func NewBuffer(buf []byte) *Buffer { return &Buffer{buf: buf} }

type buffer struct {
	b    []byte
	miss int
}

// BufferPool is a 'buffer pool'.
type BufferPool struct {
	pool      [6]chan []byte
	size      [5]uint32
	sizeMiss  [5]uint32
	sizeHalf  [5]uint32
	baseline  [4]int
	baseline0 int

	mu     sync.RWMutex
	closed bool
	closeC chan struct{}

	get     uint32
	put     uint32
	half    uint32
	less    uint32
	equal   uint32
	greater uint32
	miss    uint32
}

func (p *BufferPool) poolNum(n int) int {
	if n <= p.baseline0 && n > p.baseline0/2 {
		return 0
	}
	for i, x := range p.baseline {
		if n <= x {
			return i + 1
		}
	}
	return len(p.baseline) + 1
}

// Get returns buffer with length of n.
func (p *BufferPool) Get(n int) []byte {
	if p == nil {
		return make([]byte, n)
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return make([]byte, n)
	}

	atomic.AddUint32(&p.get, 1)

	poolNum := p.poolNum(n)
	pool := p.pool[poolNum]
	if poolNum == 0 {
		// Fast path.
		select {
		case b := <-pool:
			switch {
			case cap(b) > n:
				if cap(b)-n >= n {
					atomic.AddUint32(&p.half, 1)
					select {
					case pool <- b:
					default:
					}
					return make([]byte, n)
				} else {
					atomic.AddUint32(&p.less, 1)
					return b[:n]
				}
			case cap(b) == n:
				atomic.AddUint32(&p.equal, 1)
				return b[:n]
			default:
				atomic.AddUint32(&p.greater, 1)
			}
		default:
			atomic.AddUint32(&p.miss, 1)
		}

		return make([]byte, n, p.baseline0)
	} else {
		sizePtr := &p.size[poolNum-1]

		select {
		case b := <-pool:
			switch {
			case cap(b) > n:
				if cap(b)-n >= n {
					atomic.AddUint32(&p.half, 1)
					sizeHalfPtr := &p.sizeHalf[poolNum-1]
					if atomic.AddUint32(sizeHalfPtr, 1) == 20 {
						atomic.StoreUint32(sizePtr, uint32(cap(b)/2))
						atomic.StoreUint32(sizeHalfPtr, 0)
					} else {
						select {
						case pool <- b:
						default:
						}
					}
					return make([]byte, n)
				} else {
					atomic.AddUint32(&p.less, 1)
					return b[:n]
				}
			case cap(b) == n:
				atomic.AddUint32(&p.equal, 1)
				return b[:n]
			default:
				atomic.AddUint32(&p.greater, 1)
				if uint32(cap(b)) >= atomic.LoadUint32(sizePtr) {
					select {
					case pool <- b:
					default:
					}
				}
			}
		default:
			atomic.AddUint32(&p.miss, 1)
		}

		if size := atomic.LoadUint32(sizePtr); uint32(n) > size {
			if size == 0 {
				atomic.CompareAndSwapUint32(sizePtr, 0, uint32(n))
			} else {
				sizeMissPtr := &p.sizeMiss[poolNum-1]
				if atomic.AddUint32(sizeMissPtr, 1) == 20 {
					atomic.StoreUint32(sizePtr, uint32(n))
					atomic.StoreUint32(sizeMissPtr, 0)
				}
			}
			return make([]byte, n)
		} else {
			return make([]byte, n, size)
		}
	}
}

// Put adds given buffer to the pool.
func (p *BufferPool) Put(b []byte) {
	if p == nil {
		return
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return
	}

	atomic.AddUint32(&p.put, 1)

	pool := p.pool[p.poolNum(cap(b))]
	select {
	case pool <- b:
	default:
	}

}

func (p *BufferPool) Close() {
	if p == nil {
		return
	}

	p.mu.Lock()
	if !p.closed {
		p.closed = true
		p.closeC <- struct{}{}
	}
	p.mu.Unlock()
}

func (p *BufferPool) String() string {
	if p == nil {
		return "<nil>"
	}

	return fmt.Sprintf("BufferPool{B·%d Z·%v Zm·%v Zh·%v G·%d P·%d H·%d <·%d =·%d >·%d M·%d}",
		p.baseline0, p.size, p.sizeMiss, p.sizeHalf, p.get, p.put, p.half, p.less, p.equal, p.greater, p.miss)
}

func (p *BufferPool) drain() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			for _, ch := range p.pool {
				select {
				case <-ch:
				default:
				}
			}
		case <-p.closeC:
			close(p.closeC)
			for _, ch := range p.pool {
				close(ch)
			}
			return
		}
	}
}

// NewBufferPool creates a new initialized 'buffer pool'.
func NewBufferPool(baseline int) *BufferPool {
	if baseline <= 0 {
		panic("baseline can't be <= 0")
	}
	p := &BufferPool{
		baseline0: baseline,
		baseline:  [...]int{baseline / 4, baseline / 2, baseline * 2, baseline * 4},
		closeC:    make(chan struct{}, 1),
	}
	for i, cap := range []int{2, 2, 4, 4, 2, 1} {
		p.pool[i] = make(chan []byte, cap)
	}
	go p.drain()
	return p
}
