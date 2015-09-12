package model

import (
	"fmt"
	"strconv"
	"strings"
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

	err = p.Rebate()
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
	transaction := p.GetTransaction()
	if transaction == nil {
		transaction = p.GetResource().BeginTransaction()
		p.SetTransaction(transaction)
		defer transaction.Commit()
	}

	p.Rebate()

	p.SetData("isgetprize", 2)
	return p.Save()
}

func (p *Projects) FlushTask() error {
	tsk := NewTasks()
	tsk.SetTransaction(p.GetTransaction())
	tsk.Load(p.GetInt("taskid"))
	return tsk.Flush(p)
}

func (p *Projects) GetCurrentUserMaxPoint() float64 {
	pg := NewPrizegroup()
	return pg.GetPointByPgcodeId(p.GetInt64("lotteryid"), p.GetString("maxmodel"))
}

func (p *Projects) GetCurrentUserPoint() float64 {
	pg := NewPrizegroup()
	return pg.GetPointByPgcodeId(p.GetInt64("lotteryid"), p.GetString("omodel"))
}

func (p *Projects) GetUserPgcode(userid int) string {
	u := NewUsers()
	u.Load(userid)
	return u.GetPgcodeByLotteryId(p.GetInt64("lotteryid"))
}

func (p *Projects) GetUserPoint(userid int) float64 {
	pgcode := p.GetUserPgcode(userid)
	if pgcode != "" {
		pg := NewPrizegroup()
		return pg.GetPointByPgcodeId(p.GetInt64("lotteryid"), pgcode)
	}

	return 0
}

func (p *Projects) Rebate() error {
	var (
		point float64
	)

	// 判断是否属于特殊玩法
	// if "syxw" == getLotteryName() && (p.FuncName == "Odd" || p.FuncName == "MidSize") {
	// 	if p.Omodel >= 1850.0 {
	// 		step := 2
	// 		uintRate := 0.001
	// 		tmpId := ""
	// 		pgCodeA := 0
	// 		pgCodeB := 0
	// 		for _, parentId := range strings.Split(p.ParentTree, ",") {
	// 			parentId = strings.TrimSpace(parentId)
	// 			if parentId != "" {
	// 				pgcode := GetUserPgcode(parentId)
	// 				if pgcode != "" {
	// 					pgCodeA = pgCodeB
	// 					pgCodeB, _ = strconv.Atoi(pgcode)
	// 					if pgCodeB > 0 && pgCodeA > 0 && float64((pgCodeA-pgCodeB)/step)*uintRate > 0 {
	// 						money := float64((pgCodeA-pgCodeB)/step) * uintRate * p.TotalPrice
	// 						insertOrder(p, money, tmpId, "4") //返点插入订单需要在更新额度前完成，不然会统计不到转变前额度
	// 					}
	// 					tmpId = parentId
	// 				}
	// 			}
	// 		}

	// 		// 计算自己直接上级的返点
	// 		pgCodeA, _ = strconv.Atoi(p.MaxModel)
	// 		if float64((pgCodeB-pgCodeA)/step)*uintRate > 0 {
	// 			money := float64((pgCodeB-pgCodeA)/step) * uintRate * p.TotalPrice
	// 			insertOrder(p, money, tmpId, "4") //返点插入订单需要在更新额度前完成，不然会统计不到转变前额度
	// 		}

	// 		//计算当前用户选择的投注模式和最大模式之间的返点差
	// 		pgCodeB = int(p.Omodel)
	// 		if float64((pgCodeA-pgCodeB)/step)*uintRate > 0 {
	// 			money := float64((pgCodeA-pgCodeB)/step) * uintRate * p.TotalPrice
	// 			insertOrder(p, money, p.UserId, "4") //返点插入订单需要在更新额度前完成，不然会统计不到转变前额度
	// 		}
	// 	}
	// } else {

	//计算父路径用户的返点差,数据库存放顶级降序，所以第一个是代理
	lastParentPoint := 0.0
	parentPoint := 0.0
	lastParentId := 0
	maxPoint := p.GetCurrentUserMaxPoint()
	for _, idStr := range strings.Split(p.GetUserTree(), ",") {
		parentId, err := strconv.Atoi(idStr)
		if err != nil {
			if p.GetTransaction() != nil {
				p.GetTransaction().Rollback()
			}
			return err
		}

		parentPoint := p.GetUserPoint(parentId)
		if parentPoint < 0 {
			if p.GetTransaction() != nil {
				p.GetTransaction().Rollback()
			}
			return fmt.Errorf("")
		}

		if lastParentId > 0 {
			if lastParentPoint-parentPoint > 0 {
				amount := (lastParentPoint - parentPoint) * p.GetFloat64("totalprice")
				_, err := p.CreateOrder(OrderRebate, amount)
				if err != nil {
					return err
				}
			}
		}

		lastParentId = parentId
		lastParentPoint = parentPoint
	}

	// 计算自己直接上级的返点
	if parentPoint > 0 && parentPoint-maxPoint > 0 {
		amount := (parentPoint - maxPoint) * p.GetFloat64("totalprice")
		_, err := p.CreateOrder(OrderRebate, amount) //返点插入订单需要在更新额度前完成，不然会统计不到转变前额度
		if err != nil {
			return err
		}
	}

	//计算当前用户选择的投注模式和最大模式之间的返点差
	point = p.GetCurrentUserPoint()
	if maxPoint > point {
		amount := (maxPoint - point) * p.GetFloat64("totalprice")
		_, err := p.CreateOrder(OrderRebate, amount) //返点插入订单需要在更新额度前完成，不然会统计不到转变前额度
		if err != nil {
			return err
		}
	}

	return nil
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
