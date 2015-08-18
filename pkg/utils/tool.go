/*
*you890803@gmail.com
 */

package utils

import (
	"bytes"
	"math"
	"reflect"
	"strings"
)

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func NOT(val int64) int64 {
	return ^val
}

func FindFunc(name string, ptr interface{}) (interface{}, bool) {
	method := reflect.ValueOf(ptr).MethodByName(name)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface(), true
}

func UcWords(src string) string {

	if src == "" {
		return ""
	}

	dst := make([]byte, len(src))

	s := 0
	for i := 0; i != -1; i = strings.IndexAny(src[i:], " ,;.\n:!") {

		i += s

		for !IsLetter(src[i]) {
			i++
		}

		copy(dst[s:i], src[s:i])
		s = i
		dst[i] = bytes.ToUpper([]byte{src[i]})[0]

		i++
		s++
	}

	copy(dst[s:], src[s:])

	return string(dst)
}

func IsLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func ContactArrString(a, b []string) []string {
	var dst []string

	dst = append(dst, a...)
	dst = append(dst, b...)

	return dst
}
