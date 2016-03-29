package lottery

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/liuzhiyi/anys/instantiate/lottery/model"
	"github.com/liuzhiyi/anys/log"
	"github.com/liuzhiyi/anys/pkg/utils"
)

type Method struct {
	maxReward    float64
	combinations int64
	rate         float64
	maxPunts     int64
	currPunts    int64
}

const (
	success = iota
	done
	cancle
)

var errBetFormat = errors.New("unexpected format")

func NewMethod(conf string) (*Method, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(conf), &m)
	if err != nil {
		log.Info("config of method '%s' has an error '%s'", conf, err.Error())
		return nil, err
	}

	method := &Method{}

	var exists bool
	if method.maxReward, exists = m["maxReward"].(float64); !exists {
		method.maxReward = 200000
	}

	if combinations, exists := m["combinations"].(float64); !exists {
		return nil, fmt.Errorf("combinations must be set.")
	} else {
		method.combinations = int64(combinations)
	}

	if maxPunts, exists := m["maxPunts"].(float64); !exists {
		method.maxPunts = 1
	} else {
		method.maxPunts = int64(maxPunts)
	}

	if method.rate, exists = m["rate"].(float64); !exists {
		method.rate = 1.0
	}

	return method, nil
}

func (m *Method) getReward(p *model.Projects, bets float64) float64 {
	reward := p.GetFloat64("omodel") /
		float64(1000*p.GetMode()) *
		float64(m.combinations*p.GetInt64("multiple")) *
		bets / m.rate

	if reward > m.maxReward {
		reward = m.maxReward
	}
	return utils.Round(reward, 4)
}

func (m *Method) getFuncName(name string) string {
	return "Call" + strings.ToUpper(name[:1]) + name[1:]
}

func (m *Method) splitBit(bet string) []string {
	return strings.Split(bet, bitSeparator)
}

func (m *Method) splitNum(src string) []string {
	return strings.Split(src, numSeparator)
}

func (m *Method) addRecord(lty *Lottery, key int, id int64, reward float64) {
	lty.t.AddRecord(key, id, reward)
}

func (m *Method) checkPunts(lty *Lottery, id int64, key int) bool {
	n := lty.t.Get(key)
	if n == nil {
		return true
	}

	currPunts := n.getRecordCurrPunts(id)
	return currPunts < m.maxPunts
}

// 定位胆
func (m *Method) CallOneNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	for i, item := range set {

		if item != placeHolder {

			nums := m.splitNum(item)
			for _, num := range nums {

				key, err := lty.cvt.getFlag(num)
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				key = lty.cvt.shiftLeft(key, i+1)

				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}

			}
			return nil, success
		}
	}

	return errBetFormat, cancle
}

func (m *Method) twoNum(lty *Lottery, bet string, p *model.Projects, frist, second int) (error, int) {
	set := m.splitBit(bet)
	if frist < 1 || second > len(set) {
		return errBetFormat, cancle
	}

	one := set[frist-1]
	two := set[second-1]

	if one == placeHolder || two == placeHolder {
		return errBetFormat, cancle
	}

	var key int
	var err error
	twoPart := m.splitNum(two)
	for _, num := range m.splitNum(one) {
		key, err = lty.cvt.getFlag(num)
		if err != nil {
			return err, cancle
		}

		key = lty.cvt.shiftLeft(key, frist)

		for _, n := range twoPart {
			k, err := lty.cvt.getFlag(n)
			if err != nil {
				return err, cancle
			}

			k = lty.cvt.shiftLeft(k, second)

			k |= key

			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, k) {
				m.addRecord(lty, k, id, m.getReward(p, 1))
			}

		}
	}

	return nil, success
}

// 前二直选
func (m *Method) CallPrevTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoNum(lty, bet, p, 1, 2)
}

// 后二直选
func (m *Method) CallLastTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	first := lty.getLast(2)
	second := first + 1
	return m.twoNum(lty, bet, p, first, second)
}

