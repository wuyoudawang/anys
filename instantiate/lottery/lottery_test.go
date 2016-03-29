package lottery

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/liuzhiyi/anys/config"
	"github.com/liuzhiyi/anys/jobs"
	"github.com/liuzhiyi/anys/log"
	"github.com/liuzhiyi/anys/pkg/db"
	"github.com/liuzhiyi/anys/services"
)

func TestLottery(t *testing.T) {
	c := &config.Config{}
	initMaster(c)
	eng := jobs.NewEngine(3)
	eng.RegisterServer(services.NewTimerServer(eng), "test")
	eng.RegisterServer(services.NewLocalServer(eng, processSig, os.Kill), "sfe")

	ltyconf := GetConf(c)

	i := 0
	for _, lty := range ltyconf.GetAllLottery() {
		job := eng.NewJob(&lotteryJob{lty}, "test")
		job.Ticker(time.Duration(30+i*10) * time.Second)
		eng.Pending(job)

		i++
		if i == 2 {
			// break
		}
	}

	eng.Serve()

	exitMaster(c)
}

func processSig(sig os.Signal) error {
	fmt.Println(sig)
	return nil
}

func initMaster(c *config.Config) {
	loadCoreMudule(c)

	err := c.SortModules()
	if err != nil {
		panic(err.Error())
	}

	c.CreateConfModules()
	c.InitConfModules()

	err = c.Parse("../../conf/example.conf")
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
	c.LoadModule(ModuleName)
}
