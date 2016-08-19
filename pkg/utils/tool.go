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
	"regexp"
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
		return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt))
	}

	e := bytes.IndexByte(byt[s+1:], ':')
	if e == -1 {
		return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[s+1:]))
	}

	e += s + 1
	src := byt[s+1 : e]
	length, err := strconv.Atoi(string(src))
	if err != nil {
		return "", err
	}

	s = bytes.IndexByte(byt[e+1:], '{')
	if s == -1 {
		return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[e+1:]))
	}

	s += e + 1
	suffixLen := 1
	if !bytes.HasSuffix(byt[s+1:], []byte{'}'}) {
		if !bytes.HasSuffix(byt[s+1:], []byte{'}', ';'}) {
			return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[s+1:]))
		}
		suffixLen = 2
	}

	src = byt[s+1 : len(byt)-suffixLen]
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
				return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(src[i:]))
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
				return nil, err
			}

			val := inter.(int64)
			name = strconv.FormatInt(val, 10)
		}

		j++
		i = j
		if j >= len(src) {
			break
		}

		c := src[i]
		var inter interface{}
		switch c {
		case ps.phpArray, ps.phpObject:
			count := 1
			j = bytes.IndexByte(src[i:], '{')
			if j == -1 {
				return nil, fmt.Errorf("invalid string '%s'", string(src[i:]))
			}

			j += i + 1
			for k := j; k < len(src); k++ {
				if src[k] == '{' {
					count++
				} else if src[k] == '}' {
					if count == 0 {
						return nil, fmt.Errorf("invalid string '%s'", string(src))
					}

					count--

					if count == 0 {
						j = k
						break
					}
				}
			}

			if count > 0 {
				return nil, fmt.Errorf("invalid string '%s'", string(src))
			}

			inter, err = ps.toMap(src[i : j+1])
			if err != nil {
				return nil, err
			}

		case ps.phpString:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[i:]))
			}

			j += i
			inter, err = ps.toString(src[i : j+1])
			if err != nil {
				return nil, err
			}

		case ps.phpInt:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[i:]))
			}

			j += i
			inter, err = ps.toInt(src[i : j+1])
			if err != nil {
				return nil, err
			}

		case ps.phpFloat:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[i:]))
			}

			j += i
			inter, err = ps.toFloat64(src[i : j+1])
			if err != nil {
				return nil, err
			}

		case ps.phpBool:
			j = bytes.IndexByte(src[i:], ';')
			if j == -1 {
				return nil, fmt.Errorf("this sub-string dose not found in the '%s'", string(byt[i:]))
			}

			j += i
			inter, err = ps.toBool(src[i : j+1])
			if err != nil {
				return nil, err
			}

		default:
			return nil, fmt.Errorf("invalid string '%s'", string(src))
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

/**
* time format to timestemp
* @author；rey
* @date: 2015.9.11
* @param: @layout format, @val: timestemp
* @return；timestemp and error
 */
func StringToTime(layout string, val string) (int64, error) {
	t, err := time.Parse(layout, val)
	if err != nil {
		return 0, err
	}

	return t.UnixNano(), nil
}

/**
* timestemp to foramt string
* @author；rey
* @date: 2015.9.11
* @param: layout: time format, @ns: nanoseconds
* @return；string
 */
func TimeToString(layout string, ns int64) string {
	t := time.Time{}
	t.Add(time.Duration(ns))

	return t.Format(layout)
}

/**
* input will be filtered by the xsser
* @author；rey
* @date: 2015.9.11
* @param: @src string
* @return；string
 */
func XSS(src string) string {
	// remove all non-printable characters. CR(0a) and LF(0b) and TAB(9) are allowed
	// this prevents some character re-spacing such as <java\0script>
	// note that you have to handle splits with \n, \r, and \t later since they *are* allowed in some inputs
	reg := regexp.MustCompile("[\x00-\x08\x0b-\x0c\x0e-\x19]")
	src = reg.ReplaceAllString(src, "")
	// straight replacements, the user should never need these since they're normal characters
	// this prevents like <IMG SRC=@avascript:alert('XSS')>
	search := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"1234567890!@#$%^&*()" +
		"~`\";:?+/={}[]-_|'\\"

	for i := 0; i < len(search); i++ {
		// ;? matches the ;, which is optional
		// 0{0,7} matches any padded zeros, which are optional and go up to 8 chars
		// @ @ search for the hex values
		reg, err := regexp.Compile(fmt.Sprintf("(?i:&#[xX]0{0,8}%X;?)", search[i]))
		if err != nil {
			continue
		}
		src = reg.ReplaceAllString(src, string(search[i]))
		// @ @ 0{0,7} matches '0' zero to seven times
		reg, err = regexp.Compile(fmt.Sprintf("(?i:�{0,8}%d;?)", search[i]))
		if err != nil {
			continue
		}
		src = reg.ReplaceAllString(src, string(search[i])) //with a
	}

	// now the only remaining whitespace attacks are \t, \n, and \r
	labels := []string{
		"javascript", "vbscript", "expression",
		"applet", "meta", "xml",
		"blink", "link", "style",
		"object", "iframe", "frame",
		"ilayer", "layer", "bgsound",
		"title", "base",
	}

	eventTags := []string{
		"onabort", "onactivate", "onafterprint",
		"onafterupdate", "onbeforeactivate", "onbeforecopy",
		"onbeforecut", "onbeforedeactivate", "onbeforeeditfocus",
		"onbeforepaste", "onbeforeprint", "onbeforeunload",
		"onbeforeupdate", "onblur", "onbounce",
		"oncellchange", "onchange", "onclick",
		"oncontextmenu", "oncontrolselect", "oncopy",
		"oncut", "ondataavailable", "ondatasetchanged",
		"ondatasetcomplete", "ondblclick", "ondeactivate",
		"ondrag", "ondragend", "ondragenter",
		"ondragleave", "ondragover", "ondragstart",
		"ondrop", "onerror", "onerrorupdate",
		"onfilterchange", "onfinish", "onfocus",
		"onfocusin", "onfocusout", "onhelp",
		"onkeydown", "onload", "onlosecapture",
		"onmousedown", "onmouseenter", "onmouseleave",
		"onmousemove", "onmouseout", "onmouseover",
		"onmouseup", "onmousewheel", "onmove",
		"onmoveend", "onmovestart", "onpaste",
		"onpropertychange", "onreadystatechange", "onreset",
		"onresize", "onresizeend", "onresizestart",
		"onrowenter", "onrowexit", "onrowsdelete",
		"onrowsinserted", "onscroll", "onselect",
		"onselectionchange", "onselectstart", "onstart",
		"onstop", "onsubmit", "onunload",
	}

	set := append(eventTags, labels...)

	subPattern := "((&#[xX]0{0,8}([9ab]);)|(�{0,8}([9|10|13]);))*"
	pattern := ""
	for _, label := range set {
		pattern = ""
		for i, char := range label {
			if i > 0 {
				pattern += subPattern
			}
			pattern += string(char)
		}
		reg, err := regexp.Compile(fmt.Sprintf("(?i:%s)", pattern))
		if err != nil {
			continue
		}
		repl := label[0:2] + "<x>" + label[2:] // add in <> to nerf the tag
		src = reg.ReplaceAllString(src, repl)
	}

	return src
}
