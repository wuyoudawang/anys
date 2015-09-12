package lottery

import (
	"fmt"
	"os"
	"testing"
	"time"

	"anys/config"
	"anys/jobs"
	"anys/log"
	"anys/pkg/db"
	"anys/services"
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
		job.Ticker(30 * time.Second)
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
