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
	id        int64
	name      string
	gross     float64
	hedging   float64
	maxReward float64
	count     int
	cvt       *convert
	t         *Table
	nums      []string
	curIssue  *model.Issueinfo
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

func (l *Lottery) setId(id int64) {
	l.id = id
}

func (l *Lottery) getLast(n int) int {
	return l.count - n + 1
}

func (l *Lottery) getMid(n int) int {
	return (l.count-n)/2 + 1
}

func (l *Lottery) GetId() int64 {
	return l.id
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

func (l *Lottery) addNums(nums ...string) {
	l.nums = append(l.nums, nums...)
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

	name := utils.UcWords(p.GetString("customname"))
	fn, exists := l.methods[name]
	if !exists {
		return fmt.Errorf("call a non-existent method '%s'", name)
	}

	code := strings.Trim(p.GetString("code"), betSeparator)
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

func (l *Lottery) Clone() (*Lottery, error) {
	lty := &Lottery{}

	lty.cvt = l.cvt
	lty.t = l.t.Clone(lty)
	lty.methods = l.methods

	return lty, nil
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

func (l *Lottery) Persist(winNum string) error {
	return l.curIssue.FinishDraw(winNum)
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

func (l *Lottery) Reset() {
	l.gross = 0
	l.maxReward = 0
	l.nums = []string{}
	l.curIssue = nil
	l.t.reset()
}

func (l *Lottery) Task() error {
	return l.curIssue.Task()
}

func (l *Lottery) Spread() {
	set := l.curIssue.GetCurrentProjects()
	for _, item := range set {
		err := l.Dispatch(item)

		if err != nil {
			log.Error("project(%d) has an error '%s'", item.GetId(), err.Error())

			err = item.Cancel()
			if err != nil {
				log.Error("cancel the project(%d) is not successful with an error '%s' ", item.GetId(), err.Error())
			}
		}
	}
}

func (l *Lottery) SendReward(key int) error {
	records := l.GetRecords(key)
	project := model.NewProjects()
	for _, r := range records {
		project.Load(int(r.projectid))

		err := project.Reward(r.amount)
		if err != nil {
			log.Error("An error occurred during the project(%d) be sent reward: '%s'", r.projectid, err.Error())
		}

		if project.GetInt64("taskid") > 0 {

			err = project.FlushTask()
			if err != nil {
				log.Error("An error occurred while freshing the task(%d): '%s'", project.GetInt64("taskid"), err.Error())
			}
		}

	}

	set := l.curIssue.GetCurrentProjects()
	for _, item := range set {
		err := item.Unreward()
		if err != nil {
			log.Error("the project(%d) has an error:%s", item.GetId(), err.Error())
		}

		if project.GetInt64("taskid") > 0 {
			err = project.FlushTask()
			if err != nil {
				log.Error("An error occurred while freshing the task(%d): '%s'", project.GetInt64("taskid"), err.Error())
			}
		}
	}

	return l.curIssue.FinishSendReward()
}

func (l *Lottery) Process() {
	var err error

	l.curIssue = model.GetCurrentIssue(l.id)
	if l.curIssue == nil {
		return
	}

	err = l.Task()
	if err != nil {
		log.Error("during the tasking has an error '%s'", err.Error())
		return
	}

	l.Spread()
	l.Reduce()
	winNum := l.Draw()
	key, err := l.GenerateKey(winNum)
	if err != nil {
		log.Error("the winNum('%s') is invaild ", winNum)
		return
	}

	err = l.Persist(winNum)
	if err != nil {
		log.Error("persist the winNum has an error '%s'", err.Error())
		return
	}

	l.SendReward(key)
}
