package model

import (
	"github.com/liuzhiyi/go-db"
)

type Issueerror struct {
	db.Item
}

func NewIssueerror() *Issueerror {
	i := new(Issueerror)
	i.Init("issueerror", "entry")

	return i
}

func (i *Issueerror) GetRow(lotteryId int64) *Issueerror {
	isuerr := NewIssueerror()
	isuerr.SetData("statuscancelbonus", 0)
	isuerr.SetData("statusrepeal", 0)
	isuerr.SetData("lotteryid", lotteryId)
	isuerr.Row()

	return isuerr
}

// fetch the lately issueerrors by the lotteryid
func (i *Issueerror) GetLately(lotteryId int64) (set []*Issueerror) {
	collection := i.GetCollection()

	collection.AddFieldToFilter("statuscancelbonus", "eq", 0)
	collection.AddFieldToFilter("statusrepeal", "eq", 0)
	collection.AddFieldToFilter("lotteryid", "eq", lotteryId)

	collection.AddOrder("`writetime` ASC")
	collection.Load()

	for _, item := range collection.GetItems() {
		isee := &Issueerror{}
		isee.Item = *item
		set = append(set, isee)
	}

	return
}

func (i *Issueerror) Process() error {
	reopen := false

	if i.GetInt("errortype") == 2 {
		reopen = true
	}

	if reopen {
		ise := NewIssueinfo()
		ise.SetData("lotteryid", i.GetInt64("lotteryid"))
		ise.SetData("issue", i.GetString("issue"))
		ise.Row()
		ise.Reset(i.GetString("code"))
	} else {
		set := i.GetProjects()

		// cacelRebate := false
		// cancleReward := false
		// if i.GetInt("oldstatusbonus") == 2 {
		// 	cancleReward = true
		// }

		// if i.GetInt("oldstatususerpoint") == 2 {
		// 	cancelRebate = true
		// }

		for _, item := range set {
			err := item.Cancel()
			if err != nil {
				return err
			}
		}
	}

	return i.Finish()
}

func (i *Issueerror) GetProjects() []*Projects {
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

func (i *Issueerror) Finish() error {
	i.SetData("statusrepeal", 2)
	i.SetData("statuscancelbonus", 2)
	i.SetData("statusdeduct", 2)
	err := i.Save()
	return err
}
