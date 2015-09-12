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

func TestGenerateIssue(t *testing.T) {
	initConfig()

	lottery := NewLottery()
	lottery.Load(1)
	lottery.AutoClearIssues()
	err := lottery.AutoGenerateIssues()
	fmt.Println(err)
}

func TestPgcode(t *testing.T) {
	initConfig()

	u := NewUsers()
	u.SetData("pgcfg", "{1:1956,3:1956,5:1936,6:1956,7:1936,8:1936,11:1926,12:1926,13:1956}")
	pgcode := u.GetPgcodeByLotteryId(7)
	fmt.Println(pgcode)
}

func initConfig() {
	c.LoadModule("db")
	c.SortModules()

	c.CreateConfModules()
	c.InitConfModules()
	c.Parse("../../../pkg/db/db.conf")
	c.InitModules()
}
