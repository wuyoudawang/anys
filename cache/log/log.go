package log

const (
	KZeroType = iota
	KFullType

	// For fragments
	KFirstType
	KMiddleType
	KLastType

	// for last type
	KEof
	KBadType
)

const (
	blockSize  = 32 * 1024
	headerSize = 4 + 2 + 1
)
