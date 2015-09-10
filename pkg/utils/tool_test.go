package utils

import (
	"fmt"
	"testing"
)

func TestUcWords(t *testing.T) {
	str := "hello word!i'm; jim"
	rel := "Hello Word!I'm; Jim"

	if rel != UcWords(str) {
		t.Fatal(UcWords(str))
	}
}

func TestNOT(t *testing.T) {
	i := 1
	rel := 0
	if int64(rel) != NOT(int64(i)) {
		// t.Fatal(NOT(int64(i)))
		t.Fatal(^1)
	}
}

func TestPHPSerialize(t *testing.T) {
	str := `a:1:{i:1;a:9:{s:9:"starttime";s:8:"droptime";s:6:"status";a:1:{i:0;i:3;}}}`
	data, err := PHPSerialize([]byte(str))
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}
