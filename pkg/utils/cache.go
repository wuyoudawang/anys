package utils

import (
	"os"
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
