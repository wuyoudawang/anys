package model

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"anys/pkg/utils"
	"github.com/liuzhiyi/go-db"
)

// a:3:{i:1;a:9:{s:9:"starttime";s:8:"00:00:00";
// s:12:"firstendtime";s:8:"00:05:00";s:7:"endtime";
// s:8:"01:55:00";s:4:"sort";i:1;s:5:"cycle";
// i:300;s:7:"endsale";i:90;s:13:"inputcodetime";
// i:60;s:8:"droptime";i:30;s:6:"status";i:1;}
// i:2;a:9:{s:9:"starttime";s:8:"07:00:00";
// s:12:"firstendtime";s:8:"10:00:00";s:7:"endtime";
// s:8:"22:00:00";s:4:"sort";i:2;s:5:"cycle";i:600;
// s:7:"endsale";i:140;s:13:"inputcodetime";i:60;
// s:8:"droptime";i:30;s:6:"status";i:1;}
// i:3;a:9:{s:9:"starttime";s:8:"22:00:00";
// s:12:"firstendtime";s:8:"22:05:00";s:7:"endtime";s:8:"23:55:00";
// s:4:"sort";i:3;s:5:"cycle";i:300;s:7:"endsale";i:60;s:13:"inputcodetime";
// i:60;s:8:"droptime";i:30;s:6:"status";i:1;}}

type Lottery struct {
	db.Item
}

func NewLottery() *Lottery {
	l := new(Lottery)
	l.Init("lottery", "lotteryid")
	return l
}

func (l *Lottery) GetLotteryIdByName(name string) *db.Item {
	collection := l.GetCollection()

	collection.AddFieldToFilter("m.enname", "eq", name)
	collection.Load()

	if len(collection.GetItems()) > 0 {
		return collection.GetItems()[0]
	}

	return nil
}

func (l *Lottery) GetIssueset() (rel []map[string]interface{}) {
	if l.GetId() == 0 {
		return
	}

	src := l.GetString("issueset")

	v, err := utils.PHPSerialize([]byte(src))
	if err != nil {
		return
	}

	dst := v.(map[string]interface{})
	rel = make([]map[string]interface{}, len(dst))
	sortKeys := make([]string, len(dst))
	i := 0
	for key, _ := range dst {
		pos := i
		for j := 0; j < i; j++ {
			val1, _ := strconv.Atoi(key)
			val2, _ := strconv.Atoi(sortKeys[j])
			if val1 < val2 {
				for k := i; k > j; k-- {
					sortKeys[k] = sortKeys[k-1]
				}
				pos = j
				break
			}
		}
		sortKeys[pos] = key
		i++
	}

	for i, key := range sortKeys {
		rel[i] = (dst[key]).(map[string]interface{})
	}

	return rel
}

func (l *Lottery) IssueFormat() (prefix, format string) {
	src := l.GetString("issuerule")
	reg := regexp.MustCompile(`(.*)\[n([0-9]+)\]`)
	matchs := reg.FindStringSubmatch(src)
	prefix = "20060102"
	format = "%s%d"
	if len(matchs) == 3 {
		prefix = strings.Replace(matchs[1], "Y", "2006", -1)
		prefix = strings.Replace(prefix, "m", "01", -1)
		prefix = strings.Replace(prefix, "d", "02", -1)
		prefix = strings.Replace(prefix, "H", "15", -1)
		prefix = strings.Replace(prefix, "i", "04", -1)
		prefix = strings.Replace(prefix, "s", "05", -1)
		format = "%s%0" + matchs[2] + "d"
	} else {
		fmt.Println(l.GetString("cnname"), src, "格式有错误")
	}
	return
}

func (l *Lottery) AutoGenerateIssues() error {
	now := time.Now().Add(24 * time.Hour)
	pf, format := l.IssueFormat()
	prefix := now.Format(pf)
	ise := NewIssueinfo()
	ise.SetData("lotteryid", l.GetId())
	ise.SetData("belongdate", now.Format("2006-01-02"))
	ise.Row()
	if ise.GetId() > 0 {
		return fmt.Errorf("these issues has been generated at '%s'", now.Format("2006-01-02"))
	}

	issueset := l.GetIssueset()
	transaction := l.GetResource().BeginTransaction()
	l.SetTransaction(transaction)
	defer transaction.Commit()

	id := 1

	for _, item := range issueset {
		set := item
		val, exist := set["starttime"]
		if !exist {
			goto invalidError
		}
		starttime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), val.(string)))
		if err != nil {
			goto invalidError
		}

		val, exist = set["endtime"]
		if !exist {
			goto invalidError
		}
		endtime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), val.(string)))
		if err != nil {
			goto invalidError
		}

		val, exist = set["firstendtime"]
		if !exist {
			goto invalidError
		}
		firstEndTime, err := time.Parse("2006-01-02 15:04:05", fmt.Sprintf("%s %s", now.Format("2006-01-02"), val.(string)))
		if err != nil {
			goto invalidError
		}

		val, exist = set["droptime"]
		if !exist {
			goto invalidError
		}
		droptime := val.(int64)
		droptime *= int64(time.Second)

		val, exist = set["cycle"]
		if !exist {
			goto invalidError
		}
		cycle := val.(int64)
		cycle *= int64(time.Second)

		issue := fmt.Sprintf(format, prefix, id)
		id++
		err = l.CreateCurrentIssue(issue, starttime, firstEndTime, now, droptime)
		if err != nil {
			return err
		}

		s := firstEndTime
		e := firstEndTime
		for e = s.Add(time.Duration(cycle)); e.Before(endtime) || e.Equal(endtime); e = s.Add(time.Duration(cycle)) {
			issue := fmt.Sprintf(format, prefix, id)
			err := l.CreateCurrentIssue(issue, s, e, now, droptime)
			if err != nil {
				return err
			}
			id++
			s = e
		}
	}

	return nil

invalidError:
	transaction.Rollback()
	return fmt.Errorf("invalid issue set")
}

func (l *Lottery) CreateCurrentIssue(issueNumber string, s, e, now time.Time, droptime int64) error {
	issueinfo := NewIssueinfo()
	issueinfo.SetData("issue", issueNumber)
	issueinfo.SetData("salestart", s.Format("2006-01-02 15:04:05"))
	issueinfo.SetData("saleend", e.Format("2006-01-02 15:04:05"))
	issueinfo.SetData("canneldeadline", e.Add(time.Duration(droptime)).Format("2006-01-02 15:04:05"))
	issueinfo.SetData("lotteryid", l.GetInt64("lotteryid"))
	issueinfo.SetData("belongdate", now.Format("2006-01-02"))
	issueinfo.SetTransaction(l.GetTransaction())
	return issueinfo.Save()
}

func (l *Lottery) AutoClearIssues() {
	now := time.Now()
	ago := now.Add(-90 * 24 * time.Hour).Format("2006-01-02")

	collection := NewIssueinfo().GetCollection()
	collection.AddFieldToFilter("belongdate", "lt", ago)
	collection.Load()

	for _, item := range collection.GetItems() {
		item.Delete()
	}
}

func (l *Lottery) GetLatelyIssueerrors() []*Issueerror {
	return NewIssueerror().GetLately(l.GetInt64("lotteryid"))
}

func (l *Lottery) ProcessIssueError() {
	set := l.GetLatelyIssueerrors()
	for _, isee := range set {
		isee.Process()
	}
}
