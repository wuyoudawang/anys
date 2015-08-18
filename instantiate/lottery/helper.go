package lottery

import (
	"fmt"
	"strings"

	"anys/pkg/utils"
)

const (
	mask = 0x0000000f
)

type convert struct {
	bits     int
	terms    int
	cache    map[string]int
	start    int
	split    func(string) []string
	generate func(n int, fix ...int) []string
}

// range (a, b)
func newConvert(a, b int) *convert {
	c := &convert{}
	c.cache = make(map[string]int)

	for tmp := b; tmp > 0; tmp /= 10 {
		c.terms++
	}

	r := b - a + 1
	key := a
	for i := 1; i <= r; i++ {
		c.cache[fmt.Sprintf("%d", key)] = i
		key++
	}

	c.bits = 0
	c.start = a
	for ; r > 0; r >>= 1 {
		c.bits += 1
	}
	return c
}

func (c *convert) getRangeSize() int {
	return len(c.cache)
}

func (c *convert) setSplitFunc(h func(string) []string) *convert {
	c.split = h

	return c
}

func (c *convert) getSplitFunc() func(string) []string {
	if c.split == nil {
		c.split = func(str string) []string {
			var set []string

			if len(str)%c.terms != 0 {
				return set
			}

			for i := 0; i < len(str); i += c.terms {
				if c.terms > 1 {
					set = append(set, strings.TrimLeft(str[i:i+c.terms], "0"))
				} else {
					set = append(set, str[i:i+c.terms])
				}
			}

			return set
		}
	}

	return c.split
}

func (c *convert) getFlag(r string) (int, error) {
	if f, ok := c.cache[r]; ok {
		return f, nil
	}

	return -1, fmt.Errorf("key(%s) do not exist", r)
}

func (c *convert) shiftLeft(src, n int) int {
	if n > 1 {
		src <<= uint(c.bits * (n - 1))
	}
	return src
}

func (c *convert) shiftRight(src, n int) int {
	if n > 1 {
		src >>= uint(c.bits * (n - 1))
	}
	return src
}

func (c *convert) integer(str string) (int, error) {
	val := 0
	var b uint = 0

	set := c.getSplitFunc()(str)
	if len(set) == 0 {
		return -1, fmt.Errorf("string(%s) is invalid", str)
	}

	for _, s := range set {
		f, err := c.getFlag(s)
		if err != nil {
			return -1, err
		}

		if b > 0 {
			val |= int(f << b)
		} else {
			val |= f
		}

		b += uint(c.bits)
	}

	return val, nil
}

func (c *convert) number(val int) string {
	str := ""
	stack := []string{}

	for i := 1; val > 0; val >>= uint(c.bits) {

		i++
		stack = append(stack, fmt.Sprintf("%0*d", c.terms, val&mask+c.start-1))
	}

	for i := len(stack) - 1; i >= 0; i-- {
		str += stack[i]
	}

	return str
}

func (c *convert) getGenerateFn() func(n int, fix ...int) (nums []string) {
	if c.generate == nil {
		c.generate = func(n int, fix ...int) (nums []string) {
			i := n - len(fix)
			prefix := ""
			for _, v := range fix {
				prefix += fmt.Sprintf("%0*d", c.terms, v)
			}

			if i >= 0 {
				for j := 0; j < i; j++ {

					var temp []string
					for s := 0; s < len(c.cache); s++ {

						if len(nums) > 0 {
							for _, item := range nums {
								num := item + fmt.Sprintf("%0*d", c.terms, c.start+s)
								temp = append(temp, num)
							}
						} else {
							num := prefix + fmt.Sprintf("%0*d", c.terms, c.start+s)
							temp = append(temp, num)
						}
					}
					nums = temp
				}
			}

			return
		}
	}

	return c.generate
}

func (c *convert) formatInt(v ...int) []string {
	var rel []string

	for _, val := range v {
		s := fmt.Sprintf("%0*d", c.terms, val)
		rel = append(rel, s)
	}

	return rel
}

func (c *convert) formatString(s ...string) []string {
	var rel []string

	for _, str := range s {
		if len(str) < c.terms {
			str = strings.Repeat("0", c.terms-len(str)) + str
			rel = append(rel, str)
		}
	}

	return rel
}

func (c *convert) match(src int, r string, n int) (bool, error) {
	i, err := c.getFlag(r)
	if err != nil {
		return false, err
	}

	i = c.shiftLeft(i, n)
	return c.rawMatch(src, i), nil
}

func (c *convert) rawMatch(a, b int) bool {
	return (b != 0 && a&b-b == 0)
}

func (c *convert) getLen(key int) int {

	l := 0
	for ; key > 0; key >>= uint(c.bits) {

		if key&mask > 0 {
			l++
		}
	}

	return l
}

func (c *convert) getChildrenKey(parent int, n int) (children []int) {
	if c.getLen(parent) == n {
		return []int{parent}
	}

	switch n {
	case 1:
		for i := 1; parent > 0; parent >>= uint(c.bits) {
			if k := parent & mask; k > 0 {

				k = c.shiftLeft(k, i)

				children = append(children, k)
			}
			i++
		}
	default:
		for i := 1; parent > 0; parent >>= uint(c.bits) {
			if k := parent & mask; k > 0 {

				k = c.shiftLeft(k, i)

				if tmp := c.shiftLeft(parent&int(utils.NOT(mask)), i); tmp > 0 {
					keys := c.getChildrenKey(tmp, n-1)

					for _, key := range keys {
						key |= k
						children = append(children, key)
					}
				}

			}
			i++
		}
	}

	return
}

func (c *convert) getPermutation(set []string, n int) (rel [][]string) {
	if len(set) == 0 || n > len(set) {
		return
	}

	for i, num := range set {

		if n == 1 {
			rel = append(rel, []string{num})
		} else {
			tmp := c.getPermutation(utils.ContactArrString(set[0:i], set[i+1:]), n-1)

			for _, s := range tmp {
				s = append(s, num)
				rel = append(rel, s)
			}
		}

	}

	return
}

func (c *convert) getCombination(set []string, n int) (rel [][]string) {
	if len(set) == 0 || n > len(set) {
		return
	}

	if n == len(set) {
		rel = append(rel, set)
		return
	}

	if n == 1 {
		for _, num := range set {
			rel = append(rel, []string{num})
		}

		return
	}

	firstNum := set[0]
	sub := c.getCombination(set[1:], n-1)
	for _, e := range sub {
		e = append(e, firstNum)
		rel = append(rel, e)
	}

	self := c.getCombination(set[1:], n)
	rel = append(rel, self...)

	return

}

func (c *convert) repeatNum(set []string, num string, n int) [][]string {
	var dst [][]string

	for i := 0; i < n; i++ {
		var tmp [][]string

		for j := 0; j <= len(set); j++ {

			if len(dst) > 0 {
				for _, item := range dst {
					pos := 0
					for k, s := range item {
						if j == pos {
							newItem := []string{}
							newItem = append(newItem, item[0:k]...)
							newItem = append(newItem, num)
							newItem = append(newItem, item[k:]...)
							tmp = append(tmp, newItem)
							break
						}

						if s != num {
							pos++
						}
					}
				}
			} else {
				item := []string{}
				item = append(item, set[0:j]...)
				item = append(item, num)
				item = append(item, set[j:]...)
				tmp = append(tmp, item)
			}

		}

		dst = tmp
	}

	return dst
}
