package model

import (
	"github.com/liuzhiyi/go-db"
)

type Method struct {
	db.Item
}

func NewMethod() *Method {
	m := new(Method)
	m.Init("method", "methodid")

	return m
}
