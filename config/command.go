package config

import (
	"path/filepath"
	"strings"
)

type Command struct {
	T       int
	Name    string
	Handler func(c *Config) error
}

var (
	BLOCK  = 0x00000100
	NUMBER = 0x000000ff
	ANY    = 0x00000400
	FLAG   = 0x00000200
	MORE1  = 0x00000800
	MORE2  = 0x00001000
	MULTI  = 0x00000000
)

func (c *Config) Block() error {
	return c.Parse("")
}

func (c *Config) Include(name string) error {
	if !strings.ContainsAny(name, "*?[") {
		return c.Parse(name)
	}

	matches, err := filepath.Glob(name)
	if err != nil {
		return err
	}

	for _, match := range matches {
		err = c.Parse(match)
		if err != nil {
			return err
		}
	}
	return nil
}

type caller struct {
	args    []string
	handler func(c *Config) error
}

func (c *Config) DelayCaller(handler func(c *Config) error) {
	dc := &caller{
		handler: handler,
		args:    c.args,
	}
	c.delayCall = append(c.delayCall, dc)
}
