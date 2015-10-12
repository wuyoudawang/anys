package storeEngine

import (
// "anys/cache/iterator"
)

type memTableIterator struct {
}

type MemTable struct {
}

func (m *MemTable) Put(seq int, valueType int, key []int, value []int) {
	// Format of an entry is concatenation of:
	//  key_size     : varint32 of internal_key.size()
	//  key bytes    : char[internal_key.size()]
	//  value_size   : varint32 of value.size()
	//  value bytes  : char[value.size()]
}
