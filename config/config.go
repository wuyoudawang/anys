package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	file    *os.File
	buf     *bufio.Reader
	args    []string
	ctx     []interface{}
	modules []*Module
}

func (c *Config) GetConf(m *Module) interface{} {
	ctx := c.ctx
	if m.Parent != nil {
		ctx = c.GetConf(m.Parent).([]interface{})
	}

	return ctx[m.index]
}

func (c *Config) SetConf(m *Module, conf interface{}) {
	ctx := c.ctx
	if m.Parent != nil {
		ctx = c.GetConf(m.Parent).([]interface{})
	}

	ctx[m.index] = conf
}

func (c *Config) GetArgs() []string {
	return c.args
}

func (c *Config) CheckArgs(n int) ([]string, error) {
	args := c.GetArgs()
	if len(args) != n {
		return args, fmt.Errorf("invalid number of arguments in \"%s\" directive", args[0])
	}

	return args, nil
}

func (c *Config) Int(v string) (int, error) {
	return strconv.Atoi(v)
}

func (c *Config) Bool(v string) (bool, error) {
	return strconv.ParseBool(v)
}

func (c *Config) Float64(v string) (float64, error) {
	return strconv.ParseFloat(v, 10)
}

func (c *Config) Getwd() (string, error) {
	return os.Getwd()
}
