package utils

import (
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
