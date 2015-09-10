package model

import (
	"fmt"
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
// s:12:"firstendtime";s:8:"22:05:00";s:7:"endtime";s:8:"00:00:00";
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

func (l *Lottery) GetIssueset() (dst map[string]map[string]interface{}) {
	if l.GetId() == 0 {
		return
	}

	src := l.GetString("issueset")
	v, err := utils.PHPSerialize([]byte(src))
	if err != nil {
		return
	}

	dst = v.(map[string]map[string]interface{})
	return
}

func (l *Lottery) AutoGenerateIssues() error {
	now := time.Now()
	prefix := now.Format("20060102")
	date, _ := time.Parse("20060102", prefix)
	issueset := l.GetIssueset()
	transaction := l.GetResource().BeginTransaction()

	id := 1
	for i := 1; i <= len(issueset); i++ {
		index := fmt.Sprintf("%d", i)
		set := issueset[index]
		val, exist := set["starttime"]
		if !exist {
			goto invalidError
		}
		starttime, err := utils.StringToTime("15:04:05", val.(string))
		if err != nil {
			goto invalidError
		}

		val, exist = set["endtime"]
		if !exist {
			goto invalidError
		}
		endtime, err := utils.StringToTime("15:04:05", val.(string))
		if err != nil {
			goto invalidError
		}

		// val, exist = set["endsale"]
		// if !exist {
		// 	goto invalidError
		// }
		// endsale := val.(int64)

		val, exist = set["cycle"]
		if !exist {
			goto invalidError
		}
		cycle := val.(int64)

		for s := starttime; s+cycle < endtime; s += cycle {
			issueinfo := NewIssueinfo()
			issue := fmt.Sprintf("%s%40d", prefix, id)
			issueinfo.SetData("issue", issue)
			issueinfo.SetData("salestart", date.Add(time.Duration(s)).Format("2006-01-02 15:04:05"))
			issueinfo.SetData("saleend", date.Add(time.Duration(s+cycle)).Format("2006-01-02 15:04:05"))
			issueinfo.SetData("lotteryid", l.GetInt64("lotteryid"))
			issueinfo.SetData("belongdate", now.Format("2006-01-02"))
			issueinfo.SetTransaction(transaction)
			err := issueinfo.Save()
			if err != nil {
				return err
			}
		}
	}

	return nil

invalidError:
	transaction.Rollback()
	return fmt.Errorf("invalid issue set")
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
