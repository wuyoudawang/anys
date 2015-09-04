package model

import (
	"github.com/liuzhiyi/go-db"
)

type Singlesale struct {
	db.Item
}

func NewSinglesale() *Singlesale {
	s := new(Singlesale)
	s.Init("tasks", "taskid")
	return s
}
