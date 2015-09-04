package model

import (
	"time"

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

func (p *Projects) Create() error {
	p.SetData("writetime", time.Now().Format("2006-01-02 15:04:05"))

	return p.Save()
}

func (p *Projects) Reward(reward float64) error {
	transaction := p.GetTransaction()
	if transaction == nil {
		transaction = p.GetResource().BeginTransaction()
		p.SetTransaction(transaction)
		defer transaction.Commit()
	}

	p.SetData("isgetprize", 1)
	p.SetData("bonus", reward)
	p.SetData("bonustime", time.Now().Format("2006-01-02 15:04:05"))
	err := p.Save()
	if err != nil {
		return err
	}

	_, err = p.CreateOrder(OrderReward, reward)
	return err
}

func (p *Projects) GetUserTree() string {
	if p.GetData("parenttree") == nil {
		usertree := NewUsertree()
		usertree.SetData("userid", p.GetData("userid"))
		usertree.Row()
		p.SetData("parenttree", usertree.GetData("parenttree"))
	}

	return p.GetString("parenttree")
}

func (p *Projects) Unreward() error {
	p.SetData("isgetprize", 2)
	return p.Save()
}

func (p *Projects) FlushTask() error {
	tsk := NewTasks()
	tsk.SetTransaction(p.GetTransaction())
	tsk.Load(p.GetInt("taskid"))
	return tsk.Flush(p)
}

func (p *Projects) Cancel() error {
	transaction := p.GetTransaction()
	if transaction == nil {
		transaction = p.GetResource().BeginTransaction()
		p.SetTransaction(transaction)
		defer transaction.Commit()
	}

	p.SetData("iscancel", 3)
	err := p.Save()
	if err != nil {
		return err
	}

	ordertype := OrderCancelGame
	if p.GetInt("taskid") > 0 {
		ordertype = OrderCancelTask
	}

	_, err = p.CreateOrder(ordertype, p.GetFloat64("totalprice"))
	if err != nil {
		return err
	}

	if p.GetInt("isgetprize") == 1 {

		_, err = p.CreateOrder(OrderCancelReward, p.GetFloat64("bonus"))
		if err != nil {
			return err
		}
	}

	err = p.CancelRebate()
	if err != nil {
		return err
	}

	return p.FlushTask()
}

func (p *Projects) CreateOrder(ordertype int, amount float64) (*Orders, error) {
	order := NewOrders()
	order.SetTransaction(p.GetTransaction())
	order.SetData("lotteryid", p.GetInt64("lotteryid"))
	order.SetData("methodid", p.GetInt64("methodid"))
	order.SetData("projectid", p.GetId())
	order.SetData("packageid", p.GetInt64("packageid"))
	order.SetData("taskid", p.GetInt64("taskid"))
	order.SetData("formuserid", p.GetInt64("userid"))
	order.SetData("ordertypeid", ordertype)
	order.SetData("title", order.TypeString())
	order.SetData("amount", amount)
	order.SetData("modes", p.GetString("modes"))
	order.SetData("pgcfg", p.GetString("omodel"))
	order.SetData("parenttree", p.GetUserTree())
	err := order.Create()

	return order, err
}

func (p *Projects) CancelRebate() error {
	collection := NewOrders().GetCollection()

	collection.AddFieldToFilter("ordertypeid", "eq", OrderRebate)
	collection.AddFieldToFilter("projectid", "eq", p.GetId())

	collection.Load()
	for _, item := range collection.GetItems() {
		_, err := p.CreateOrder(OrderCancelRebate, item.GetFloat64("amount"))
		if err != nil {
			return err
		}
	}

	return nil
}
