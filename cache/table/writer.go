package table

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/golang/snappy"

	"anys/cache/filter"
	"anys/cache/option"
	"anys/pkg/comparator"
	"anys/pkg/utils"
)

func sharedPrefixLen(a, b []byte) int {
	i, n := 0, len(a)
	if n > len(b) {
		n = len(b)
	}
	for i < n && a[i] == b[i] {
		i++
	}
	return i
}

type blockWriter struct {
	restartInterval int
	buf             utils.Buffer
	nEntries        int
	prevKey         []byte
	restarts        []uint32
	scratch         []byte
}

func (bw *blockWriter) append(key, value []byte) {
	nShared := 0
	if bw.nEntries%bw.restartInterval == 0 {
		bw.restarts = append(bw.restarts, uint32(bw.buf.Len()))
	} else {
		nShared = sharedPrefixLen(bw.prevKey, key)
	}
	n := binary.PutUvarint(bw.scratch[0:], uint64(nShared))
	n += binary.PutUvarint(bw.scratch[n:], uint64(len(key)-nShared))
	n += binary.PutUvarint(bw.scratch[n:], uint64(len(value)))
	bw.buf.Write(bw.scratch[:n])
	bw.buf.Write(key[nShared:])
	bw.buf.Write(value)
	bw.prevKey = append(bw.prevKey[:0], key...)
	bw.nEntries++
}

func (bw *blockWriter) finish() {
	// Write restarts entry.
	if bw.nEntries == 0 {
		// Must have at least one restart entry.
		bw.restarts = append(bw.restarts, 0)
	}
	bw.restarts = append(bw.restarts, uint32(len(bw.restarts)))
	for _, x := range bw.restarts {
		buf4 := bw.buf.Alloc(4)
		binary.LittleEndian.PutUint32(buf4, x)
	}
}

func (bw *blockWriter) reset() {
	bw.buf.Reset()
	bw.nEntries = 0
	bw.restarts = bw.restarts[:0]
}

func (bw *blockWriter) bytesLen() int {
	restartsLen := len(bw.restarts)
	if restartsLen == 0 {
		restartsLen = 1
	}
	return bw.buf.Len() + 4*restartsLen + 4
}

type filterWriter struct {
	generator filter.FilterGenerator
	buf       utils.Buffer
	nKeys     int
	offsets   []uint32
}

func (fw *filterWriter) add(key []byte) {
	if fw.generator == nil {
		return
	}
	fw.generator.Add(key)
	fw.nKeys++
}

func (fw *filterWriter) flush(offset uint64) {
	if fw.generator == nil {
		return
	}
	for x := int(offset / filterBase); x > len(fw.offsets); {
		fw.generate()
	}
}

func (fw *filterWriter) finish() {
	if fw.generator == nil {
		return
	}
	// Generate last keys.

	if fw.nKeys > 0 {
		fw.generate()
	}
	fw.offsets = append(fw.offsets, uint32(fw.buf.Len()))
	for _, x := range fw.offsets {
		buf4 := fw.buf.Alloc(4)
		binary.LittleEndian.PutUint32(buf4, x)
	}
	fw.buf.WriteByte(filterBaseLg)
}

func (fw *filterWriter) generate() {
	// Record offset.
	fw.offsets = append(fw.offsets, uint32(fw.buf.Len()))
	// Generate filters.
	if fw.nKeys > 0 {
		fw.generator.Generate(&fw.buf)
		fw.nKeys = 0
	}
}

type Writer struct {
	writer io.Writer
	err    error
	// Options
	cmp             comparator.Comparator
	filter          filter.Filter
	compressionType int
	blockSize       int

	dataBlock   blockWriter
	indexBlock  blockWriter
	filterBlock filterWriter
	pendingBH   blockHandle
	offset      uint64
	nEntries    int
	// Scratch allocated enough for 5 uvarint. Block writer should not use
	// first 20-bytes since it will be used to encode block handle, which
	// then passed to the block writer itself.
	scratch            [50]byte
	comparerScratch    []byte
	compressionScratch []byte
}

func (w *Writer) writeBlock(buf *utils.Buffer, cprt int) (bh blockHandle, err error) {
	// Compress the buffer if necessary.
	var b []byte
	if cprt == option.KSnappyCompression {
		// Allocate scratch enough for compression and block trailer.
		if n := snappy.MaxEncodedLen(buf.Len()) + blockTrailerLen; len(w.compressionScratch) < n {
			w.compressionScratch = make([]byte, n)
		}
		compressed := snappy.Encode(w.compressionScratch, buf.Bytes())
		n := len(compressed)
		b = compressed[:n+blockTrailerLen]
		b[n] = option.KSnappyCompression
	} else {
		tmp := buf.Alloc(blockTrailerLen)
		tmp[0] = option.KNoCompression
		b = buf.Bytes()
	}

	// Calculate the checksum.
	n := len(b) - 4
	checksum := utils.NewCRC(b[:n]).Value()
	binary.LittleEndian.PutUint32(b[n:], checksum)

	// Write the buffer to the file.
	_, err = w.writer.Write(b)
	if err != nil {
		return
	}
	bh = blockHandle{w.offset, uint64(len(b) - blockTrailerLen)}
	w.offset += uint64(len(b))
	return
}

