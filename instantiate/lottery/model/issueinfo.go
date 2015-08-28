package model

import (
	"time"

	db "github.com/liuzhiyi/go-db"
)

type Issueinfo struct {
	db.Item
}

func NewIssueinfo() *Issueinfo {
	i := new(Issueinfo)
	i.Init("issueinfo", "issueid")
	return i
}

func (i *Issueinfo) GetCurrentIssue(lotteryId int64) *Issueinfo {
	now := time.Now()
	current := now.Format("2006-01-02 15:04:05")
	before := now.Add(-20 * time.Minute).Format("2006-01-02 15:04:05")

	collection := i.GetCollection()
	collection.AddFieldToSelect(
		"issueid, issue",
		collection.GetMainAlias(),
	)

	collection.AddFieldToFilter("m.statusfetch", "eq", 0)
	collection.AddFieldToFilter("m.statuscode", "eq", 0)
	collection.AddFieldToFilter("m.statuscheckbonus", "eq", 0)
	collection.AddFieldToFilter("m.saleend", "lt", current)
	collection.AddFieldToFilter("m.saleend", "gt", before)
	collection.AddFieldToFilter("m.lotteryid", "eq", lotteryId)
	collection.AddOrder("saleend desc")
	collection.SetPageSize(1)

	collection.Load()

	if collection.Count() > 0 {
		ni := new(Issueinfo)
		item := collection.GetItems()[0]
		ni.Item = *item

		return ni
	}

	return nil
}

func (i *Issueinfo) FinishDraw(code string) {
	i.SetData("code", code)
	i.SetData("statuscode", 2)
	i.SetData("statusfetch", 2)
	i.SetData("statuslocks", 2)
	i.Save()

}

func (i *Issueinfo) FinishSendReward() {
	i.SetData("statusbonus", 2)
	i.SetData("statususerpoint", 2)
}

func (i *Issueinfo) FinishTask() {
	i.SetData("statusdeduct", 2)
	i.SetData("statustasktoproject", 2)
}
