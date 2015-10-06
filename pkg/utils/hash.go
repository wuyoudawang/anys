package utils

import (
	"encoding/binary"
)

func Hash(data []byte, seed uint32) uint32 {
	var (
		m uint32 = 0xc6a4a793
		r uint32 = 24
		h uint32 = seed ^ (len(data) * m)
	)

	i := 4
	for ; i < len(data); i += 4 {
		w := binary.LittleEndian.Uint32(data)
		h += w
		h *= m
		h ^= (h >> 16)
	}

	switch len(data) - i {
	case 3:
		h += uint8(data[len(data)-1]) << 16
	case 2:
		h += uint8(data[len(data)-1]) << 8
	case 1:
		h += uint8(data[len(data)-1])
		h *= m
		h ^= (h >> r)
	}

	return h
}
