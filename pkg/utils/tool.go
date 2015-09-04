/*
*you890803@gmail.com
 */

package utils

import (
	"bytes"
	"encoding/gob"
	"math"
	"reflect"
	"strconv"
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

func SelectK(arr []int, k int) int {
	if k > len(arr) {
		return -1
	}

	tmp := arr[0]
	j := 0
	pos := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] < tmp {
			arr[j] = arr[pos]
			arr[pos] = arr[i]
			j = i
			pos++
		}
	}
	arr[j] = arr[pos]
	arr[pos] = tmp

	if pos == k-1 {
		return arr[pos]
	} else if pos > k-1 {
		return SelectK(arr[0:pos], k)
	} else {
		return SelectK(arr[pos+1:], k-pos-1)
	}
}

func Serialize(value interface{}) ([]byte, error) {
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	}

	var b bytes.Buffer
	encoder := gob.NewEncoder(&b)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func Unserialize(byt []byte, ptr interface{}) (err error) {
	if bytes, ok := ptr.(*[]byte); ok {
		*bytes = byt
		return
	}

	if v := reflect.ValueOf(ptr); v.Kind() == reflect.Ptr {
		switch p := v.Elem(); p.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var i int64
			i, err = strconv.ParseInt(string(byt), 10, 64)
			if err != nil {

			} else {
				p.SetInt(i)
			}
			return

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var i uint64
			i, err = strconv.ParseUint(string(byt), 10, 64)
			if err != nil {

			} else {
				p.SetUint(i)
			}
			return
		}
	}

	b := bytes.NewBuffer(byt)
	decoder := gob.NewDecoder(b)
	if err = decoder.Decode(ptr); err != nil {

		return
	}
	return
}
