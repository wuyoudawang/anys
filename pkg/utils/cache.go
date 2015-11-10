package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"sync"
)

func Rename(oldname string, newname string) {
	os.Rename(oldname, newname)
}

func Itoa(buf []byte, i int, wid int) []byte {
	var u uint = uint(i)
	if u == 0 && wid <= 1 {
		return append(buf, '0')
	}

	// Assemble decimal in reverse order.
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	return append(buf, b[bp:]...)
}

var table = crc32.MakeTable(crc32.Castagnoli)

// CRC is a CRC-32 checksum computed using Castagnoli's polynomial.
type CRC uint32

// NewCRC creates a new crc based on the given bytes.
func NewCRC(b []byte) CRC {
	return CRC(0).Update(b)
}

// Update updates the crc with the given bytes.
func (c CRC) Update(b []byte) CRC {
	return CRC(crc32.Update(uint32(c), table, b))
}

// Value returns a masked crc.
func (c CRC) Value() uint32 {
	return uint32(c>>15|c<<17) + 0xa282ead8
}

// 用循环双链表实现lru
type LruNode struct {
	value   interface{}
	deleter func([]byte, interface{})
	key     []byte
	prev    *LruNode
	next    *LruNode

	charge    int
	refs      uint32
	hash      uint32
	next_hash *LruNode
}

func (n *LruNode) insert(newNode *LruNode) {
	nt := n.next
	n.next = newNode
	newNode.next = nt

	nt.prev = newNode
	newNode.prev = n
}

func (n *LruNode) delete() {
	n.next.prev = n.prev
	n.prev.next = n.next
}

type LruTable struct {
	length uint32
	elems  uint32
	list   []*LruNode
}

// hash列表，碰撞用链表解决
func NewLruTable() *LruTable {
	lt := new(LruTable)
	lt.length = 0
	lt.elems = 0
	lt.resize()
	return lt
}

func (lt *LruTable) findPtr(key []byte, hash uint32) **LruNode {
	ptr := &(lt.list[hash&(lt.length-1)])
	for *ptr != nil && ((*ptr).hash != hash || bytes.Equal(key, (*ptr).key)) {
		ptr = &((*ptr).next_hash)
	}
	return ptr
}

func (lt *LruTable) resize() {
	var newLen uint32 = 4
	for newLen < lt.elems {
		newLen *= 2
	}

	newList := make([]*LruNode, newLen)
	for i := uint32(0); i < lt.length; i++ {
		node := lt.list[i]
		for node != nil {
			next := node.next_hash
			ptr := &newList[node.hash&(newLen-1)]
			node.next_hash = *ptr
			*ptr = node
			node = next
		}
	}

	lt.list = newList
	lt.length = newLen
}

func (lt *LruTable) Lookup(key []byte, hash uint32) *LruNode {
	return *lt.findPtr(key, hash)
}

func (lt *LruTable) Insert(node *LruNode) *LruNode {
	ptr := lt.findPtr(node.key, node.hash)
	old := *ptr
	if old == nil {
		node.next_hash = nil
	} else {
		node.next_hash = old.next_hash
	}

	*ptr = node
	if old == nil {
		lt.elems++
		if lt.elems > lt.length {
			lt.resize()
		}
	}
	return old
}

func (lt *LruTable) Delete(key []byte, hash uint32) *LruNode {
	ptr := lt.findPtr(key, hash)
	result := *ptr
	if result != nil {
		*ptr = result.next_hash
		lt.elems--
	}
	return result
}

// 线程安全的lru缓存
type LruCache struct {
	m        sync.Mutex
	lru      LruNode
	table    LruTable
	capacity int
	used     int
}

func NewLruCache() *LruCache {
	lc := new(LruCache)
	lc.used = 0
	lc.lru.next = &lc.lru
	lc.lru.prev = &lc.lru
	return lc
}

