package main

import (
	"github.com/liuzhiyi/anys/builtin"
	"github.com/liuzhiyi/anys/config"
	"github.com/liuzhiyi/anys/instantiate/lottery"
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

	//app start
	Shedule(c)

	exitMaster(c)
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
