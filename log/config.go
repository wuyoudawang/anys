package log

import (
	"anys/config"
	"path"

	"github.com/astaxie/beego/logs"
)

const (
	ModuleName = "log"
)

type LogConf struct {
	BufSize  int
	FileName string
	Level    int
	Daily    bool
	Maxdays  int
}

var (
	commands = []*config.Command{
		{config.BLOCK, "log", logBlock},
		{config.MORE1, "bufSize", setBufSize},
		{config.MORE1, "fileName", setFileName},
		{config.MORE1, "level", setLevel},
		{config.MORE1, "daily", setDaily},
		{config.MORE1, "maxdays", setMaxdays},
	}

	logModule = &config.Module{
		Commands: commands,

		Create_conf: createConf,
		Init_conf:   initConf,

		Init_module: initModule,
		Exit_module: exitModule,
	}
)

func init() {
	config.RegisterModule(ModuleName, logModule)
}

func createConf(c *config.Config) {
	conf := &LogConf{}
	c.SetConf(config.GetModule(ModuleName), conf)
}

func initConf(c *config.Config) error {
	conf := GetConf(c)
	conf.BufSize = 10000
	dir, _ := c.Getwd()
	conf.FileName = path.Join(dir, "anys.log")
	conf.Level = logs.LevelWarning
	conf.Daily = true
	conf.Maxdays = 10

	return nil
}

func initModule(c *config.Config) error {
	return NewLogger(c)
}

func exitModule(c *config.Config) error {
	logger.Close()

	return nil
}

func GetConf(c *config.Config) *LogConf {
	conf := c.GetConf(config.GetModule(ModuleName))
	return conf.(*LogConf)
}

func logBlock(c *config.Config) error {

	return c.Block()
}

func setBufSize(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.BufSize, err = c.Int(args[1])
	return err
}

func setFileName(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.FileName = args[1]
	return nil
}

func setLevel(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Level, err = c.Int(args[1])
	return err
}

func setDaily(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Daily, err = c.Bool(args[1])
	return err
}

func setMaxdays(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Maxdays, err = c.Int(args[1])
	return err
}
