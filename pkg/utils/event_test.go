package utils

import (
	"fmt"
	"testing"
)

type TestEventer struct {
	Event
	name string
}

func (t *TestEventer) Say() {
	fmt.Println("hello")
	t.DispatchEvent("say", t, "liming")
}

func TestEvent(t *testing.T) {
	ter := new(TestEventer)
	ter.AddEventListener("say", func(obj *TestEventer, name string) {
		fmt.Println(obj.name)
		obj.name = name
		t.Log(name)
	})
	ter.Say()
	fmt.Println(ter.name)
}
