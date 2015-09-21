package builtin

import (
	"anys/config"
	"anys/instantiate/lottery"
	"anys/jobs"
	"anys/services"
)

func RegisterServices(c *config.Config) {
	conf := jobs.GetConf(c)
	eng := conf.GetEngine()
	eng.RegisterServer(services.NewTimerServer(eng), "MAIN_TIMER")
	eng.RegisterServer(services.NewDefaultLocalServer(eng), "MAIN_SIG")
}

func RegisterJobs(c *config.Config) {
	// eng := jobs.GetConf(c).GetEngine()
	// eng.Register(job)
}