func (lc *LruCache) unref(node *LruNode) error {
	if node.refs <= 0 {
		return fmt.Errorf("this node ")
	}

	node.refs--
	if node.refs == 0 {
		lc.used -= node.charge
		node.deleter(node.key, node.value)
	}
	return nil
}

func (lc *LruCache) Release(node *LruNode) {
	lc.m.Lock()
	defer lc.m.Unlock()
	lc.unref(node)
}

func (lc *LruCache) Delete(key []byte, hash uint32) {
	lc.m.Lock()
	defer lc.m.Unlock()

	node := lc.table.Delete(key, hash)
	if node != nil {
		node.delete()
		lc.unref(node)
	}
}

func (lc *LruCache) Insert(key []byte, value interface{}, h uint32, charge int, deleter func([]byte, interface{})) *LruNode {
	lc.m.Lock()
	defer lc.m.Unlock()

	node := new(LruNode)
	node.key = key
	node.value = value
	node.hash = h
	node.deleter = deleter
	node.charge = charge
	node.refs = 2 // one from LruCache, one for the returned node
	node.insert(&lc.lru)
	lc.used += charge

	old := lc.table.Insert(node)
	if old != nil {
		old.delete()
		lc.unref(node)
	}

	for lc.used > lc.capacity && lc.lru.next != &lc.lru {
		old := lc.lru.next
		old.delete()
		lc.table.Delete(old.key, old.hash)
		lc.unref(node)
	}

	return node
}

func (lc *LruCache) Lookup(key []byte, hash uint32) *LruNode {
	lc.m.Lock()
	defer lc.m.Unlock()

	node := lc.table.Lookup(key, hash)
	if node != nil {
		node.refs++
		node.delete()
		lc.lru.prev.insert(node)
	}

	return node
}

func (lc *LruCache) Erase(key []byte, hash uint32) {
	lc.m.Lock()
	e := lc.table.Delete(key, hash)
	if e != nil {
		e.delete()
		lc.unref(e)
	}
}

const (
	kNumShardBits = 4
	kNumShards    = 1 << kNumShardBits
)

// 提高并发，减少互斥锁
type ShardedLruCache struct {
	shardC [kNumShards]LruCache
	lastId uint64
	mu     sync.Mutex
}

func (*ShardedLruCache) hash(src []byte) uint32 {
	return Hash(src, 0)
}

func (*ShardedLruCache) shard(h uint32) uint32 {
	return h >> (32 - kNumShardBits)
}

func (slc *ShardedLruCache) Insert(key []byte, value interface{}, charge int, deleter func([]byte, interface{})) *LruNode {
	h := slc.hash(key)
	return slc.shardC[slc.shard(h)].Insert(key, value, h, charge, deleter)
}

func (slc *ShardedLruCache) Lookup(key []byte) *LruNode {
	h := slc.hash(key)
	return slc.shardC[slc.shard(h)].Lookup(key, h)
}

func (slc *ShardedLruCache) Release(node *LruNode) {
	slc.shardC[slc.shard(node.hash)].Release(node)
}

func (slc *ShardedLruCache) Erase(key []byte) {
	h := slc.hash(key)
	slc.shardC[slc.shard(h)].Erase(key, h)
}

func (slc *ShardedLruCache) Value(node *LruNode) interface{} {
	return node.value
}

func PutLenPrefixedBytes(dst, src *[]byte) int {
	encodeLen := VarintLength(uint64(len(*src)))
	*dst = make([]byte, encodeLen)
	binary.PutVarint(*dst, int64(len(*src)))
	*dst = append(*dst, *src...)
	return encodeLen + len(*src)
}

func GetLenPrefixedBytes(dst, src *[]byte) bool {
	val, offset := binary.Varint(*src)
	if val > 0 {
		*src = (*src)[offset:]
		*dst = append(*dst, *src...)
		return true
	} else {
		return false
	}
}

func VarintLength(v uint64) int {
	l := 1
	for v >= 128 {
		v >>= 7
		l++
	}
	return l
}