func (m *Method) threeNum(lty *Lottery, bet string, p *model.Projects, frist, second, third int) (error, int) {
	set := m.splitBit(bet)
	if len(set) < third {
		return errBetFormat, cancle
	}

	one := set[frist-1]
	two := set[second-1]
	three := set[third-1]

	if one == placeHolder || two == placeHolder || three == placeHolder {
		return errBetFormat, cancle
	}

	var key int
	var err error
	twoPart := m.splitNum(two)
	threePart := m.splitNum(three)
	for _, num := range m.splitNum(one) {
		key, err = lty.cvt.getFlag(num)
		if err != nil {
			return err, cancle
		}

		key = lty.cvt.shiftLeft(key, frist)

		for _, num := range twoPart {
			k, err := lty.cvt.getFlag(num)
			if err != nil {
				return err, cancle
			}

			k = lty.cvt.shiftLeft(k, second)

			k |= key

			for _, num := range threePart {
				tk, err := lty.cvt.getFlag(num)
				if err != nil {
					return err, cancle
				}

				tk = lty.cvt.shiftLeft(tk, third)

				tk |= k
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, tk) {
					m.addRecord(lty, tk, id, m.getReward(p, 1))
				}
			}

		}
	}

	return nil, success
}

// 前三直选
func (m *Method) CallPrevThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeNum(lty, bet, p, 1, 2, 3)
}

// 后三直选
func (m *Method) CallLastThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	f := lty.getLast(3)
	s := f + 1
	t := s + 1

	if lty.count < t {
		return errBetFormat, cancle
	}

	return m.threeNum(lty, bet, p, f, s, t)
}

// 中三直选
func (m *Method) CallMidThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	f := lty.getMid(3)
	s := f + 1
	t := s + 1

	if lty.count < t {
		return errBetFormat, cancle
	}

	return m.threeNum(lty, bet, p, f, s, t)
}

/**
*五星直选
***/
func (m *Method) CallAllNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyNum(lty, bet, p)
}

func (m *Method) getKeys(lty *Lottery, bet string, pos ...int) ([]int, error) {
	var keys, prev []int
	var err error

	if len(pos) == 0 {
		if strings.Contains(bet, placeHolder) {
			return keys, errBetFormat
		}

		set := m.splitBit(bet)
		for i := 1; i <= lty.count; i++ {

			keys = []int{}
			for _, num := range m.splitNum(set[i-1]) {
				key, err := lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = lty.cvt.shiftLeft(key, i)
				if len(prev) > 0 {
					for _, k := range prev {
						k |= key
						keys = append(keys, k)
					}
				} else {
					keys = append(keys, key)
				}
			}
			prev = keys
		}
	} else {
		set := m.splitBit(bet)
		h := 0

		for j, i := range pos {

			if len(set) == lty.count {
				h = i - 1
			} else {
				h = j
			}

			if i > lty.count || set[h] == placeHolder {
				return keys, errBetFormat
			}
		}

		for j, i := range pos {

			if len(set) == lty.count {
				h = i - 1
			} else {
				h = j
			}

			keys = []int{}
			for _, num := range m.splitNum(set[h]) {
				key, err := lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = lty.cvt.shiftLeft(key, i)
				if len(prev) > 0 {
					for _, k := range prev {
						k |= key
						keys = append(keys, k)
					}
				} else {
					keys = append(keys, key)
				}
			}
			prev = keys
		}
	}

	return keys, err
}

