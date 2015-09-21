package model

import (
	"fmt"
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

func GetCurrentIssue(lotteryId int64) *Issueinfo {
	now := time.Now()
	current := now.Format("2006-01-02 15:04:05")
	before := now.Add(-20 * time.Minute).Format("2006-01-02 15:04:05")

	ise := NewIssueinfo()
	collection := ise.GetCollection()
	collection.AddFieldToSelect(
		"issueid, issue, belongdate, lotteryid",
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
		item := collection.GetItems()[0]
		ise.Item = *item

		return ise
	}

	return nil
}

func (i *Issueinfo) GetCurrentTasks() []*Tasks {
	var tsks []*Tasks

	if i.GetId() <= 0 {
		return tsks
	}

	collection := NewTasks().GetCollection()

	collection.JoinLeft(
		"userfund as u",
		"m.userid = u.userid",
		"availablebalance")
	collection.JoinLeft(
		"method as me",
		"me.methodid=m.methodid",
		"bei as maxmult")

	collection.AddFieldToFilter("m.lotteryid", "eq", i.GetInt64("lotteryid"))
	collection.AddFieldToFilter("m.status", "eq", 0)
	collection.AddFieldToFilter("m.beginissue", "lteq", i.GetString("issue"))

	collection.Load()
	for _, item := range collection.GetItems() {
		tsk := &Tasks{}
		tsk.Item = *item
		tsk.SetData("issue", i.GetString("issue"))
		tsks = append(tsks, tsk)
	}

	return tsks
}

func (i *Issueinfo) Task() error {
	transaction := i.GetResource().BeginTransaction()
	defer transaction.Commit()
	i.SetTransaction(transaction)

	tsks := i.GetCurrentTasks()
	for _, tsk := range tsks {
		tsk.SetTransaction(transaction)
		if err := tsk.AutoOrderBet(); err != nil {
			return err
		}
	}

	return i.FinishTask()
}

func (i *Issueinfo) FinishDraw(code string) error {
	transaction := i.GetTransaction()
	if transaction == nil {
		transaction = i.GetResource().BeginTransaction()
		i.SetTransaction(transaction)
		defer transaction.Commit()
	}

	i.SetData("code", code)
	i.SetData("statuscode", 2)
	i.SetData("statusfetch", 2)
	i.SetData("statuslocks", 2)
	i.SetData("writetime", i.Date())
	err := i.Save()
	if err != nil {
		return err
	}

	isehsty := NewIssuehistory()
	isehsty.SetTransaction(transaction)
	return isehsty.AddRow(i)

}

func (i *Issueinfo) GetCurrentProjects() []*Projects {
	collection := NewProjects().GetLotteryProjects(i.GetInt64("lotteryid"), i.GetString("issue"))
	collection.Load()

	var set []*Projects
	for _, item := range collection.GetItems() {
		p := new(Projects)
		p.Item = *item
		set = append(set, p)
	}

	return set
}

func (i *Issueinfo) FinishSendReward() error {
	i.SetData("statusbonus", 2)
	i.SetData("statuscheckbonus", 2)
	i.SetData("statususerpoint", 2)
	i.SetData("bonustime", i.Date())
	return i.Save()
}

func (i *Issueinfo) FinishTask() error {
	i.SetData("statusdeduct", 2)
	i.SetData("statustasktoproject", 2)
	return i.Save()
}

func (i *Issueinfo) Reset(winNum string) error {
	transaction := i.GetTransaction()
	if transaction == nil {
		transaction = i.GetResource().BeginTransaction()
		i.SetTransaction(transaction)
		defer transaction.Commit()
	}

	i.SetData("statuscheckbonus", 0)
	i.SetData("statusbonus", 0)
	i.SetData("statususerpoint", 0)
	i.SetData("statusdeduct", 0)
	i.SetData("code", winNum)
	i.SetData("bonustime", "0000-00-00 00:00:00")
	err := i.Save()
	if err != nil {
		return err
	}

	//更新历史开奖号码
	sql := "update issuehistory set code = ? where lotteryid = ? and issue= ?"
	_, err = transaction.Exec(sql, winNum, i.GetInt64("lotteryid"), i.GetString("issue"))
	if err != nil {
		transaction.Rollback()
		return err
	}

	return nil
}

func (i *Issueinfo) GetSubTotal(istester bool, mode int) map[string]float64 {
	read := i.GetResource().GetReadAdapter()

	sql := `select sum(IF(ordertypeid='3' or ordertypeid='6',amount, 0)) game_total,
sum(IF(ordertypeid='5',amount, 0)) reward_total,
sum(IF(ordertypeid='4',amount, 0)) rebate_total,
sum(IF(ordertypeid='9' or ordertypeid='7',amount, 0)) game_cancel_total,
sum(IF(ordertypeid='11',amount, 0)) rebate_cancel_total,
sum(IF(ordertypeid='12',amount, 0)) reward_cancel_total
FROM orders as o LEFT JOIN projects as p ON o.projectid = p.projectid
where p.issue = ? and o.lotteryid=? and o.istester = ? and p.modes = ?
GROUP BY p.issue`

	testFlag := "0"
	if istester {
		testFlag = "1"
	}

	data := make(map[string]float64)
	game_total := 0.0
	reward_total := 0.0
	rebate_total := 0.0
	game_cancel_total := 0.0
	rebate_cancel_total := 0.0
	reward_cancel_total := 0.0

	row, err := read.QueryRow(sql, i.GetData("issue"), i.GetData("lotteryid"), testFlag, mode)
	if err != nil {
		return data
	}

	row.Scan(
		&game_total,
		&reward_total,
		&rebate_total,
		&game_cancel_total,
		&rebate_cancel_total,
		&reward_cancel_total,
	)

	data["game_total"] = game_total
	data["reward_total"] = reward_total
	data["rebate_total"] = rebate_total
	data["game_cancel_total"] = game_cancel_total
	data["rebate_cancel_total"] = rebate_cancel_total
	data["reward_cancel_total"] = reward_cancel_total

	return data
}

func (i *Issueinfo) statisticInternel(mode int) error {
	data := i.GetSubTotal(false, mode)
	testdata := i.GetSubTotal(true, mode)

	singlesale := NewSinglesale()
	singlesale.SetData("sell", data["game_total"]-data["game_cancel_total"])
	singlesale.SetData("bonus", data["reward_total"]-data["reward_cancel_total"])
	singlesale.SetData("return", data["rebate_total"]-data["rebate_cancel_total"])

	singlesale.SetData("test_sell", testdata["game_total"]-testdata["game_cancel_total"])
	singlesale.SetData("test_bonus", testdata["reward_total"]-testdata["reward_cancel_total"])
	singlesale.SetData("test_return", testdata["rebate_total"]-testdata["rebate_cancel_total"])

	singlesale.SetData("lotteryid", i.GetInt64("lotteryid"))
	singlesale.SetData("modes", mode)
	singlesale.SetData("joindate", singlesale.Date())
	singlesale.SetData("issue", i.GetData("issue"))
	return singlesale.Save()
}

func (i *Issueinfo) Statistic() {
	transaction := i.GetResource().BeginTransaction()
	i.SetTransaction(transaction)
	defer transaction.Commit()

	if err := i.statisticInternel(1); err != nil {
		fmt.Println(err)
	}
	if err := i.statisticInternel(2); err != nil {
		fmt.Println(err)
	}
	if err := i.statisticInternel(3); err != nil {
		fmt.Println(err)
	}
}
