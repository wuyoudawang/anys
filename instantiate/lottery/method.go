package lottery

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"anys/instantiate/lottery/model"
	"anys/log"
	"anys/pkg/utils"
)

type Method struct {
	lty *Lottery

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

func NewMethod(l *Lottery, conf string) (*Method, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal([]byte(conf), &m)
	if err != nil {
		log.Info("config of method '%s' has an error '%s'", conf, err.Error())
		return nil, err
	}

	method := &Method{lty: l}

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

func (m *Method) addRecord(key int, id int64, reward float64) {
	m.lty.t.AddRecord(key, id, reward)
}

func (m *Method) checkPunts(id int64, key int) bool {
	n := m.lty.t.Get(key)
	if n == nil {
		return true
	}

	currPunts := n.getRecordCurrPunts(id)
	return currPunts < m.maxPunts
}

// 定位胆
func (m *Method) CallOneNum(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	for i, item := range set {

		if item != placeHolder {

			nums := m.splitNum(item)
			for _, num := range nums {

				key, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				key = m.lty.cvt.shiftLeft(key, i+1)

				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}

			}
			return nil, success
		}
	}

	return errBetFormat, cancle
}

func (m *Method) twoNum(bet string, p *model.Projects, frist, second int) (error, int) {
	set := m.splitBit(bet)
	one := set[frist-1]
	two := set[second-1]

	if one == placeHolder || two == placeHolder {
		return errBetFormat, cancle
	}

	var key int
	var err error
	twoPart := m.splitNum(two)
	for _, num := range m.splitNum(one) {
		key, err = m.lty.cvt.getFlag(num)
		if err != nil {
			return err, cancle
		}

		key = m.lty.cvt.shiftLeft(key, frist)

		for _, n := range twoPart {
			k, err := m.lty.cvt.getFlag(n)
			if err != nil {
				return err, cancle
			}

			k = m.lty.cvt.shiftLeft(k, second)

			k |= key

			id := p.GetInt64("projectid")
			if m.checkPunts(id, k) {
				m.addRecord(k, id, m.getReward(p, 1))
			}

		}
	}

	return nil, success
}

// 前二直选
func (m *Method) CallPrevTwoNum(bet string, p *model.Projects) (error, int) {
	return m.twoNum(bet, p, 1, 2)
}

// 后二直选
func (m *Method) CallLastTwoNum(bet string, p *model.Projects) (error, int) {
	first := m.lty.getLast(2)
	second := first + 1
	return m.twoNum(bet, p, first, second)
}

func (m *Method) threeNum(bet string, p *model.Projects, frist, second, third int) (error, int) {
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
		key, err = m.lty.cvt.getFlag(num)
		if err != nil {
			return err, cancle
		}

		key = m.lty.cvt.shiftLeft(key, frist)

		for _, num := range twoPart {
			k, err := m.lty.cvt.getFlag(num)
			if err != nil {
				return err, cancle
			}

			k = m.lty.cvt.shiftLeft(k, second)

			k |= key

			for _, num := range threePart {
				tk, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return err, cancle
				}

				tk = m.lty.cvt.shiftLeft(tk, third)

				tk |= k
				id := p.GetInt64("projectid")
				if m.checkPunts(id, tk) {
					m.addRecord(tk, id, m.getReward(p, 1))
				}
			}

		}
	}

	return nil, success
}

// 前三直选
func (m *Method) CallPrevThreeNum(bet string, p *model.Projects) (error, int) {
	return m.threeNum(bet, p, 1, 2, 3)
}

// 后三直选
func (m *Method) CallLastThreeNum(bet string, p *model.Projects) (error, int) {
	f := m.lty.getLast(3)
	s := f + 1
	t := s + 1

	if m.lty.count < t {
		return errBetFormat, cancle
	}

	return m.threeNum(bet, p, f, s, t)
}

