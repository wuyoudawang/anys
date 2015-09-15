package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	db "github.com/liuzhiyi/go-db"
)

const (
	AllBetween = -1

	RebateFixed = iota
	RebateTotalFixed
	RebateSteadyGrow
)

type RebateRule struct {
	betweens [][]int
	points   []float64
	modes    []int
	step     int
}

func (r *RebateRule) IsAll() bool {
	return len(r.betweens) == 1 &&
		len(r.betweens[0]) == 1 &&
		r.betweens[0][0] == AllBetween
}

func (r *RebateRule) Add(between []int, point float64, mode int) {
	if len(between) == 1 && between[0] == AllBetween {
		r.betweens = make([][]int, 1)
		r.betweens[0] = between
		r.points[0] = point
		r.modes[0] = mode
		return
	}

}

func (r *RebateRule) inBetween(index int, dst []int, point float64, mode int) (sub [][]int, flag int) {
	src := r.betweens[index]
	if len(src) == 1 {
		if len(dst) == 1 {
			if src[0] == dst[0] {
				r.points[index] = point
				r.modes[index] = mode
				return sub, 0
			} else {
				return [][]int{dst}, src[0] - dst[0]
			}
		}

		if src[0] > dst[0] && src[0] < dst[1] {
			r.points[index] = point
			r.modes[index] = mode
			return [][]int{{dst[0], src[0] - r.step}, {src[0] + r.step, dst[1]}}, 0
		} else if src[0] == dst[0] {
			r.remove(index)
			return [][]int{dst}, 1
		} else if src[0] == dst[1] {
			r.remove(index)
			r.modes[index] = mode
			return [][]int{dst}, -1
		}
	} else if len(src) == 2 {
		if len(dst) == 1 {
			if dst[0] > src[0] && dst[0] < src[1] {
				lastPoint := r.points[index]
				lastMode := r.modes[index]
				r.betweens[index] = []int{src[0], dst[0]}
				r.points[index] = point
				r.modes[index] = mode
				r.append([]int{dst[0] + r.step, src[1]}, index+1, lastPoint, lastMode)
				return
			} else if dst[0] == src[0] {
				lastPoint := r.points[index]
				lastMode := r.modes[index]
				r.betweens[index] = []int{dst[0]}
				r.points[index] = point
				r.modes[index] = mode
				r.append([]int{src[0] + r.step, src[1]}, index+1, lastPoint, lastMode)
				return
			} else if dst[0] == src[1] {
				r.betweens[index] = []int{src[0], src[1] - r.step}
				r.append([]int{dst[0]}, index+1, point, mode)
				return
			}

		}

		// var subst []int
		// if dst[0] =
	}

	return
}

func (r *RebateRule) remove(index int) {
	for i := index; i < len(r.betweens); i-- {
		r.betweens[i] = r.betweens[i+1]
		r.points[i] = r.points[i+1]
		r.modes[i] = r.modes[i+1]
	}

	r.betweens = r.betweens[0 : len(r.betweens)-1]
	r.points = r.points[0 : len(r.points)-1]
	r.modes = r.modes[0 : len(r.modes)-1]
}

func (r *RebateRule) append(between []int, index int, point float64, mode int) {
	r.betweens = append(r.betweens, between)
	r.points = append(r.points, point)
	r.modes = append(r.modes, mode)

	for i := len(r.betweens) - 2; i >= index; i-- {
		r.betweens[i+1] = r.betweens[i]
		r.points[i+1] = r.points[i]
		r.modes[i+1] = r.modes[i]
	}

	r.betweens[index] = between
	r.points[index] = point
	r.modes[index] = mode
}

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

func (p *Projects) CustomRebate(rule string) {

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

	err = p.CancelReward()
	if err != nil {
		return err
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
	order.SetData("fromuserid", p.GetInt64("userid"))
	order.SetData("ordertypeid", ordertype)
	order.SetData("title", order.TypeString())
	order.SetData("amount", amount)
	order.SetData("modes", p.GetString("modes"))
	order.SetData("pgcfg", p.GetString("omodel"))
	order.SetData("parenttree", p.GetUserTree())
	err := order.Create()

	return order, err
}

func (p *Projects) CancelReward() error {
	if p.GetInt("isgetprize") == 1 {

		_, err := p.CreateOrder(OrderCancelReward, p.GetFloat64("bonus"))
		if err != nil {
			return err
		}
	}

	return nil
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
