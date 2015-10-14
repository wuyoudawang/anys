package jobs

import (
	"time"

	"anys/config"
	"anys/log"
)

const (
	ModuleName = "jobs"
)

func Install(c *config.Config) {
	c.LoadModule(ModuleName, log.ModuleName)
}

type jobsConf struct {
	goroutines     int
	cycleQueueSize int
	eng            *Engine
}

func (jc *jobsConf) GetEngine() *Engine {
	return jc.eng
}

var (
	commands = []*config.Command{
		{config.BLOCK, "job", jobBlock},
		{config.MORE1, "goroutines", setGoroutines},
		{config.MORE1, "cycleQueueSize", setCycleQueueSize},
		{config.MORE1, "create", createJob},
	}

	jobsModule = &config.Module{
		Commands: commands,

		Create_conf: createConf,
		Init_conf:   initConf,

		Init_module: initModule,
	}
)

func init() {
	config.RegisterModule(ModuleName, jobsModule)
}

func createConf(c *config.Config) {
	conf := &jobsConf{}
	c.SetConf(config.GetModule(ModuleName), conf)
}

func initConf(c *config.Config) error {

	return nil
}

func initModule(c *config.Config) error {
	conf := GetConf(c)
	if conf.goroutines > MAXGOROUTINE || conf.goroutines < 5 {
		conf.goroutines = defaultGoRoutine / 2
	}
	conf.eng = NewEngine(conf.goroutines)
	return nil
}

func GetConf(c *config.Config) *jobsConf {
	conf := c.GetConf(config.GetModule(ModuleName))
	return conf.(*jobsConf)
}

func jobBlock(c *config.Config) error {

	return c.Block()
}

func setGoroutines(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.goroutines, err = c.Int(args[1])
	return err
}

func setCycleQueueSize(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.cycleQueueSize, err = c.Int(args[1])
	return err
}

func createJob(c *config.Config) error {
	h := func(c *config.Config) error {
		args, err := c.CheckArgs(3)
		if err != nil {
			return err
		}

		eng := GetConf(c).GetEngine()
		name := args[1]
		val, err := c.Int(args[2])
		if err != nil {
			return err
		}

		job := eng.Job(name)
		if job != nil {
			job.Ticker(time.Duration(val) * time.Second)
			eng.Pending(job)
		}

		return nil
	}
	c.DelayCaller(h)
	return nil
}
