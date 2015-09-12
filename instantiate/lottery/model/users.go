package model

import (
	"fmt"
	"regexp"

	"github.com/liuzhiyi/go-db"
)

type Users struct {
	db.Item
}

func NewUsers() *Users {
	u := new(Users)
	u.Init("users", "userid")

	return u
}

func (u *Users) GetPgcodeByLotteryId(lotteryId int64) string {
	pgcfg := u.GetString("pgcfg")
	if pgcfg == "" {
		return ""
	}

	txt := fmt.Sprintf("[{,]%d:([0-9]{4})[,}]", lotteryId)
	reg, _ := regexp.Compile(txt)
	if matches := reg.FindStringSubmatch(pgcfg); len(matches) != 2 {
		return ""
	} else {
		return matches[1]
	}
}
