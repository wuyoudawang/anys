package model

import (
	db "github.com/liuzhiyi/go-db"
)

type Projects struct {
	db.Item
}

func NewProjects() *Projects {
	p := new(Projects)
	p.Init("projects", "projectid")
	return p
}

func (p *Projects) GetLotteryProjects(lotteryId int64, issue string) *db.Collection {
	collection := p.GetCollection()
	collection.JoinLeft(
		"method as me",
		"m.methodid=me.methodid",
		"customname, functionname, newcount, methodname, bei",
	)
	collection.AddFieldToSelect(
		"projectid, methodid, userid, code, packageid, maxmodel, multiple, "+
			"codetype, taskid, modes, omodel, parenttree,totalprice",
		collection.GetMainAlias(),
	)
	collection.AddFieldToFilter("m.issue", "eq", issue)
	collection.AddFieldToFilter("m.lotteryid", "eq", lotteryId)
	collection.AddFieldToFilter("m.isgetprize", "eq", 0)
	collection.AddFieldToFilter("m.iscancel", "eq", 0)

	return collection
}

func (p *Projects) GetMode() int64 {
	switch p.GetString("modes") {
	case "2":
		return 10
	case "3":
		return 100
	default:
		return 1
	}
}
