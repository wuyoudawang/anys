package cache

import (
	"github.com/liuzhiyi/anys/pkg/comparator"
)

type iComparator struct {
	cmp comparator.Comparator
}

func (icmp *iComparator) Name() string {
	return "leveldb.InternalKeyComparator"
}

func (icmp *iComparator) Compare(a, b []byte) int {
	r := icmp.cmp.Compare(internalKey(a).userKey(), internalKey(b).userKey())
	if r == 0 {
		if m, n := internalKey(a).num(), internalKey(b).num(); m < n {
			r = -1
		} else if m > n {
			r = +1
		}
	}
	return r
}

func (icmp *iComparator) Separator(dst, a, b []byte) []byte {
	ua, ub := iKey(a).ukey(), iKey(b).ukey()
	dst = icmp.cmp.Separator(dst, ua, ub)
	if dst == nil {
		return nil
	}
	if len(dst) < len(ua) && icmp.cmp.Compare(ua, dst) < 0 {
		dst = append(dst, kMaxNumBytes...)
	} else {
		// Did not close possibilities that n maybe longer than len(ub).
		dst = append(dst, a[len(a)-8:]...)
	}
	return dst
}

func (icmp *iComparator) Successor(dst, b []byte) []byte {
	ub := iKey(b).ukey()
	dst = icmp.cmp.Successor(dst, ub)
	if dst == nil {
		return nil
	}
	if len(dst) < len(ub) && icmp.cmp.Compare(ub, dst) < 0 {
		dst = append(dst, kMaxNumBytes...)
	} else {
		// Did not close possibilities that n maybe longer than len(ub).
		dst = append(dst, b[len(b)-8:]...)
	}
	return dst
}
