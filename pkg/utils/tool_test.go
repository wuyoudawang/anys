package utils

import (
	"fmt"
	"testing"
	"time"
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
	str := `a:2:{i:1;a:2:{s:9:"starttime";s:8:"droptime";s:6:"status";i:1;}i:2;a:2:{s:9:"starttime";s:8:"droptime";s:6:"status";i:1;}}`
	data, err := PHPSerialize([]byte(str))
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
}

func TestTime(t *testing.T) {
	test := 7 * time.May
	fmt.Println(test.String())
	timestamp, _ := StringToTime("15:04:05", "22:00:00")
	fmt.Println(timestamp)
	fmt.Println(TimeToString("15:04:05", timestamp))
}
