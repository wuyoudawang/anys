package model

import (
	"time"

	"github.com/liuzhiyi/go-db"
)

const (
	OrderJoin         = 3
	OrderRebate       = 4
	OrderReward       = 5
	OrderGame         = 8
	OrderTask         = 6
	OrderCancelTask   = 7
	OrderCancelGame   = 9
	OrderCancelReward = 12
	OrderCancelRebate = 11
)

type Orders struct {
	db.Item
}

func NewOrders() *Orders {
	o := new(Orders)
	o.Init("orders", "orderid")

	return o
}

func (o *Orders) Create() error {

	transaction := o.GetTransaction()
	if transaction == nil {
		transaction = o.GetResource().BeginTransaction()
		o.SetTransaction(transaction)
		defer transaction.Commit()
	}

	o.SetData("times", time.Now().Format("2006-01-02 15:04:05"))
	o.SetData("actiontime", time.Now().Format("2006-01-02 15:04:05"))

	//不明字段（暂时不处理）
	o.SetData("clientip", "127.0.0.1")
	o.SetData("proxyip", "127.0.0.1")

	err := o.Save()
	if err != nil {
		return err
	}

	userfund := NewUserfund()
	userfund.SetData("userid", o.GetData("userid"))
	userfund.SetTransaction(transaction)
	switch o.GetInt("ordertypeid") {
	case OrderJoin:
	case OrderGame:
	case OrderTask:
		err = userfund.AddSaleTotal(o.GetFloat64("amount"))
	case OrderRebate:
		err = userfund.AddRebateTotal(o.GetFloat64("amount"))
	case OrderReward:
		err = userfund.AddRewardTotal(o.GetFloat64("amount"))
	case OrderCancelGame:
	case OrderCancelTask:
		err = userfund.AddSaleTotal(-o.GetFloat64("amount"))
	case OrderCancelReward:
		err = userfund.AddRewardTotal(-o.GetFloat64("amount"))
	case OrderCancelRebate:
		err = userfund.AddRebateTotal(-o.GetFloat64("amount"))
	}

	return err
}

func (o *Orders) TypeString() string {
	switch o.GetData("ordertypeid") {
	case OrderJoin:
		return "加入游戏"
	case OrderRebate:
		return "投注返点"
	case OrderReward:
		return "奖金派送"
	case OrderTask:
		return "追号扣款"
	case OrderGame:
		return "游戏扣款"
	case OrderCancelTask:
		return "当前追号返款"
	case OrderCancelGame:
		return "撤单返款"
	case OrderCancelRebate:
		return "撤销返点"
	case OrderCancelReward:
		return "撤销派奖"
	}

	return "未知订单类型"
}
