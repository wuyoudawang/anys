package utils

func Key(name string) int64 {
	var k int64
	for i, v := range name {
		k += int64(i) * int64(v)
	}

	return k
}
