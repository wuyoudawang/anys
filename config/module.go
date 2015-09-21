package config

import (
	"fmt"
)

type Module struct {
	index    int
	active   bool
	Commands []*Command
	depends  []*Module
	Parent   *Module

	Create_conf func(c *Config)
	Init_conf   func(c *Config) error

	Init_module func(c *Config) error
	Exit_module func(c *Config) error
}

var moduleMap = make(map[string]*Module)

func RegisterModule(name string, m *Module) {
	_, exists := moduleMap[name]
	if exists {
		panic("Can't overwrite module:" + name)
	}

	moduleMap[name] = m
}

func GetModule(name string) *Module {
	m := moduleMap[name]
	if m != nil {
		if m.active {
			return m
		}
	}

	return nil
}

func (c *Config) LoadModule(name string, depends ...string) {
	m := moduleMap[name]
	for _, d := range depends {
		m.depends = append(m.depends, moduleMap[d])
	}

	m.active = true
	c.modules = append(c.modules, m)
}

func (c *Config) SetParentModule() {

}

func (c *Config) SortModules() error {
	ms := make([]*Module, len(c.modules))
	copy(ms, c.modules)

	last := 0
	index := 0

	for ms[0] != nil {

		tmp := []*Module{}
		for i := 0; i < len(ms) && ms[i] != nil; i++ {

			if len(ms[i].depends) == 0 ||
				c.brokenDepends(ms[i].depends, c.modules[:last]) {

				ms[i].index = index
				tmp = append(tmp, ms[i])

				index++

				if i == len(ms) {
					ms[i] = nil
					break
				} else {
					n := copy(ms[i:], ms[i+1:])
					ms[i+n] = nil
					i--
				}
			}
		}

		if len(tmp) == 0 {
			return fmt.Errorf("exist cycle depend module")
		}

		copy(c.modules[last:], tmp)
		last += len(tmp)
	}

	if last != len(c.modules) {

		return fmt.Errorf("exist unkown error")
	}

	c.ctx = make([]interface{}, len(c.modules))

	return nil
}

func (c *Config) brokenDepends(d, b []*Module) bool {
	n := 0
	for _, dm := range d {
		for _, bm := range b {
			if dm == bm {
				n++
				break
			}
		}
	}

	return n == len(d)
}

func (c *Config) InitModules() {
	for i := 0; i < len(c.modules); i++ {
		if c.modules[i].Init_module != nil {
			c.modules[i].Init_module(c)
		}
	}
}

func (c *Config) ExitModules() {
	for i := len(c.modules) - 1; i >= 0; i-- {
		if c.modules[i].Exit_module != nil {
			c.modules[i].Exit_module(c)
		}
	}
}

func (c *Config) InitConfModules() {
	for i := 0; i < len(c.modules); i++ {
		if c.modules[i].Init_conf != nil {
			c.modules[i].Init_conf(c)
		}
	}
}

func (c *Config) CreateConfModules() {
	for i := 0; i < len(c.modules); i++ {
		if c.modules[i].Create_conf != nil {
			c.modules[i].Create_conf(c)
		}
	}
}
