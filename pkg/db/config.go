package db

import (
	// "fmt"

	"anys/config"
	// "gopkg.in/gorp.v1"
)

const (
	ModuleName = "db"
)

type DnsConf struct {
	Addr    string
	Port    int
	User    string
	Passwd  string
	Dbname  string
	Charset string
}

var (
	commands = []*config.Command{
		{config.BLOCK, "db", dbBlock},
		{config.MORE1, "addr", setAddr},
		{config.MORE1, "port", setPort},
		{config.MORE1, "user", setUser},
		{config.MORE1, "passwd", setPasswd},
		{config.MORE1, "dbname", setDbname},
		{config.MORE1, "charset", setCharset},
	}

	dbModule = &config.Module{
		Commands: commands,

		Create_conf: createConf,
		Init_conf:   initConf,

		Init_module: initModule,
	}
)

func init() {
	config.RegisterModule(ModuleName, dbModule)
}

func createConf(c *config.Config) {
	conf := &DnsConf{}
	c.SetConf(config.GetModule(ModuleName), conf)
}

func initConf(c *config.Config) error {
	conf := GetConf(c)
	conf.Addr = "127.0.0.1"
	conf.Port = 3306
	conf.User = "root"
	conf.Passwd = ""
	conf.Dbname = "test"

	return nil
}

func initModule(c *config.Config) error {
	initDB(c)

	return nil
}

func GetConf(c *config.Config) *DnsConf {
	conf := c.GetConf(config.GetModule(ModuleName))
	return conf.(*DnsConf)
}

func dbBlock(c *config.Config) error {

	return c.Block()
}

func setAddr(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Addr = args[1]
	return nil
}

func setPort(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Port, err = c.Int(args[1])
	if err != nil {
		return err
	}

	return nil
}

func setUser(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.User = args[1]
	return nil
}

func setPasswd(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Passwd = args[1]
	return nil
}

func setDbname(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Dbname = args[1]
	return nil
}

func setCharset(c *config.Config) error {
	args, err := c.CheckArgs(2)
	if err != nil {
		return err
	}

	conf := GetConf(c)
	conf.Charset = args[1]
	return nil
}
