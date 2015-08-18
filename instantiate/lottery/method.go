package lottery

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"anys/instantiate/lottery/model"
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

	if maxPunts, exists := m["combinations"].(float64); !exists {
		method.maxPunts = 1
	} else {
		method.maxPunts = int64(maxPunts)
	}

	if method.rate, exists = m["rate"].(float64); !exists {
		method.rate = 1.0
	}

	return method, nil
}

func (m *Method) getReward(p *model.Projects, bets int64) float64 {
	reward := p.GetFloat64("omodel") /
		float64(1000*p.GetMode()) *
		float64(m.combinations*bets*p.GetInt64("multiple"))

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
	return m.twoNum(bet, p, 4, 5)
}

func (m *Method) threeNum(bet string, p *model.Projects, frist, second, third int) (error, int) {
	set := m.splitBit(bet)
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
	return m.threeNum(bet, p, 3, 4, 5)
}

// 中三直选
func (m *Method) CallMidThreeNum(bet string, p *model.Projects) (error, int) {
	return m.threeNum(bet, p, 2, 3, 4)
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

	if len(pos) < 0 {
		if strings.Contains(bet, placeHolder) {
			return keys, errBetFormat
		}

		set := m.splitBit(bet)
		for _, i := range pos {

			keys = []int{}
			for _, num := range m.splitNum(set[i]) {
				key, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = m.lty.cvt.shiftLeft(key, i+1)
				if len(prev) > 0 {
					for _, k := range keys {
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
		set := m.splitNum(bet)
		for _, i := range pos {
			if i > m.lty.count || set[i] == placeHolder {
				return keys, errBetFormat
			}
		}

		for _, i := range pos {

			keys = []int{}
			for _, num := range m.splitNum(set[i]) {
				key, err := m.lty.cvt.getFlag(num)
				if err != nil {
					return keys, err
				}

				key = m.lty.cvt.shiftLeft(key, i+1)
				if len(prev) > 0 {
					for _, k := range keys {
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
				m.addRecord(key, id, m.getReward(p, 1/int64(math.Pow10(i-1))))
			}

			key = key & utils.NOT(m.lty.cvt.shiftLeft(mask, i))
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
			set = m.lty.cvt.repeatNum(set, elem[1], 2)
			for _, s := range set {
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
				m.addRecord(key, id, m.getReward(p, 1/int64(math.Pow10(i-2))))
			}

			key = key & utils.NOT(m.lty.cvt.shiftLeft(mask, i))
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
	permutation := m.lty.cvt.getPermutation(set, 5)
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
	p.Combinations = 360
	set := strings.Split(p.Number, "|")
	if len(set) != 2 || set[0] == "*" {
		return settlement.ExcNumErr
	}
	numsStr := strings.Split(set[0], ",")
	if len(numsStr) < 1 {
		return settlement.ExcNumErr
	}
	for _, numStr := range numsStr {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return err
		}
		tmp := settlement.ArrayDel(t.code[1:], num)
		if len(tmp) == 2 {
			if strings.Contains(set[1], fmt.Sprintf("%d", tmp[0])) && strings.Contains(set[1], fmt.Sprintf("%d", tmp[1])) {
				p.Bets = 1
			}
		}
	}

	return nil
}

/**
*四星组选6
***/
func (t *tjssc) CallFour6Numerical(p *settlement.Project) error {
	p.Combinations = 45
	numsStr := strings.Split(p.Number, ",")
	if len(numsStr) < 2 {
		return settlement.ExcNumErr
	}
	tmp := t.code[1:]
	for _, numStr := range numsStr {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return err
		}
		tmp := settlement.ArrayDel(tmp, num)
		if len(tmp) == 0 {
			p.Bets = 1
		}
	}

	return nil
}

/**
*四星组选4
***/
func (t *tjssc) CallFour4Numerical(p *settlement.Project) error {
	p.Combinations = 90
	set := strings.Split(p.Number, "|")
	if len(set) != 2 || set[0] == "*" {
		return settlement.ExcNumErr
	}
	numsStr := strings.Split(set[0], ",")
	if len(numsStr) < 1 {
		return settlement.ExcNumErr
	}
	for _, numStr := range numsStr {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			return err
		}
		tmp := settlement.ArrayDel(t.code[1:], num)
		if len(tmp) == 1 {
			if strings.Contains(set[1], fmt.Sprintf("%d", tmp[0])) {
				p.Bets = 1
			}
		}
	}

	return nil
}
