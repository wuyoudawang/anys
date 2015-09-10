/*
*you890803@gmail.com
 */

package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
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

func SubBytes(src []byte, s byte, e byte) ([]byte, error) {
	var dst []byte

	i := bytes.IndexByte(src, s)
	if i == -1 {
		return dst, fmt.Errorf("this sub-string dose not found")
	}

	j := bytes.IndexByte(src[i+1:], e)
	if j == -1 {
		return dst, fmt.Errorf("this sub-string dose not found")
	}

	j += i + 1
	dst = append(dst, src[i+1:j]...)
	return dst, nil
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

type PHPSerializer struct {
	phpArray  byte
	phpString byte
	phpInt    byte
	phpFloat  byte
	phpObject byte
	phpBool   byte
}

func NewDefaultPHPSerializer() *PHPSerializer {
	return &PHPSerializer{
		phpArray:  'a',
		phpString: 's',
		phpInt:    'i',
		phpFloat:  'd',
		phpObject: 'o',
		phpBool:   'b',
	}
}

func (ps *PHPSerializer) toString(byt []byte) (interface{}, error) {

	s := bytes.IndexByte(byt, ':')
	if s == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	e := bytes.IndexByte(byt[s+1:], ':')
	if e == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	e += s + 1
	src := byt[s+1 : e]
	length, err := strconv.Atoi(string(src))
	if err != nil {
		return "", err
	}

	s = e + 1
	e = bytes.IndexByte(byt[s:], ';')
	if e == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	e += s
	src = byt[s+1 : e]
	src = bytes.Trim(src, "\"")
	if len(src) != length {
		return "", fmt.Errorf("the length of string is not equal")
	}

	dst := make([]byte, length)
	copy(dst, src)
	return string(dst), nil
}

func (ps *PHPSerializer) toFloat64(byt []byte) (interface{}, error) {
	src, err := SubBytes(byt, ':', ';')
	if err != nil {
		return nil, err
	}

	return strconv.ParseFloat(string(src), 64)
}

func (ps *PHPSerializer) toInt(byt []byte) (interface{}, error) {
	src, err := SubBytes(byt, ':', ';')
	if err != nil {
		return nil, err
	}

	return strconv.ParseInt(string(src), 10, 64)
}

func (ps *PHPSerializer) toBool(byt []byte) (interface{}, error) {
	src, err := SubBytes(byt, ':', ';')
	if err != nil {
		return nil, err
	}

	return strconv.ParseBool(string(src))
}

func (ps PHPSerializer) toMap(byt []byte) (interface{}, error) {
	s := bytes.IndexByte(byt, ':')
	if s == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	e := bytes.IndexByte(byt[s+1:], ':')
	if e == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	e += s + 1
	src := byt[s+1 : e]
	length, err := strconv.Atoi(string(src))
	if err != nil {
		return "", err
	}

	s = bytes.IndexByte(byt[e+1:], '{')
	if s == -1 {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	s += e + 1
	if !bytes.HasSuffix(byt[s+1:], []byte{'}'}) {
		return nil, fmt.Errorf("this sub-string dose not found")
	}

	src = byt[s+1 : len(byt)-1]
	src = bytes.TrimSpace(src)

	n := 0
	data := make(map[string]interface{})
	i := 0
	j := 0

	for j < len(src) {
		j = bytes.IndexByte(src[i:], ';')
		if j == -1 {
			if n == length {
				break
			} else {
				return nil, fmt.Errorf("this sub-string dose not found")
			}
		}

		j += i
		var name string
		if src[i] == ps.phpString {
			inter, err := ps.toString(src[i : j+1])
			if err != nil {
				return nil, err
			}

			name = inter.(string)
		} else {
			inter, err := ps.toInt(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

			val := inter.(int64)
			name = strconv.FormatInt(val, 10)
		}

		j++
		i = j
		c := src[i]
		var inter interface{}
		switch c {
		case ps.phpArray, ps.phpObject:
			count := 1
			j = bytes.IndexByte(src[i:], '{')
			if j == -1 {
				return nil, fmt.Errorf("invalid string")
			}

			for k := j + i + 1; k < len(src); k++ {
				if src[k] == '{' {
					count++
				} else if src[k] == '}' {
					if count == 0 {
						return nil, fmt.Errorf("invalid string")
					}
					count--
				} else if count == 0 {
					j = k
					break
				}
				j = k
			}

			if count > 0 {
				return nil, fmt.Errorf("invalid string")
			}

			inter, err = ps.toMap(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

		case ps.phpString:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

			j += i
			inter, err = ps.toString(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

		case ps.phpInt:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

			j += i
			inter, err = ps.toInt(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

		case ps.phpFloat:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

			j += i
			inter, err = ps.toFloat64(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

		case ps.phpBool:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

			j += i
			inter, err = ps.toBool(src[i : j+1])
			if err != nil {
				return nil, fmt.Errorf("this sub-string dose not found")
			}

		default:
			return nil, fmt.Errorf("invalid string")
		}

		j++
		i = j
		n++
		data[name] = inter
	}

	return data, nil
}

func (ps *PHPSerializer) Serialize(byt []byte) (interface{}, error) {
	byt = bytes.TrimSpace(byt)
	if len(byt) == 0 {
		return nil, nil
	}

	c := byt[0]
	switch c {
	case ps.phpObject, ps.phpArray:
		return ps.toMap(byt)

	case ps.phpString:
		return ps.toString(byt)

	case ps.phpInt:
		return ps.toInt(byt)

	case ps.phpFloat:
		return ps.toFloat64(byt)

	case ps.phpBool:
		return ps.toBool(byt)

	default:
		return nil, fmt.Errorf("invalid string")
	}

}

func PHPSerialize(byt []byte) (interface{}, error) {
	serializer := NewDefaultPHPSerializer()
	return serializer.Serialize(byt)
}

func StringToTime(layout string, val string) (int64, error) {
	t, err := time.Parse(layout, val)
	if err != nil {
		return 0, err
	}

	return t.UnixNano(), nil
}