func (w *Writer) flushPendingBH(key []byte) {
	if w.pendingBH.length == 0 {
		return
	}
	var separator []byte
	if len(key) == 0 {
		separator = w.cmp.Successor(w.comparerScratch[:0], w.dataBlock.prevKey)
	} else {
		separator = w.cmp.Separator(w.comparerScratch[:0], w.dataBlock.prevKey, key)
	}
	if separator == nil {
		separator = w.dataBlock.prevKey
	} else {
		w.comparerScratch = separator
	}
	n := encodeBlockHandle(w.scratch[:20], w.pendingBH)
	// Append the block handle to the index block.
	w.indexBlock.append(separator, w.scratch[:n])
	// Reset prev key of the data block.
	w.dataBlock.prevKey = w.dataBlock.prevKey[:0]
	// Clear pending block handle.
	w.pendingBH = blockHandle{}
}

func (w *Writer) finishBlock() error {
	w.dataBlock.finish()
	bh, err := w.writeBlock(&w.dataBlock.buf, w.compressionType)
	if err != nil {
		return err
	}
	w.pendingBH = bh
	// Reset the data block.
	w.dataBlock.reset()
	// Flush the filter block.
	w.filterBlock.flush(w.offset)
	return nil
}

// Append appends key/value pair to the table. The keys passed must
// be in increasing order.
//
// It is safe to modify the contents of the arguments after Append returns.
func (w *Writer) Append(key, value []byte) error {
	if w.err != nil {
		return w.err
	}
	if w.nEntries > 0 && w.cmp.Compare(w.dataBlock.prevKey, key) >= 0 {
		w.err = fmt.Errorf("leveldb/table: Writer: keys are not in increasing order: %q, %q", w.dataBlock.prevKey, key)
		return w.err
	}

	w.flushPendingBH(key)
	// Append key/value pair to the data block.
	w.dataBlock.append(key, value)
	// Add key to the filter block.
	w.filterBlock.add(key)

	// Finish the data block if block size target reached.
	if w.dataBlock.bytesLen() >= w.blockSize {
		if err := w.finishBlock(); err != nil {
			w.err = err
			return w.err
		}
	}
	w.nEntries++
	return nil
}

// BlocksLen returns number of blocks written so far.
func (w *Writer) BlocksLen() int {
	n := w.indexBlock.nEntries
	if w.pendingBH.length > 0 {
		// Includes the pending block.
		n++
	}
	return n
}

// EntriesLen returns number of entries added so far.
func (w *Writer) EntriesLen() int {
	return w.nEntries
}

// BytesLen returns number of bytes written so far.
func (w *Writer) BytesLen() int {
	return int(w.offset)
}

// Close will finalize the table. Calling Append is not possible
// after Close, but calling BlocksLen, EntriesLen and BytesLen
// is still possible.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}

	// Write the last data block. Or empty data block if there
	// aren't any data blocks at all.
	if w.dataBlock.nEntries > 0 || w.nEntries == 0 {
		if err := w.finishBlock(); err != nil {
			w.err = err
			return w.err
		}
	}
	w.flushPendingBH(nil)

	// Write the filter block.
	var filterBH blockHandle
	w.filterBlock.finish()
	if buf := &w.filterBlock.buf; buf.Len() > 0 {
		filterBH, w.err = w.writeBlock(buf, option.KNoCompression)
		if w.err != nil {
			return w.err
		}
	}

	// Write the metaindex block.
	if filterBH.length > 0 {
		key := []byte("filter." + w.filter.Name())
		n := encodeBlockHandle(w.scratch[:20], filterBH)
		w.dataBlock.append(key, w.scratch[:n])
	}
	w.dataBlock.finish()
	metaindexBH, err := w.writeBlock(&w.dataBlock.buf, w.compressionType)
	if err != nil {
		w.err = err
		return w.err
	}

	// Write the index block.
	w.indexBlock.finish()
	indexBH, err := w.writeBlock(&w.indexBlock.buf, w.compressionType)
	if err != nil {
		w.err = err
		return w.err
	}

	// Write the table footer.
	footer := w.scratch[:footerLen]
	for i := range footer {
		footer[i] = 0
	}
	n := encodeBlockHandle(footer, metaindexBH)
	encodeBlockHandle(footer[n:], indexBH)
	copy(footer[footerLen-len(magic):], magic)
	if _, err := w.writer.Write(footer); err != nil {
		w.err = err
		return w.err
	}
	w.offset += footerLen

	w.err = errors.New("leveldb/table: writer is closed")
	return nil
}

// NewWriter creates a new initialized table writer for the file.
//
// Table writer is not goroutine-safe.
func NewWriter(f io.Writer, opt *option.Options) *Writer {
	w := &Writer{
		writer:          f,
		cmp:             opt.Compare,
		filter:          opt.Filter,
		compressionType: opt.CompressionType,
		blockSize:       opt.BlockSize,
		comparerScratch: make([]byte, 0),
	}
	// data block
	w.dataBlock.restartInterval = opt.BlockRestartInterval
	// The first 20-bytes are used for encoding block handle.
	w.dataBlock.scratch = w.scratch[20:]
	// index block
	w.indexBlock.restartInterval = 1
	w.indexBlock.scratch = w.scratch[20:]
	// filter block
	if w.filter != nil {
		w.filterBlock.generator = w.filter.NewGenerator()
		w.filterBlock.flush(0)
	}
	return w
}
