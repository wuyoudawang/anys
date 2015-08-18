package log

import (
	"fmt"

	"anys/config"

	"github.com/astaxie/beego/logs"
)

var logger *logs.BeeLogger

func NewLogger(c *config.Config) error {
	conf := GetConf(c)
	ini := fmt.Sprintf(`{"filename":"%s", "level":%d, "maxdays":%d, "daily":%v}`,
		conf.FileName,
		conf.Level,
		conf.Maxdays,
		conf.Daily,
	)

	logger = logs.NewLogger(int64(conf.BufSize))

	return logger.SetLogger("file", ini)
}

func Trace(format string, v ...interface{}) {
	logger.Trace(format, v...)
}

func Debug(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

func Warning(format string, v ...interface{}) {
	logger.Warning(format, v...)
}

func Error(format string, v ...interface{}) {
	logger.Error(format, v...)
}

func Info(format string, v ...interface{}) {
	logger.Info(format, v...)
}