// 中三直选
func (m *Method) CallMidThreeNum(bet string, p *model.Projects) (error, int) {
	f := m.lty.getMid(3)
	s := f + 1
	t := s + 1

	if m.lty.count < t {
		return errBetFormat, cancle
	}

	return m.threeNum(bet, p, f, s, t)
}

/**
*五星直选
***/
func (m *Method) CallAllNum(bet string, p *model.Projects) (error, int) {
	return m.anyNum(bet, p)
}

func (m *Method) getKeys(bet string, pos ...int) ([]int, error) {
	var keys, prev []int
	var err error

	if len(pos) == 0 {
		if strings.Contains(bet, placeHolder) {
			return keys, errBetFormat
		}

		set := m.splitBit(bet)
		for i := 1; i <= m.lty.count; i++ {

			keys = []int{}
			for _, num := range m.splitNum(set[i-1]) {
				key, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = m.lty.cvt.shiftLeft(key, i)
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

			if len(set) == m.lty.count {
				h = i - 1
			} else {
				h = j
			}

			if i > m.lty.count || set[h] == placeHolder {
				return keys, errBetFormat
			}
		}

		for j, i := range pos {

			if len(set) == m.lty.count {
				h = i - 1
			} else {
				h = j
			}

			keys = []int{}
			for _, num := range m.splitNum(set[h]) {
				key, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = m.lty.cvt.shiftLeft(key, i)
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

func (m *Method) anyNum(bet string, p *model.Projects, pos ...int) (error, int) {
	keys, err := m.getKeys(bet, pos...)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		id := p.GetInt64("projectid")
		if m.checkPunts(id, key) {
			m.addRecord(key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*五星组合
***/
func (m *Method) CallAllComNum(bet string, p *model.Projects) (error, int) {

	keys, err := m.getKeys(bet)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		for i := 1; i <= m.lty.count; i++ {
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {

				m.addRecord(key, id, m.getReward(p, 1/math.Pow10(i-1)))
			}

			key = key & int(utils.NOT(int64(m.lty.cvt.shiftLeft(mask, i))))
		}
	}

	return nil, success
}

/**
*五星组选120
***/
func (m *Method) CallAll120Num(bet string, p *model.Projects) (error, int) {

	if strings.Contains(bet, placeHolder) {
		return errBetFormat, cancle
	}

	set := m.splitNum(bet)
	set = m.lty.cvt.formatString(set...)
	permutation := m.lty.cvt.getPermutation(set, 5)

	for _, s := range permutation {
		key, err := m.lty.cvt.integer(strings.Join(s, ""))
		if err != nil {
			return err, cancle
		}

		id := p.GetInt64("projectid")
		if m.checkPunts(id, key) {
			m.addRecord(key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*五星组选60 10*(9*8*7/3*2)
***/
func (m *Method) CallAll60Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = m.lty.cvt.formatString(singlePart...)
	singleSet := m.lty.cvt.getPermutation(singlePart, 3)
	doubleNums := m.splitNum(set[0])
	doubleNums = m.lty.cvt.formatString(doubleNums...)

	for _, num := range doubleNums {
		for _, nums := range singleSet {

			permutation := m.lty.cvt.repeatNum(nums, num, 2)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选30
***/
func (m *Method) CallAll30Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = m.lty.cvt.formatString(singlePart...)
	doubleNums := m.splitNum(set[0])
	doubleNums = m.lty.cvt.formatString(doubleNums...)
	com := m.lty.cvt.getCombination(doubleNums, 2)
	if len(com) == 0 {
		return errBetFormat, cancle
	}

	for _, singleNum := range singlePart {
		for _, elem := range com {
			set := m.lty.cvt.repeatNum([]string{singleNum}, elem[0], 2)

			for _, item := range set {
				data := m.lty.cvt.repeatNum(item, elem[1], 2)

				for _, s := range data {
					key, err := m.lty.cvt.integer(strings.Join(s, ""))
					if err != nil {
						return err, cancle
					}

					id := p.GetInt64("projectid")
					if m.checkPunts(id, key) {
						m.addRecord(key, id, m.getReward(p, 1))
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
func (m *Method) CallAll20Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = m.lty.cvt.formatString(singlePart...)
	singleSet := m.lty.cvt.getPermutation(singlePart, 2)
	trebleNums := m.splitNum(set[0])
	trebleNums = m.lty.cvt.formatString(trebleNums...)

	for _, num := range trebleNums {
		for _, nums := range singleSet {

			permutation := m.lty.cvt.repeatNum(nums, num, 3)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选10
***/
func (m *Method) CallAll10Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	doubleNums := m.splitNum(set[1])
	doubleNums = m.lty.cvt.formatString(doubleNums...)
	trebleNums := m.splitNum(set[0])
	trebleNums = m.lty.cvt.formatString(trebleNums...)

	for _, doubleNum := range doubleNums {
		for _, trebleNum := range trebleNums {

			permutation := m.lty.cvt.repeatNum([]string{doubleNum, doubleNum}, trebleNum, 3)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*五星选5
***/
func (m *Method) CallAll5Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singleNums := m.splitNum(set[1])
	singleNums = m.lty.cvt.formatString(singleNums...)
	fourfoldNums := m.splitNum(set[0])
	fourfoldNums = m.lty.cvt.formatString(fourfoldNums...)

	for _, fourfoldNum := range fourfoldNums {
		for _, singleNum := range singleNums {

			set := []string{fourfoldNum, fourfoldNum, fourfoldNum, fourfoldNum}
			permutation := m.lty.cvt.repeatNum(set, singleNum, 1)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*四星直选
***/
func (m *Method) CallFourNum(bet string, p *model.Projects) (error, int) {
	return m.anyNum(bet, p, 2, 3, 4, 5)
}

/**
*四星组合
***/
func (m *Method) CallFourComNum(bet string, p *model.Projects) (error, int) {
	keys, err := m.getKeys(bet, 2, 3, 4, 5)
	if err != nil {
		return err, cancle
	}

	for _, key := range keys {
		for i := 2; i <= m.lty.count; i++ {
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1/math.Pow10(i-2)))
			}

			key = key & int(utils.NOT(int64(m.lty.cvt.shiftLeft(mask, i))))
		}
	}

	return nil, success
}

/**
*四星组选24
***/
func (m *Method) CallFour24Num(bet string, p *model.Projects) (error, int) {

	if strings.Contains(bet, placeHolder) {
		return errBetFormat, cancle
	}

	set := m.splitNum(bet)
	set = m.lty.cvt.formatString(set...)
	permutation := m.lty.cvt.getPermutation(set, 4)
	for _, s := range permutation {

		key, err := m.lty.cvt.integer(strings.Join(s, ""))
		if err != nil {
			return err, cancle
		}

		key = m.lty.cvt.shiftLeft(key, 2)
		id := p.GetInt64("projectid")
		if m.checkPunts(id, key) {
			m.addRecord(key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/**
*四星组选12
***/
func (m *Method) CallFour12Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = m.lty.cvt.formatString(singlePart...)
	singleSet := m.lty.cvt.getPermutation(singlePart, 2)
	doubleNums := m.splitNum(set[0])
	doubleNums = m.lty.cvt.formatString(doubleNums...)

	for _, num := range doubleNums {
		for _, nums := range singleSet {

			permutation := m.lty.cvt.repeatNum(nums, num, 2)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, 2)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*四星组选6
***/
func (m *Method) CallFour6Num(bet string, p *model.Projects) (error, int) {

	doubleNums := m.splitNum(bet)
	doubleNums = m.lty.cvt.formatString(doubleNums...)
	com := m.lty.cvt.getCombination(doubleNums, 2)

	for _, item := range com {

		permutation := m.lty.cvt.repeatNum([]string{item[0], item[0]}, item[1], 2)
		for _, s := range permutation {
			key, err := m.lty.cvt.integer(strings.Join(s, ""))
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, 2)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/**
*四星组选4
***/
func (m *Method) CallFour4Num(bet string, p *model.Projects) (error, int) {
	set := m.splitBit(bet)
	if len(set) != 2 {
		return errBetFormat, cancle
	}

	singlePart := m.splitNum(set[1])
	singlePart = m.lty.cvt.formatString(singlePart...)
	trebleNums := m.splitNum(set[0])
	trebleNums = m.lty.cvt.formatString(trebleNums...)

	for _, trebleNum := range trebleNums {
		for _, singleNum := range singlePart {

			permutation := m.lty.cvt.repeatNum([]string{singleNum}, trebleNum, 3)
			for _, s := range permutation {
				key, err := m.lty.cvt.integer(strings.Join(s, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, 2)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

func (m *Method) anyOneNum(bet string, p *model.Projects, n, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)
	all := m.lty.cvt.getAllNum()
	all = m.lty.cvt.formatString(all...)
	permu := m.lty.cvt.getSelection(all, n-1)

	for _, num := range nums {

		for _, item := range permu {

			set := m.lty.cvt.repeatNum(item, num, 1)

			for _, str := range set {
				key, err := m.lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/**
*前三(不定胆)
***/
func (m *Method) CallPrevThreeOneNum(bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(bet, p, 3, 1)
}

/**
*中三(不定胆)
***/
func (m *Method) CallMidThreeOneNum(bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(bet, p, 3, m.lty.getMid(2))
}

/**
*后三(不定胆)
***/
func (m *Method) CallLastThreeOneNum(bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(bet, p, 3, m.lty.getLast(3))
}

func (m *Method) anyAnyNum(bet string, p *model.Projects, r, n, s int) (error, int) {
	if n < r {
		return fmt.Errorf("args 'n(%d)'' should greater than %d", n, r), cancle
	}

	if n > m.lty.count {
		return fmt.Errorf("overflow lottery's max bit"), cancle
	}

	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)
	permu := m.lty.cvt.getPermutation(nums, r)
	all := m.lty.cvt.getAllNum()
	all = m.lty.cvt.formatString(all...)

	// for _, item := range permu {

	// 	for i := 0; i < m.lty.count; i++ {

	// 		num := m.lty.cvt.formatInt(m.lty.cvt.start + i)[0]
	// 		set := m.lty.cvt.repeatNum(item, num, n-r)

	// 		for _, str := range set {
	// 			key, err := m.lty.cvt.integer(strings.Join(str, ""))
	// 			if err != nil {
	// 				return err, cancle
	// 			}

	// 			key = m.lty.cvt.shiftLeft(key, s)
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
						childSet := m.lty.cvt.repeatNum(elem, num, 1)
						tmp = append(tmp, childSet...)
					}
				} else {
					childSet := m.lty.cvt.repeatNum(item, num, 1)
					tmp = append(tmp, childSet...)
				}

			}

			set = tmp
		}

		for _, str := range set {

			key, err := m.lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}

	}

	return nil, success
}

/**
*前三二字不定胆
***/
func (m *Method) CallPrevThreeTwoNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 2, 3, 1)
}

/**
*中三二字不定胆
***/
func (m *Method) CallMidThreeTwoNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 2, 3, m.lty.getMid(3))
}

/**
*后三二字不定胆
***/
func (m *Method) CallLastThreeTwoNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 2, 3, m.lty.getLast(3))
}

func (m *Method) threeComNum(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)

	com := m.lty.cvt.getCombination(nums, 2)
	for _, item := range com {
		for i := 0; i < 2; i++ {
			var set [][]string
			if i == 1 {
				set = m.lty.cvt.repeatNum([]string{item[i], item[i]}, item[0], 1)
			} else {
				set = m.lty.cvt.repeatNum([]string{item[i], item[i]}, item[1], 1)
			}

			for _, str := range set {
				key, err := m.lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}

	}

	return nil, success
}

/***
*组选3(前三，中三，后三)
***/
func (m *Method) CallPrevThreeComNum(bet string, p *model.Projects) (error, int) {
	return m.threeComNum(bet, p, 1)
}

func (m *Method) CallMidThreeComNum(bet string, p *model.Projects) (error, int) {
	return m.threeComNum(bet, p, m.lty.getLast(3))
}

func (m *Method) CallLastThreeComNum(bet string, p *model.Projects) (error, int) {
	return m.threeComNum(bet, p, m.lty.getLast(3))
}

func (m *Method) sixComNum(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)

	permu := m.lty.cvt.getPermutation(nums, 3)
	for _, str := range permu {
		key, err := m.lty.cvt.integer(strings.Join(str, ""))
		if err != nil {
			return err, cancle
		}

		key = m.lty.cvt.shiftLeft(key, s)
		id := p.GetInt64("projectid")
		if m.checkPunts(id, key) {
			m.addRecord(key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/***
*组选6(前三，中三，后三)
***/
func (m *Method) CallPrevSixComNum(bet string, p *model.Projects) (error, int) {
	return m.sixComNum(bet, p, 1)
}

func (m *Method) CallMidSixComNum(bet string, p *model.Projects) (error, int) {
	return m.sixComNum(bet, p, m.lty.getMid(3))
}

func (m *Method) CallLastSixComNum(bet string, p *model.Projects) (error, int) {
	return m.sixComNum(bet, p, m.lty.getLast(3))
}

func (m *Method) twoComNum(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)

	permu := m.lty.cvt.getPermutation(nums, 2)
	for _, str := range permu {
		key, err := m.lty.cvt.integer(strings.Join(str, ""))
		if err != nil {
			return err, cancle
		}

		key = m.lty.cvt.shiftLeft(key, s)
		id := p.GetInt64("projectid")
		if m.checkPunts(id, key) {
			m.addRecord(key, id, m.getReward(p, 1))
		}
	}

	return nil, success
}

/***
*二星组选(前二， 后二)
***/
func (m *Method) CallPrevTwoTwoNum(bet string, p *model.Projects) (error, int) {
	return m.twoComNum(bet, p, 1)
}

func (m *Method) CallLastTwoTwoNum(bet string, p *model.Projects) (error, int) {
	return m.twoComNum(bet, p, m.lty.getLast(2))
}

func (m *Method) twoSizeOdd(bet string, p *model.Projects, s int) (error, int) {
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
				tmp = m.lty.cvt.size(true)
			case "1": // 小
				tmp = m.lty.cvt.size(false)
			case "2": // 单
				tmp = m.lty.cvt.odd(true)
			case "3": // 双
				tmp = m.lty.cvt.odd(false)
			}

			if i == 0 {
				part1 = append(part1, tmp...)
			} else {
				part2 = append(part2, tmp...)
			}
		}
	}

	part1 = m.lty.cvt.formatString(part1...)
	part2 = m.lty.cvt.formatString(part2...)

	for _, num1 := range part1 {

		for _, num2 := range part2 {
			num2 = num1 + num2

			key, err := m.lty.cvt.integer(num2)
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/***
*二星大小单双(前二， 后二)
***/
func (m *Method) CallPrevTwoSizeOdd(bet string, p *model.Projects) (error, int) {
	return m.twoSizeOdd(bet, p, 1)
}

func (m *Method) CallLastTwoSizeOdd(bet string, p *model.Projects) (error, int) {
	return m.twoSizeOdd(bet, p, m.lty.getLast(2))
}

func (m *Method) threeSumCom(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := m.lty.cvt.getSumCom(v, 3)

		for _, str := range com {

			if str[0] != str[1] && str[0] != str[2] && str[1] != str[2] { // 组六
				key, err := m.lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			} else if str[0] != str[1] || str[0] != str[2] || str[1] != str[2] { // 组三
				key, err := m.lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.rate /= 2
					m.addRecord(key, id, m.getReward(p, 1))
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
func (m *Method) CallPrevThreeSumCom(bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(bet, p, 1)
}

func (m *Method) CallMidThreeSumCom(bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(bet, p, m.lty.getMid(3))
}

func (m *Method) CallLastThreeSumCom(bet string, p *model.Projects) (error, int) {
	return m.threeSumCom(bet, p, m.lty.getLast(3))
}

/***
*前三直选和值
***/
func (m *Method) CallPrevThreeSum(bet string, p *model.Projects) (error, int) {
	return m.threeSum(bet, p, 1)
}

/**
* 中三直选和值
 */
func (m *Method) CallMidThreeSum(bet string, p *model.Projects) (error, int) {
	return m.threeSum(bet, p, m.lty.getMid(3))
}

/**
* 后三直选和值
 */
func (m *Method) CallLastThreeSum(bet string, p *model.Projects) (error, int) {
	return m.threeSum(bet, p, m.lty.getLast(3))
}

func (m *Method) threeSum(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := m.lty.cvt.getSumCom(v, 3)

		for _, str := range com {

			key, err := m.lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

func (m *Method) twoSum(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)
	nums = m.lty.cvt.formatString(nums...)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := m.lty.cvt.getSumCom(v, 2)

		for _, str := range com {

			key, err := m.lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}

/***
*二字直选选和
***/
func (m *Method) CallPrevTwoSum(bet string, p *model.Projects) (error, int) {
	return m.twoSum(bet, p, 1)
}

func (m *Method) CallLastTwoSum(bet string, p *model.Projects) (error, int) {
	return m.twoSum(bet, p, m.lty.getLast(2))
}

func (m *Method) twoSumCom(bet string, p *model.Projects, s int) (error, int) {
	nums := m.splitNum(bet)

	for _, num := range nums {
		v, _ := strconv.Atoi(num)
		com := m.lty.cvt.getSumCom(v, 2)

		for _, str := range com {

			if str[0] != str[1] {
				key, err := m.lty.cvt.integer(strings.Join(str, ""))
				if err != nil {
					return err, cancle
				}

				key = m.lty.cvt.shiftLeft(key, s)
				id := p.GetInt64("projectid")
				if m.checkPunts(id, key) {
					m.addRecord(key, id, m.getReward(p, 1))
				}
			}
		}
	}

	return nil, success
}

/***
*二字和数(二字组选和)
***/
func (m *Method) CallPrevTwoSumCom(bet string, p *model.Projects) (error, int) {
	return m.twoSumCom(bet, p, 1)
}

func (m *Method) CallLastTwoSumCom(bet string, p *model.Projects) (error, int) {
	return m.twoSumCom(bet, p, m.lty.getLast(2))
}

func (m *Method) mixThreeNum(bet string, p *model.Projects, s int) (error, int) {
	key, err := m.lty.cvt.integer(bet)
	if err != nil {
		return err, cancle
	}

	key = m.lty.cvt.shiftLeft(key, s)
	id := p.GetInt64("projectid")
	if m.checkPunts(id, key) {
		m.addRecord(key, id, m.getReward(p, 1))
	}

	return nil, success
}

/***
*混合组选
***/
func (m *Method) CallPrevMixThreeNum(bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(bet, p, 1)
}

func (m *Method) CallMidMixThreeNum(bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(bet, p, m.lty.getMid(3))
}

func (m *Method) CallLastMixThreeNum(bet string, p *model.Projects) (error, int) {
	return m.mixThreeNum(bet, p, m.lty.getLast(3))
}

/***
*趣味玩法 一帆风顺
***/
func (m *Method) CallEverythingGood(bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(bet, p, 5, 1)
}

/***
*趣味玩法 好事成双
***/
func (m *Method) CallPairGood(bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(bet, p, 2, 5, 1)
}

/***
*趣味玩法 三星报喜
***/
func (m *Method) CallThreeGood(bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(bet, p, 3, 5, 1)
}

/***
*趣味玩法 四季发财
***/
func (m *Method) CallFourGood(bet string, p *model.Projects) (error, int) {
	return m.anySingleRepeatNum(bet, p, 4, 5, 1)
}

func (m *Method) anySingleRepeatNum(bet string, p *model.Projects, r, n, s int) (error, int) {
	if n < r {
		return fmt.Errorf("args 'n(%d)' should greater than %d", n, r), cancle
	}

	if n > m.lty.count {
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

	all := m.lty.cvt.getAllNum()

	for _, item := range permu {

		var set [][]string
		for i := 0; i < n-r; i++ {

			var tmp [][]string
			for _, num := range all {

				if len(set) > 0 {
					for _, elem := range set {
						childSet := m.lty.cvt.repeatNum(elem, num, 1)
						tmp = append(tmp, childSet...)
					}
				} else {
					childSet := m.lty.cvt.repeatNum(item, num, 1)
					tmp = append(tmp, childSet...)
				}

			}

			set = tmp
		}

		for _, str := range set {

			key, err := m.lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			key = m.lty.cvt.shiftLeft(key, s)
			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}

	}

	return nil, success
}

// 任选
func (m *Method) CallAnyOneNum(bet string, p *model.Projects) (error, int) {
	return m.anyOneNum(bet, p, m.lty.count, 1)
}

func (m *Method) CallAnyTwoNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 2, m.lty.count, 1)
}

func (m *Method) CallAnyThreeNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 3, m.lty.count, 1)
}

func (m *Method) CallAnyFourNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 4, m.lty.count, 1)
}

func (m *Method) CallAnyFiveNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 5, m.lty.count, 1)
}

func (m *Method) CallAnySixNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 5, m.lty.count, 1)
}

func (m *Method) CallAnySevenNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 5, m.lty.count, 1)
}

func (m *Method) CallAnyEightNum(bet string, p *model.Projects) (error, int) {
	return m.anyAnyNum(bet, p, 5, m.lty.count, 1)
}

// 定单双
func (m *Method) CallOdd(bet string, p *model.Projects) (error, int) {
	bets := strings.Split(bet, ",")
	odds := m.lty.cvt.odd(true)
	evens := m.lty.cvt.odd(false)
	odds = m.lty.cvt.formatString(odds...)
	evens = m.lty.cvt.formatString(evens...)
	c := make([]string, m.lty.count)

	for _, b := range bets {
		set := strings.Split(b, "X")
		if len(set) != 2 {
			return errBetFormat, cancle
		}

		oddCount, _ := strconv.Atoi(set[0])
		evenCount, _ := strconv.Atoi(set[1])

		if oddCount+evenCount != m.lty.count {
			return errBetFormat, cancle
		}

		chooseOdds := m.lty.cvt.getCombination(odds, oddCount)
		chooseEvens := m.lty.cvt.getCombination(evens, evenCount)

		for _, chooseOddItem := range chooseOdds {
			for _, chooseEvenItem := range chooseEvens {

				copy(c[0:oddCount], chooseOddItem)
				copy(c[oddCount:], chooseEvenItem)
				perm := m.lty.cvt.getPermutation(c, m.lty.count)

				for _, str := range perm {
					key, err := m.lty.cvt.integer(strings.Join(str, ""))
					if err != nil {
						return err, cancle
					}

					id := p.GetInt64("projectid")
					if m.checkPunts(id, key) {
						m.addRecord(key, id, m.getReward(p, 1))
					}
				}
			}
		}

	}

	return nil, success
}

func (m *Method) CallMidNum(bet string, p *model.Projects) (error, int) {

	set := m.splitNum(bet)
	set = m.lty.cvt.formatString(set...)
	all := m.lty.cvt.getAllNum()
	all = m.lty.cvt.formatString(all...)
	com := make([]string, m.lty.count)
	half := (m.lty.count - 1) / 2

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

		perm := m.lty.cvt.getPermutation(com, m.lty.count)
		for _, str := range perm {
			key, err := m.lty.cvt.integer(strings.Join(str, ""))
			if err != nil {
				return err, cancle
			}

			id := p.GetInt64("projectid")
			if m.checkPunts(id, key) {
				m.addRecord(key, id, m.getReward(p, 1))
			}
		}
	}

	return nil, success
}
