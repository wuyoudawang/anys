package main

import (
	"fmt"
	"time"

	"github.com/liuzhiyi/anys/builtin"
	"github.com/liuzhiyi/anys/config"
	"github.com/liuzhiyi/anys/instantiate/lottery"
	"github.com/liuzhiyi/anys/instantiate/lottery/model"
	"github.com/liuzhiyi/anys/jobs"
	"github.com/liuzhiyi/anys/log"
	"github.com/liuzhiyi/anys/pkg/db"
)

func Shedule(c *config.Config) {
	builtin.InitServiceAndJob(c)
	eng := jobs.GetConf(c).GetEngine()
	eng.Serve()
}

func main() {
	c := &config.Config{}
	initMaster(c)

	// Shedule(c)

	// ccssj := model.NewLottery()
	// ccssj.Load(1)
	// go generateIssue(ccssj, 60*60*10)

	lcf := lottery.GetConf(c)
	last := 1
	for name, _ /*lty*/ := range lcf.GetAllLottery() {
		if last == len(lcf.GetAllLottery()) {
			processLottery(c, name, 30)
		}
		go processLottery(c, name, 30)
		// go generateIssue(lty.GetLotteryModel(), 60*60*10)
		last++
	}

	exitMaster(c)
	fmt.Println("finish")
}

func generateIssue(ltyM *model.Lottery, interval time.Duration) {
	ltyM.AutoClearIssues()
	if err := ltyM.AutoGenerateIssues(); err != nil {
		log.Warning("%d奖期自动生成产生错误：%s", ltyM.GetId(), err.Error())
	}

	ticker := time.NewTicker(interval * time.Second)
	for range ticker.C {
		ltyM.AutoClearIssues()
		if err := ltyM.AutoGenerateIssues(); err != nil {
			log.Warning("奖期自动生成产生错误：%s", err.Error())
		}
	}
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

		issue := lty.GetCurrentIssue()
		if issue == nil {
			continue
		}

		err := lty.Task()
		if err != nil {
			fmt.Println(err)
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

		err = lty.Persist(winNum)
		fmt.Println(err)

		// lty.SendReward(key)
		// issue.Statistic()

		// ltyObj := lty.GetLotteryModel()
		// ltyObj.ProcessIssueError()

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
	lottery.Install(c)
	db.Install(c)
	log.Install(c)
	jobs.Install(c)
}
