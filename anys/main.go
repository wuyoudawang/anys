package main

import (
	"fmt"
	"time"

	"anys/config"
	"anys/instantiate/lottery"
	"anys/instantiate/lottery/model"
	"anys/log"
	"anys/pkg/db"
)

func main() {
	c := &config.Config{}
	initMaster(c)
	go processLottery(c, "tcaifive", 60)
	go processLottery(c, "TCFFC", 30)
	processLottery(c, "tcaithird", 60)

	exitMaster(c)
	fmt.Println("finish")
}

func processLottery(c *config.Config, name string, interval time.Duration) {
	p := model.NewProjects()

	lcf := lottery.GetConf(c)
	lty, err := lcf.GetLottery(name)
	if err != nil {
		panic(err.Error())
	}

	ticker := time.NewTicker(interval * time.Second)

	for range ticker.C {

		issue := model.GetCurrentIssue(lty.GetId())
		if issue == nil {
			continue
		}

		collection := p.GetLotteryProjects(lty.GetId(), issue.GetString("issue"))
		collection.Load()

		for _, item := range collection.GetItems() {
			p := &model.Projects{Item: *item}
			err = lty.Dispatch(p)
			fmt.Println(err)
		}

		lty.Reduce()
		winNum := lty.Draw()
		fmt.Println(winNum)
		key, _ := lty.GenerateKey(winNum)
		fmt.Println(lty.GetTotalReward(key))
		fmt.Println(lty.GetGross())
		// for _, record := range lty.GetRecords(key) {
		// 	fmt.Println(record)
		// }

		// key, _ = lty.GenerateKey("30030")
		// fmt.Println(lty.GetTotalReward(key))
		// for _, record := range lty.GetRecords(key) {
		// 	fmt.Println(record)
		// }

		issue.SetData("code", winNum)
		issue.SetData("statuscode", 2)
		issue.SetData("statusfetch", 2)
		issue.SetData("statuslocks", 2)
		issue.SetData("statustasktoproject", 2)
		issue.SetData("statusdeduct", 2)
		err = issue.Save()
		fmt.Println(err)

		lty.Reset()
	}
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
	c.LoadModule(
		lottery.ModuleName,
		db.ModuleName,
		log.ModuleName,
	)
}
