package db

import (
	"fmt"

	"anys/config"
	_ "github.com/go-sql-driver/mysql"
	db "github.com/liuzhiyi/go-db"
)

func initDB(c *config.Config) {
	conf := GetConf(c)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s",
		conf.User,
		conf.Passwd,
		conf.Addr,
		conf.Port,
		conf.Dbname,
		conf.Charset,
	)

	db.F.InitDb("mysql", dsn, "")
}
