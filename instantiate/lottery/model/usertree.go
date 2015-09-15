package model

import (
	"github.com/liuzhiyi/go-db"
)

type Usertree struct {
	db.Item
}

func NewUsertree() *Usertree {
	u := new(Usertree)
	u.Init("usertree", "userid")

	return u
}
