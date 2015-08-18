package model

import (
	"fmt"
	"testing"

	"anys/config"
	_ "anys/pkg/db"
)

var c = &config.Config{}

func Test_Db(t *testing.T) {
	initConfig()

	p := NewProjects()
	// p.Load(51926)
	collection := p.GetLotteryProjects(8, "2015072910")

	collection.SetPageSize(10)
	collection.SetCurPage(1)
	collection.Load()
	fmt.Println(collection.GetSelect().Assemble())
	for _, item := range collection.GetItems() {
		fmt.Println(item.GetString("functionname"))
	}
}

func initConfig() {
	c.LoadModule("db")
	c.SortModules()

	c.CreateConfModules()
	c.InitConfModules()
	c.Parse("../../../pkg/db/db.conf")
	c.InitModules()
}
