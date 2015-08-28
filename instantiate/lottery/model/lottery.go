package model

import (
	"github.com/liuzhiyi/go-db"
)

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
