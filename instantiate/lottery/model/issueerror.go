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

func (i *Issueerror) Process() {
	// cancelBonus := false
	// cancelRebate := false
	reopen := false

	// if i.GetInt("oldstatusbonus") == 2 {
	// 	cancelBonus = true
	// }

	// if i.GetInt("oldstatususerpoint") == 2 {
	// 	cancelRebate = true
	// }

	if i.GetInt("errortype") == 2 {
		reopen = true
	}

	if reopen {
		NewIssueinfo().Reset("winNum")
	}

	i.Finish()
}

func (i *Issueerror) Finish() {

}

func (i *Issueerror) Cancel() {

}
