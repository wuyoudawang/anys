package model

import (
	"github.com/liuzhiyi/go-db"
)

type Userfund struct {
	db.Item
}

func NewUserfund() *Userfund {
	u := new(Userfund)
	u.Init("userfund", "entry")

	return u
}

func (u *Userfund) LoadByUserId() {
	if u.GetId() <= 0 {
		if u.GetInt64("userid") > 0 {
			u.Row()
		}
	}
}

func (u *Userfund) AddSaleTotal(val float64) error {
	u.LoadByUserId()
	if u.GetId() <= 0 {
		if transaction := u.GetTransaction(); transaction != nil {
			transaction.Rollback()
		}
		return EmptyEntityErr
	}

	saleTotal := u.GetFloat64("lotteBalance")
	saleTotal += val
	u.SetData("lotteBalance", saleTotal)
	return u.Save()
}

func (u *Userfund) AddRebateTotal(val float64) error {
	u.LoadByUserId()
	if u.GetId() <= 0 {
		if transaction := u.GetTransaction(); transaction != nil {
			transaction.Rollback()
		}
		return EmptyEntityErr
	}

	rebateTotal := u.GetFloat64("totalpointBalance")
	rebateTotal += val
	u.SetData("totalpointBalance", rebateTotal)
	return u.Save()

}

func (u *Userfund) AddRewardTotal(val float64) error {
	u.LoadByUserId()
	if u.GetId() <= 0 {
		if transaction := u.GetTransaction(); transaction != nil {
			transaction.Rollback()
		}
		return EmptyEntityErr
	}

	rewardTotal := u.GetFloat64("winbalance")
	rewardTotal += val
	u.SetData("winbalance", rewardTotal)
	return u.Save()
}

func (u *Userfund) AddBalance(val float64) error {
	u.LoadByUserId()
	if u.GetId() <= 0 {
		if transaction := u.GetTransaction(); transaction != nil {
			transaction.Rollback()
		}
		return EmptyEntityErr
	}

	total := u.GetFloat64("availablebalance")
	total += val
	u.SetData("availablebalance", total)
	return u.Save()
}