func (m *Method) anyNum(lty *Lottery, bet string, p *model.Projects, pos ...int) (error, int) {
	keys, err := m.getKeys(lty, bet, pos...)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		id := p.GetInt64("projectid")
		if m.checkPunts(lty, id, key) {
			m.addRecord(lty, key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*五星组合
***/
func (m *Method) CallAllComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {

	keys, err := m.getKeys(lty, bet)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		for i := 1; i <= lty.count; i++ {
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {

				m.addRecord(lty, key, id, m.getReward(p, 1/math.Pow10(i-1)))
			}

			key = key & int(utils.NOT(int64(lty.cvt.shiftLeft(mask, i))))
		}
	}

	return nil, success
}

/**
*五星组选120
***/
func (m *Method) CallAll120Num(lty *Lottery, bet string, p *model.Projects) (error, int) {

	if strings.Contains(bet, placeHolder) {
		return errBetFormat, cancle
	}

	set := m.splitNum(bet)
	set = lty.cvt.formatString(set...)
	permutation := lty.cvt.getPermutation(set, 5)

	for _, s := range permutation {
		key, err := lty.cvt.integer(strings.Join(s, ""))
		if err != nil {
			return err, cancle
		}

		id := p.GetInt64("projectid")
		if m.checkPunts(lty, id, key) {
			m.addRecord(lty, key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*五星组选60 10*(9*8*7/3*2)
***/
func (m *Method) CallAll60Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = lty.cvt.formatString(singlePart...)
	singleSet := lty.cvt.getPermutation(singlePart, 3)
	doubleNums := m.splitNum(set[0])
	doubleNums = lty.cvt.formatString(doubleNums...)

	for _, num := range doubleNums {
		for _, nums := range singleSet {

			permutation := lty.cvt.repeatNum(nums, num, 2)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选30
***/
func (m *Method) CallAll30Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = lty.cvt.formatString(singlePart...)
	doubleNums := m.splitNum(set[0])
	doubleNums = lty.cvt.formatString(doubleNums...)
	com := lty.cvt.getCombination(doubleNums, 2)
	if len(com) == 0 {
		return errBetFormat, cancle
	}

	for _, singleNum := range singlePart {
		for _, elem := range com {
			set := lty.cvt.repeatNum([]string{singleNum}, elem[0], 2)

			for _, item := range set {
				data := lty.cvt.repeatNum(item, elem[1], 2)

				for _, s := range data {
					key, err := lty.cvt.integer(strings.Join(s, ""))
					if err != nil {
						return err, cancle
					}

					id := p.GetInt64("projectid")
					if m.checkPunts(lty, id, key) {
						m.addRecord(lty, key, id, m.getReward(p, 1))
					}
				}
			}
		}
	}

	return nil, success
}

/**
*五星选20
***/
func (m *Method) CallAll20Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = lty.cvt.formatString(singlePart...)
	singleSet := lty.cvt.getPermutation(singlePart, 2)
	trebleNums := m.splitNum(set[0])
	trebleNums = lty.cvt.formatString(trebleNums...)

	for _, num := range trebleNums {
		for _, nums := range singleSet {

			permutation := lty.cvt.repeatNum(nums, num, 3)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选10
***/
func (m *Method) CallAll10Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	doubleNums := m.splitNum(set[1])
	doubleNums = lty.cvt.formatString(doubleNums...)
	trebleNums := m.splitNum(set[0])
	trebleNums = lty.cvt.formatString(trebleNums...)

	for _, doubleNum := range doubleNums {
		for _, trebleNum := range trebleNums {

			permutation := lty.cvt.repeatNum([]string{doubleNum, doubleNum}, trebleNum, 3)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选5
***/
func (m *Method) CallAll5Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singleNums := m.splitNum(set[1])
	singleNums = lty.cvt.formatString(singleNums...)
	fourfoldNums := m.splitNum(set[0])
	fourfoldNums = lty.cvt.formatString(fourfoldNums...)

	for _, fourfoldNum := range fourfoldNums {
		for _, singleNum := range singleNums {

			set := []string{fourfoldNum, fourfoldNum, fourfoldNum, fourfoldNum}
			permutation := lty.cvt.repeatNum(set, singleNum, 1)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*四星直选
***/
func (m *Method) CallFourNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyNum(lty, bet, p, 2, 3, 4, 5)
}

/**
*四星组合
***/
func (m *Method) CallFourComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	keys, err := m.getKeys(lty, bet, 2, 3, 4, 5)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		for i := 2; i <= lty.count; i++ {
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1/math.Pow10(i-2)))
			}

			key = key & int(utils.NOT(int64(lty.cvt.shiftLeft(mask, i))))
		}
	}

	return nil, success
}

/**
*四星组选24
***/
func (m *Method) CallFour24Num(lty *Lottery, bet string, p *model.Projects) (error, int) {

	if strings.Contains(bet, placeHolder) {
		return errBetFormat, cancle
	}

	set := m.splitNum(bet)
	set = lty.cvt.formatString(set...)
	permutation := lty.cvt.getPermutation(set, 4)
	for _, s := range permutation {

		key, err := lty.cvt.integer(strings.Join(s, ""))
		if err != nil {
			return err, cancle
		}

		key = lty.cvt.shiftLeft(key, 2)
		id := p.GetInt64("projectid")
		if m.checkPunts(lty, id, key) {
			m.addRecord(lty, key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*四星组选12
***/
func (m *Method) CallFour12Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = lty.cvt.formatString(singlePart...)
	singleSet := lty.cvt.getPermutation(singlePart, 2)
	doubleNums := m.splitNum(set[0])
	doubleNums = lty.cvt.formatString(doubleNums...)

	for _, num := range doubleNums {
		for _, nums := range singleSet {

			permutation := lty.cvt.repeatNum(nums, num, 2)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, 2)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*四星组选6
***/
func (m *Method) CallFour6Num(lty *Lottery, bet string, p *model.Projects) (error, int) {

	doubleNums := m.splitNum(bet)
	doubleNums = lty.cvt.formatString(doubleNums...)
	com := lty.cvt.getCombination(doubleNums, 2)

	for _, item := range com {

		permutation := lty.cvt.repeatNum([]string{item[0], item[0]}, item[1], 2)
		for _, s := range permutation {
			key, err := lty.cvt.integer(strings.Join(s, ""))
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, 2)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/**
*四星组选4
***/
func (m *Method) CallFour4Num(lty *Lottery, bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = lty.cvt.formatString(singlePart...)
	trebleNums := m.splitNum(set[0])
	trebleNums = lty.cvt.formatString(trebleNums...)

	for _, trebleNum := range trebleNums {
		for _, singleNum := range singlePart {

			permutation := lty.cvt.repeatNum([]string{singleNum}, trebleNum, 3)
			for _, s := range permutation {
				key, err := lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, 2)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

func (m *Method) anyOneNum(lty *Lottery, bet string, p *model.Projects, n, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)
	all := lty.cvt.getAllNum()
	all = lty.cvt.formatString(all...)
	permu := lty.cvt.getSelection(all, n-1)

	for _, num := range nums {

		for _, item := range permu {

			set := lty.cvt.repeatNum(item, num, 1)

			for _, str := range set {
				key, err := lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*前三(不定胆)
***/
func (m *Method) CallPrevThreeOneNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(lty, bet, p, 3, 1)
}

/**
*中三(不定胆)
***/
func (m *Method) CallMidThreeOneNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(lty, bet, p, 3, lty.getMid(2))
}

/**
*后三(不定胆)
***/
func (m *Method) CallLastThreeOneNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(lty, bet, p, 3, lty.getLast(3))
}

func (m *Method) anyAnyNum(lty *Lottery, bet string, p *model.Projects, r, n, s int) (error, int) {
	if n < r {
		return fmt.Errorf("args 'n(%d)'' should greater than %d", n, r), cancle
	}

	if n > lty.count {
		return fmt.Errorf("overflow lottery's max bit"), cancle
	}

	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)
	permu := lty.cvt.getPermutation(nums, r)
	all := lty.cvt.getAllNum()
	all = lty.cvt.formatString(all...)

	// for _, item := range permu {

	// 	for i := 0; i < lty.count; i++ {

	// 		num := lty.cvt.formatInt(lty.cvt.start + i)[0]
	// 		set := lty.cvt.repeatNum(item, num, n-r)

	// 		for _, str := range set {
	// 			key, err := lty.cvt.integer(strings.Join(str, ""))
	// 			if err != nil {
	// 				return err, cancle
	// 			}

	// 			key = lty.cvt.shiftLeft(key, s)
	// 			id := p.GetInt64("projectid")
	// 			if m.checkPunts(id, key) {
	// 				m.addRecord(key, id, m.getReward(p, 1))
	// 			}
	// 		}
	// 	}

	// }

	for _, item := range permu {

		var set [][]string
		for i := 0; i < n-r; i++ {

			var tmp [][]string
			for _, num := range all {

				if len(set) > 0 {
					for _, elem := range set {
						childSet := lty.cvt.repeatNum(elem, num, 1)
						tmp = append(tmp, childSet...)
					}
				} else {
					childSet := lty.cvt.repeatNum(item, num, 1)
					tmp = append(tmp, childSet...)
				}

			}

			set = tmp
		}

		for _, str := range set {

			key, err := lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}

	}

	return nil, success
}

/**
*前三二字不定胆
***/
func (m *Method) CallPrevThreeTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 2, 3, 1)
}

/**
*中三二字不定胆
***/
func (m *Method) CallMidThreeTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 2, 3, lty.getMid(3))
}

/**
*后三二字不定胆
***/
func (m *Method) CallLastThreeTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 2, 3, lty.getLast(3))
}

func (m *Method) threeComNum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)

	com := lty.cvt.getCombination(nums, 2)
	for _, item := range com {
		for i := 0; i < 2; i++ {
			var set [][]string
			if i == 1 {
				set = lty.cvt.repeatNum([]string{item[i], item[i]}, item[0], 1)
			} else {
				set = lty.cvt.repeatNum([]string{item[i], item[i]}, item[1], 1)
			}

			for _, str := range set {
				key, err := lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}

	}

	return nil, success
}

/***
*组选3(前三，中三，后三)
***/
func (m *Method) CallPrevThreeComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeComNum(lty, bet, p, 1)
}

func (m *Method) CallMidThreeComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeComNum(lty, bet, p, lty.getMid(3))
}

func (m *Method) CallLastThreeComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeComNum(lty, bet, p, lty.getLast(3))
}

func (m *Method) sixComNum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)

	permu := lty.cvt.getPermutation(nums, 3)
	for _, str := range permu {
		key, err := lty.cvt.integer(strings.Join(str, ""))
		if err != nil {
			return err, cancle
		}

		key = lty.cvt.shiftLeft(key, s)
		id := p.GetInt64("projectid")
		if m.checkPunts(lty, id, key) {
			m.addRecord(lty, key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/***
*组选6(前三，中三，后三)
***/
func (m *Method) CallPrevSixComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.sixComNum(lty, bet, p, 1)
}

func (m *Method) CallMidSixComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.sixComNum(lty, bet, p, lty.getMid(3))
}

func (m *Method) CallLastSixComNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.sixComNum(lty, bet, p, lty.getLast(3))
}

func (m *Method) twoComNum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)

	permu := lty.cvt.getPermutation(nums, 2)
	for _, str := range permu {
		key, err := lty.cvt.integer(strings.Join(str, ""))
		if err != nil {
			return err, cancle
		}

		key = lty.cvt.shiftLeft(key, s)
		id := p.GetInt64("projectid")
		if m.checkPunts(lty, id, key) {
			m.addRecord(lty, key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/***
*二星组选(前二， 后二)
***/
func (m *Method) CallPrevTwoTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoComNum(lty, bet, p, 1)
}

func (m *Method) CallLastTwoTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoComNum(lty, bet, p, lty.getLast(2))
}

func (m *Method) twoSizeOdd(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	var part1, part2 []string
	for i, item := range set {
		for _, num := range m.splitNum(item) {
			var tmp []string

			switch num {
			case "0": // 大
				tmp = lty.cvt.size(true)
			case "1": // 小
				tmp = lty.cvt.size(false)
			case "2": // 单
				tmp = lty.cvt.odd(true)
			case "3": // 双
				tmp = lty.cvt.odd(false)
			}

			if i == 0 {
				part1 = append(part1, tmp...)
			} else {
				part2 = append(part2, tmp...)
			}
		}
	}

	part1 = lty.cvt.formatString(part1...)
	part2 = lty.cvt.formatString(part2...)

	for _, num1 := range part1 {

		for _, num2 := range part2 {
			num2 = num1 + num2

			key, err := lty.cvt.integer(num2)
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/***
*二星大小单双(前二， 后二)
***/
func (m *Method) CallPrevTwoSizeOdd(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSizeOdd(lty, bet, p, 1)
}

func (m *Method) CallLastTwoSizeOdd(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSizeOdd(lty, bet, p, lty.getLast(2))
}

func (m *Method) threeSumCom(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := lty.cvt.getSumCom(v, 3)

		for _, str := range com {

			if str[0] != str[1] && str[0] != str[2] && str[1] != str[2] { // 组六
				key, err := lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			} else if str[0] != str[1] || str[0] != str[2] || str[1] != str[2] { // 组三
				key, err := lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.rate /= 2
					m.addRecord(lty, key, id, m.getReward(p, 1))
					m.rate *= 2
				}
			}

		}
	}

	return nil, success
}

/***
*组三选和
***/
func (m *Method) CallPrevThreeSumCom(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(lty, bet, p, 1)
}

func (m *Method) CallMidThreeSumCom(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(lty, bet, p, lty.getMid(3))
}

func (m *Method) CallLastThreeSumCom(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(lty, bet, p, lty.getLast(3))
}

/***
*前三直选和值
***/
func (m *Method) CallPrevThreeSum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSum(lty, bet, p, 1)
}

/**
* 中三直选和值
 */
func (m *Method) CallMidThreeSum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSum(lty, bet, p, lty.getMid(3))
}

/**
* 后三直选和值
 */
func (m *Method) CallLastThreeSum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.threeSum(lty, bet, p, lty.getLast(3))
}

func (m *Method) threeSum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := lty.cvt.getSumCom(v, 3)

		for _, str := range com {

			key, err := lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

func (m *Method) twoSum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = lty.cvt.formatString(nums...)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := lty.cvt.getSumCom(v, 2)

		for _, str := range com {

			key, err := lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/***
*二字直选选和
***/
func (m *Method) CallPrevTwoSum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSum(lty, bet, p, 1)
}

func (m *Method) CallLastTwoSum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSum(lty, bet, p, lty.getLast(2))
}

func (m *Method) twoSumCom(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := lty.cvt.getSumCom(v, 2)

		for _, str := range com {

			if str[0] != str[1] {
				key, err := lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/***
*二字和数(二字组选和)
***/
func (m *Method) CallPrevTwoSumCom(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSumCom(lty, bet, p, 1)
}

func (m *Method) CallLastTwoSumCom(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.twoSumCom(lty, bet, p, lty.getLast(2))
}

func (m *Method) mixThreeNum(lty *Lottery, bet string, p *model.Projects, s int) (error, int) {
	set := m.splitNum(bet)
	if len(set) == 0 {
		return errBetFormat, cancle
	}

	for _, item := range set {
		terms := lty.cvt.terms
		if len(item) != lty.cvt.terms*3 {
			fmt.Println("不合法")
			// log
			continue
		}

		one := item[0:terms]
		two := item[terms*1 : terms*2]
		three := item[terms*2 : terms*3]
		if one != two && one != three && two != three {
			perm := lty.cvt.getPermutation([]string{one, two, three}, 3)

			for _, pi := range perm {
				key, err := lty.cvt.integer(strings.Join(pi, ""))
				if err != nil {
					fmt.Println(err)
					// log
					continue
				}

				key = lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(lty, id, key) {
					m.addRecord(lty, key, id, m.getReward(p, 1))
				}
			}

		} else { //组三
			var selected []string
			var repeat string
			if one == two {
				selected = []string{one, three}
				repeat = one
			} else if two == three {
				selected = []string{one, three}
				repeat = two
			} else if one == three {
				selected = []string{one, two}
				repeat = one
			}

			perm := lty.cvt.getPermutation(selected, 2)

			for _, pi := range perm {
				com := lty.cvt.repeatNum(pi, repeat, 1)
				for _, c := range com {
					key, err := lty.cvt.integer(strings.Join(c, ""))
					if err != nil {
						fmt.Println(err)
						// log
						continue
					}

					key = lty.cvt.shiftLeft(key, s)
					id := p.GetInt64("projectid")
					if m.checkPunts(lty, id, key) {
						m.rate /= 2
						m.addRecord(lty, key, id, m.getReward(p, 1))
						m.rate *= 2
					}
				}
			}
		}

	}

	return nil, success
}

/***
*混合组选
***/
func (m *Method) CallPrevMixThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(lty, bet, p, 1)
}

func (m *Method) CallMidMixThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(lty, bet, p, lty.getMid(3))
}

func (m *Method) CallLastMixThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(lty, bet, p, lty.getLast(3))
}

/***
*趣味玩法 一帆风顺
***/
func (m *Method) CallEverythingGood(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(lty, bet, p, 5, 1)
}

/***
*趣味玩法 好事成双
***/
func (m *Method) CallPairGood(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(lty, bet, p, 2, 5, 1)
}

/***
*趣味玩法 三星报喜
***/
func (m *Method) CallThreeGood(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(lty, bet, p, 3, 5, 1)
}

/***
*趣味玩法 四季发财
***/
func (m *Method) CallFourGood(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(lty, bet, p, 4, 5, 1)
}

func (m *Method) anySingleRepeatNum(lty *Lottery, bet string, p *model.Projects, r, n, s int) (error, int) {
	if n < r {
		return fmt.Errorf("args 'n(%d)' should greater than %d", n, r), cancle
	}

	if n > lty.count {
		return fmt.Errorf("overflow lottery's max bit"), cancle
	}

	nums := m.splitNum(bet)
	var permu [][]string
	for _, num := range nums {
		var item []string

		for i := 0; i < r; i++ {
			item = append(item, num)
		}

		permu = append(permu, item)
	}

	all := lty.cvt.getAllNum()

	for _, item := range permu {

		var set [][]string
		for i := 0; i < n-r; i++ {

			var tmp [][]string
			for _, num := range all {

				if len(set) > 0 {
					for _, elem := range set {
						childSet := lty.cvt.repeatNum(elem, num, 1)
						tmp = append(tmp, childSet...)
					}
				} else {
					childSet := lty.cvt.repeatNum(item, num, 1)
					tmp = append(tmp, childSet...)
				}

			}

			set = tmp
		}

		for _, str := range set {

			key, err := lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}

	}

	return nil, success
}

// 任选
func (m *Method) CallAnyOneNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(lty, bet, p, lty.count, 1)
}

func (m *Method) CallAnyTwoNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 2, lty.count, 1)
}

