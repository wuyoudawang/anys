package main

import (
	"fmt"

	"anys/config"
	"anys/instantiate/lottery"
	"anys/instantiate/lottery/model"
	"anys/log"
	"anys/pkg/db"
)

func main() {
	c := &config.Config{}
	initMaster(c)

	p := model.NewProjects()
	collection := p.GetLotteryProjects(1, "20150811084")
	collection.Load()

	lcf := lottery.GetConf(c)
	lty, err := lcf.GetLottery("ssj")
	if err != nil {
		panic(err.Error())
	}

	for _, item := range collection.GetItems() {
		p := &model.Projects{Item: *item}
		err = lty.Dispatch(p)
		fmt.Println(err)
	}
	key, err := lty.GenerateKey("30090")
	fmt.Println(lty.GetTotalReward(key))
	lty.Reduce()
	winNum := lty.Draw()
	fmt.Println(winNum)
	key, _ = lty.GenerateKey(winNum)
	fmt.Println(lty.GetTotalReward(key))

	exitMaster(c)
	fmt.Println("finish")
}

func initMaster(c *config.Config) {
	loadCoreMudule(c)

	err := c.SortModules()
	if err != nil {
		panic(err.Error())
	}

	c.CreateConfModules()
	c.InitConfModules()

	err = c.Parse("../conf/example.conf")
	if err != nil {
		panic(err.Error())
	}

	c.InitModules()
}

func exitMaster(c *config.Config) {
	c.ExitModules()
}

func loadCoreMudule(c *config.Config) {
	c.LoadModule(db.ModuleName)
	c.LoadModule(log.ModuleName)
	c.LoadModule(lottery.ModuleName)
}
