package utils

import (
	"bytes"
)

func Key(name string) int64 {
	var k int64
	for i, v := range name {
		k += int64(i) * int64(v)
	}

	return k
}

func Tail(buffer *bytes.Buffer, n int) string {
	if n <= 0 {
		return ""
	}

	bytes := buffer.Bytes()
	if len(bytes) > 0 && bytes[len(bytes)-1] == '\n' {
		bytes = bytes[:len(bytes)-1]
	}

	for i := buffer.Len() - 2; i >= 0; i-- {
		if bytes[i] == '\n' {
			n--
			if n == 0 {
				return string(bytes[i+1:])
			}
		}
	}
	return string(bytes)
}