func (m *Method) CallAnyThreeNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 3, lty.count, 1)
}

func (m *Method) CallAnyFourNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 4, lty.count, 1)
}

func (m *Method) CallAnyFiveNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 5, lty.count, 1)
}

func (m *Method) CallAnySixNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 5, lty.count, 1)
}

func (m *Method) CallAnySevenNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 5, lty.count, 1)
}

func (m *Method) CallAnyEightNum(lty *Lottery, bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(lty, bet, p, 5, lty.count, 1)
}

// 定单双
func (m *Method) CallOdd(lty *Lottery, bet string, p *model.Projects) (error, int) {
	bets := strings.Split(bet, ",")
	odds := lty.cvt.odd(true)
	evens := lty.cvt.odd(false)
	odds = lty.cvt.formatString(odds...)
	evens = lty.cvt.formatString(evens...)
	c := make([]string, lty.count)

	for _, b := range bets {
		set := strings.Split(b, "X")
		if len(set) != 2 {
			return errBetFormat, cancle
		}

		oddCount, _ := strconv.Atoi(set[0])
		evenCount, _ := strconv.Atoi(set[1])

		if oddCount+evenCount != lty.count {
			return errBetFormat, cancle
		}

		chooseOdds := lty.cvt.getCombination(odds, oddCount)
		chooseEvens := lty.cvt.getCombination(evens, evenCount)

		for _, chooseOddItem := range chooseOdds {
			for _, chooseEvenItem := range chooseEvens {

				copy(c[0:oddCount], chooseOddItem)
				copy(c[oddCount:], chooseEvenItem)
				perm := lty.cvt.getPermutation(c, lty.count)

				for _, str := range perm {
					key, err := lty.cvt.integer(strings.Join(str, ""))
					if err != nil {
						return err, cancle
					}

					id := p.GetInt64("projectid")
					if m.checkPunts(lty, id, key) {
						m.addRecord(lty, key, id, m.getReward(p, 1))
					}
				}
			}
		}

	}

	return nil, success
}

func (m *Method) CallMidNum(lty *Lottery, bet string, p *model.Projects) (error, int) {

	set := m.splitNum(bet)
	set = lty.cvt.formatString(set...)
	all := lty.cvt.getAllNum()
	all = lty.cvt.formatString(all...)
	com := make([]string, lty.count)
	half := (lty.count - 1) / 2

	for _, item := range set {
		var litteNums, bigNums []string

		mid, _ := strconv.Atoi(item)
		for _, num := range all {

			val, _ := strconv.Atoi(num)
			if val < mid {
				litteNums = append(litteNums, num)
			} else if val > mid {
				bigNums = append(bigNums, num)
			}

		}

		if len(litteNums) < half || len(bigNums) < half {
			return errBetFormat, cancle
		}

		copy(com[0:half], litteNums)
		com[half] = item
		copy(com[half+1:], bigNums)

		perm := lty.cvt.getPermutation(com, lty.count)
		for _, str := range perm {
			key, err := lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			id := p.GetInt64("projectid")
			if m.checkPunts(lty, id, key) {
				m.addRecord(lty, key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}
