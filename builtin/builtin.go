package builtin

import (
	"github.com/liuzhiyi/anys/config"
	"github.com/liuzhiyi/anys/jobs"
	"github.com/liuzhiyi/anys/services"
)

func InitServiceAndJob(c *config.Config) {
	registerServices(c)
	registerJobs(c)
}

func registerServices(c *config.Config) {
	conf := jobs.GetConf(c)
	eng := conf.GetEngine()
	eng.RegisterServer(services.NewTimerServer(eng), "MAIN_TIMER")
	eng.RegisterServer(services.NewDefaultLocalServer(eng), "MAIN_SIG")
}

func registerJobs(c *config.Config) {
	// eng := jobs.GetConf(c).GetEngine()
	// eng.Register(job)
}
