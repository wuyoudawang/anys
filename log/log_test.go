package log

import (
	"fmt"
	"os"
	"testing"
	"time"

	"anys/config"
)

var c = &config.Config{}

func Test_Log(t *testing.T) {
	c.LoadModule("log")
	c.SortModules()

	c.CreateConfModules()
	c.InitConfModules()
	err := c.Parse("../conf/anys.conf")
	fmt.Println(err)
	fmt.Println(NewLogger(c))
	Info("hello world!")
	Error("error")
	Debug("debug")
	Warning("warning")
	Trace("trace")
	time.Sleep(time.Second * 4)
	os.Remove(GetConf(c).FileName)
}
