package comparator

import "bytes"

type bytesComparator struct{}

func (bytesComparator) Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func (bytesComparator) Name() string {
	return "leveldb.BytewiseComparator"
}

func (bytesComparator) Separator(dst, a, b []byte) []byte {
	i, n := 0, len(a)
	if n > len(b) {
		n = len(b)
	}
	for ; i < n && a[i] == b[i]; i++ {
	}
	if i >= n {
		// Do not shorten if one string is a prefix of the other
	} else if c := a[i]; c < 0xff && c+1 < b[i] {
		dst = append(dst, a[:i+1]...)
		dst[i]++
		return dst
	}
	return nil
}

func (bytesComparator) Successor(dst, b []byte) []byte {
	for i, c := range b {
		if c != 0xff {
			dst = append(dst, b[:i+1]...)
			dst[i]++
			return dst
		}
	}
	return nil
}

// DefaultComparer are default implementation of the Comparer interface.
// It uses the natural ordering, consistent with bytes.Compare.
var DefaultComparer = bytesComparator{}
