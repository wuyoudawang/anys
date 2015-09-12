package model

import (
	db "github.com/liuzhiyi/go-db"
)

type Issuehistory struct {
	db.Item
}

func NewIssuehistory() *Issuehistory {
	i := new(Issuehistory)
	i.Init("issuehistory", "issueid")
	return i
}

func (i *Issuehistory) AddRow(isue *Issueinfo) error {

	i.SetId(0)
	i.SetData("lotteryid", isue.GetInt64("lotteryid"))
	i.SetData("code", isue.GetString("code"))
	i.SetData("issue", isue.GetString("issue"))
	i.SetData("belongdate", isue.GetString("belongdate"))
	return i.Save()
}
