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

	issueinfo := model.NewIssueinfo()
	p := model.NewProjects()
	ticker := time.NewTicker(60 * time.Second)

	for range ticker.C {

		issue := issueinfo.GetCurrentIssue(1)
		if issue == nil {
			continue
		}

		collection := p.GetLotteryProjects(1, issue.GetString("issue"))
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

		lty.Reduce()
		winNum := lty.Draw()
		fmt.Println(winNum)
		key, _ := lty.GenerateKey(winNum)
		fmt.Println(lty.GetTotalReward(key))
		fmt.Println(lty.GetGross())
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
