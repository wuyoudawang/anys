package cache

import (
	"encoding/binary"
	"os"
	"unsafe"

	"github.com/liuzhiyi/anys/cache/iterator"
	"github.com/liuzhiyi/anys/cache/option"
	"github.com/liuzhiyi/anys/cache/table"
	"github.com/liuzhiyi/anys/pkg/utils"
)

type tableAndFile struct {
	file os.File
	tw   *table.Writer
}

func deleteTableAndFile(key []byte, val interface{}) {
	tf, ok := val.(*tableAndFile)
	if ok {

	}
}

type tableCache struct {
	cache  *utils.ShardedLruCache
	opt    *option.Options
	dbname string
}

func (tc *tableCache) FindTable(fileNumber, fileSize uint64, node **utils.LruNode) error {
	key := make([]byte, unsafe.Sizeof(fileNumber))
	binary.LittleEndian.PutUint64(key, fileNumber)
	*node = tc.cache.Lookup(key)
	if *node == nil {
		fname := storeEngine.TableFileName(tc.dbname, fileNumber)
		fd, err := os.OpenFile(fname, os.O_RDONLY, 0)
		if err != nil {
			oldFname := storeEngine.SstTableFileName(tc.dbname, fileNumber)
			fd, err = os.OpenFile(fname, os.O_RDONLY, 0)
		}

		if err == nil {
			tableWriter := table.NewWriter(fd, tc.opt)
			tf := &tableAndFile{
				file: fd,
				tw:   tableWriter,
			}
			*node = tc.cache.Insert(key, tf, 1, deleteTableAndFile)
		}
		return err
	}
	return nil
}

func (tc *tableCache) NewIterator(opt *option.ReadOptions,
	fileNumber, fileSize uint64,
	k []byte, arg interface{},
	tableptr **table.Writer) {
	if tableptr != nil {
		*tableptr = nil
	}

	var handle *utils.LruNode
	err := tc.FindTable(fileNumber, fileSize, &handle)
	if err != nil {
		return iterator.NewEmptyIterator(err)
	}

	tb := (*tableAndFile)(tc.cache.Value(handle)).tw
	// result :=
	// if (tableptr != NULL) {
	//   *tableptr = NULL;
	// }

	// Cache::Handle* handle = NULL;
	// Status s = FindTable(file_number, file_size, &handle);
	// if (!s.ok()) {
	//   return NewErrorIterator(s);
	// }

	// Table* table = reinterpret_cast<TableAndFile*>(cache_->Value(handle))->table;
	// Iterator* result = table->NewIterator(options);
	// result->RegisterCleanup(&UnrefEntry, cache_, handle);
	// if (tableptr != NULL) {
	//   *tableptr = table;
	// }
	// return result;
}

func (tc *tableCache) Get(opt *option.ReadOptions,
	fileNumber, fileSize uint64,
	k []byte, arg interface{},
	handle_result func(interface{}, []byte, []byte)) {
	// Cache::Handle* handle = NULL;
	// Status s = FindTable(file_number, file_size, &handle);
	// if (s.ok()) {
	//   Table* t = reinterpret_cast<TableAndFile*>(cache_->Value(handle))->table;
	//   s = t->InternalGet(options, k, arg, saver);
	//   cache_->Release(handle);
	// }
	// return s;
}

func (tc *tableCache) Evict(fileNumber uint64) {
	key := make([]byte, unsafe.Sizeof(fileNumber))
	binary.LittleEndian.PutUint64(key, fileNumber)
	tc.cache.Erase(key)
}
