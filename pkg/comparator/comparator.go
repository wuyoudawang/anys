package comparator

type BasicComparator interface {
	Compare(a, b []byte) int
}

type Comparator interface {
	BasicComparator
	Name() string
	Separator(dst, a, b []byte) []byte
	Successor(dst, b []byte) []byte
}
