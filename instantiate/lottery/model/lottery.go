package model

import (
	"github.com/liuzhiyi/go-db"
	"time"
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
type Issueset struct {
}

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

func (l *Lottery) GetIssueset() {

}

func (l *Lottery) AutoGenerateIssues() {

}

func (l *Lottery) CreateIssue() (*Issueinfo, error) {
	now := time.Now()
	ise := NewIssueinfo()

	ise.SetData("lotteryid", l.GetInt64("lotteryid"))
	ise.SetData("belongdate", now.Format("2006-01-02"))
	err := ise.Save()
	return ise, err
}

func (l *Lottery) AutoClearIssues() {

}
