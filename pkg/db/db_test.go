package db

import (
	"fmt"
	"testing"

	"github.com/liuzhiyi/anys/config"
)

var c = &config.Config{}

func Test_Db(t *testing.T) {
	c.LoadModule("db")
	c.SortModules()

	c.CreateConfModules()
	c.InitConfModules()
	c.Parse("db.conf")

	fmt.Println(GetConf(c))
}
