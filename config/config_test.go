package config

import (
	"fmt"
	"testing"
)

var c = &Config{}

func Test_Parse(t *testing.T) {
	config := &Config{}
	err := config.Parse("./parse.txt")
	fmt.Println(err)
}

func Test_Include(t *testing.T) {
	err := c.Include("../anys/*")
	fmt.Println(err)
}
