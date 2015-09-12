package model

import (
	"github.com/liuzhiyi/go-db"
)

type Prizegroup struct {
	db.Item
}

func NewPrizegroup() *Prizegroup {
	p := new(Prizegroup)
	p.Init("prizegroup", "prizegroupid")

	return p
}

func (p *Prizegroup) GetPointByPgcodeId(lotteryId int64, pgcode string) float64 {
	p.SetData("lotteryid", lotteryId)
	p.SetData("pgcode", pgcode)
	p.Row()
	return p.GetFloat64("userpoint")
}
