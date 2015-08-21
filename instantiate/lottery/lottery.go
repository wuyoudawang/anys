package lottery

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"anys/instantiate/lottery/model"
	"anys/log"
	"anys/pkg/utils"
)

type methodFn func(string, *model.Projects) (error, int)

type Lottery struct {
	gross     float64
	hedging   float64
	maxReward float64
	count     int
	cvt       *convert
	t         *Table
	nums      []string
	methods   map[string]methodFn
}

const (
	betSeparator = ";"
	bitSeparator = "|"
	numSeparator = ","
	placeHolder  = "*"
)

func NewLottery(a, b, c int) *Lottery {
	l := &Lottery{}

	l.cvt = newConvert(a, b)
	l.t = NewTable(l, c)
	l.methods = make(map[string]methodFn)

	return l
}

func (l *Lottery) splitBet(content string) []string {
	return strings.Split(content, betSeparator)
}

func (l *Lottery) NewNumber(str string) (*Number, error) {
	key, err := l.cvt.integer(str)
	if err != nil {
		return nil, err
	}

	i := l.t.Hash(key)
	n := &Number{
		bets:  str,
		key:   key,
		index: i,
	}

	n.records = make(map[int64]*record)

	return n, nil
}

func (l *Lottery) addGross(v float64) {
	l.gross += v
}

func (l *Lottery) GetLen() int {
	return l.count * l.cvt.terms
}

func (l *Lottery) GetGross() float64 {
	return l.gross
}

func (l *Lottery) GetMaxReward() float64 {
	if l.maxReward == 0 {
		l.maxReward = l.gross - l.gross*l.hedging
	}
	return l.maxReward
}

func (l *Lottery) GenerateKey(num string) (int, error) {
	return l.cvt.integer(num)
}

func (l *Lottery) addNum(num string) {
	l.nums = append(l.nums, num)
}

func (l *Lottery) Register(name string, fn methodFn) error {
	_, exists := l.methods[name]
	if exists {
		return fmt.Errorf("method '%s' have already existed", name)
	}

	l.methods[name] = fn
	return nil
}

func (l *Lottery) Dispatch(p *model.Projects) error {

	name := utils.UcWords(p.GetString("functionname"))
	fn, exists := l.methods[name]
	if !exists {
		return fmt.Errorf("call a non-existent method '%s'", name)
	}

	code := strings.Trim(p.GetString("code"), ";")
	bets := l.splitBet(code)

	for _, bet := range bets {

		err, flag := fn(bet, p)
		if err != nil {
			log.Warning("project(id '%d', funcname '%s', bet '%s') has an error '%s'",
				p.GetId(), name, bet, err.Error())
		}

		if flag == done {
			break
		}

		if flag == cancle {

		}
	}

	l.addGross(p.GetFloat64("totalprice"))

	return nil
}

// choose a random number
func (l *Lottery) Draw() string {

	i := 0
	if len(l.nums) > 1 {
		rand.Seed(time.Now().UnixNano())
		i = rand.Intn(len(l.nums))
	}

	return l.nums[i]
}

func (l *Lottery) GetRecords(key int) (set []*record) {
	var n *Number
	for i := 1; i <= l.count; i++ {
		keys := l.cvt.getChildrenKey(key, i)

		for _, key := range keys {
			n = l.t.Get(key)

			if n != nil {
				for _, r := range n.records {
					set = append(set, r)
				}
			}

		}
	}

	return
}

func (l *Lottery) SendReward() {}
