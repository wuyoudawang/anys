package model

import (
	"fmt"
	"math"

	"github.com/liuzhiyi/go-db"
)

var (
	EmptyEntityErr = fmt.Errorf("using an empty entity")
)

type Tasks struct {
	db.Item
}

func NewTasks() *Tasks {
	t := new(Tasks)
	t.Init("tasks", "taskid")
	return t
}

func (t *Tasks) AutoOrderBet() error {

	if t.GetId() <= 0 {
		return EmptyEntityErr
	}

	balance := t.GetBalance()
	singleprice := t.GetFloat64("singleprice")
	bei := t.GetFloat64("bei")
	maxMultiple := t.GetMaxMutilple()
	finishedcount := t.GetInt64("finishedcount")
	issuecount := t.GetInt64("issuecount")
	projectcount := t.GetInt64("projectcount")
	cancelcount := t.GetInt64("cancelcount")

	multiple := 1.0
	if t.GetInt("type") == 1 {
		multiple = math.Pow(bei, float64(projectcount-cancelcount+1))
	} else {
		multiple = bei
	}

	if multiple > maxMultiple {
		multiple = maxMultiple
	}

	price := multiple * singleprice

	if finishedcount >= issuecount || balance < price {
		return t.Finish()
	} else {

		//插入projects
		project, err := t.CreatProject(multiple, singleprice, price)
		if err != nil {
			return err
		}

		//插入orders
		_, err = t.CreateOrder(OrderTask, project.GetInt64("projectid"))
		if err != nil {
			return err
		}

		//更新投注数
		t.SetData("projectcount", projectcount+1)
		return t.Save()
	}
}

func (t *Tasks) CreatProject(multiple, singleprice, price float64) (*Projects, error) {
	project := NewProjects()
	project.SetTransaction(t.GetTransaction())
	project.SetData("userid", t.GetInt64("userid"))
	project.SetData("taskid", t.GetInt64("taskid"))
	project.SetData("lotteryid", t.GetInt64("lotteryid"))
	project.SetData("methodid", t.GetInt64("methodid"))
	project.SetData("issue", t.GetString("issue"))
	project.SetData("code", t.GetString("codes"))
	project.SetData("multiple", multiple)
	project.SetData("singleprice", singleprice)
	project.SetData("totalprice", price)
	project.SetData("modes", t.GetData("modes"))
	project.SetData("omodel", t.GetData("omodel"))
	project.SetData("maxmodel", t.GetData("maxmodel"))
	project.SetData("codetype", t.GetData("codetype"))
	project.SetData("lvtopid", t.GetData("lvtopid"))
	project.SetData("lvtoppoint", t.GetData("lvtoppoint"))
	project.SetData("parenttree", t.GetUserTree())
	project.SetData("lvproxyid", t.GetData("lvproxyid"))
	project.SetData("userip", t.GetData("userip"))
	project.SetData("cdnip", t.GetData("cdnip"))
	err := project.Create()

	return project, err
}

func (t *Tasks) CreateOrder(ordertype int, projectId int64) (*Orders, error) {
	order := NewOrders()
	order.SetTransaction(t.GetTransaction())
	order.SetData("lotteryid", t.GetInt64("lotteryid"))
	order.SetData("methodid", t.GetInt64("methodid"))
	order.SetData("projectid", projectId)
	order.SetData("packageid", 0)
	order.SetData("taskid", t.GetInt64("taskid"))
	order.SetData("formuserid", t.GetInt64("userid"))
	order.SetData("ordertypeid", ordertype)
	order.SetData("title", order.TypeString())
	order.SetData("amount", t.GetFloat64("totalprice"))
	order.SetData("modes", t.GetString("modes"))
	order.SetData("pgcfg", t.GetString("omodel"))
	order.SetData("parenttree", t.GetUserTree())
	err := order.Create()

	return order, err
}

func (t *Tasks) Flush(p *Projects) error {
	if p.GetInt("taskid") == 0 {
		return EmptyEntityErr
	}

	if p.GetInt("iscancel") > 0 {
		cancelPrice := t.GetFloat64("cancelprice")
		cancelPrice += p.GetFloat64("totalprice")
		cancelCount := t.GetInt("cancelcount") + 1

		t.SetData("cancelprice", cancelPrice)
		t.SetData("cancelPrice", cancelCount)
		return t.Save()
	} else {
		wincount := t.GetInt("wincount")
		wincount += 1
		done := false

		if t.GetInt("stoponwin") == 1 {
			done = true
		} else {
			if t.GetInt("type") == 2 {
				taskprice := t.GetFloat64("taskprice")
				lowestRate := t.GetFloat64("lowest_rate")

				if float64(wincount)*p.GetFloat64("bonus") >= taskprice+taskprice*lowestRate {
					done = true
				}
			}
		}

		t.SetData("finishprice", t.GetFloat64("finishprice")+p.GetFloat64("totalprice"))
		t.SetData("finishedcount", t.GetInt("finishedcount")+1)
		if done {
			return t.Finish()
		} else {
			return t.Save()
		}
	}
}

func (t *Tasks) GetModes() string {
	if t.GetData("modes") == nil {
		usertree := NewUsertree()
		usertree.SetData("userid", t.GetData("userid"))
		usertree.Row()
		t.SetData("modes", usertree.GetData("modes"))
	}

	return t.GetString("modes")
}

func (t *Tasks) GetMaxMutilple() float64 {
	if t.GetData("maxmult") == nil {
		method := NewMethod()
		method.SetData("methodid", t.GetData("methodid"))
		method.Row()
		t.SetData("maxmult", method.GetData("bei"))
	}

	return t.GetFloat64("maxmult")
}

func (t *Tasks) GetBalance() float64 {
	if t.GetData("availablebalance") == nil {
		userfund := NewMethod()
		userfund.SetData("userid", t.GetData("userid"))
		userfund.Row()
		t.SetData("availablebalance", userfund.GetData("availablebalance"))
	}

	return t.GetFloat64("availablebalance")
}

func (t *Tasks) GetUserTree() string {
	if t.GetData("parenttree") == nil {
		usertree := NewUsertree()
		usertree.SetData("userid", t.GetData("userid"))
		usertree.Row()
		t.SetData("parenttree", usertree.GetData("parenttree"))
	}

	return t.GetString("parenttree")
}

func (t *Tasks) Finish() error {
	t.SetData("status", 2)
	return t.Save()
}
