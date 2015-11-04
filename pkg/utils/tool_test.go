package utils

import (
	"fmt"
	"strconv"
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
	str := `a:3:{i:1;a:9:{s:9:"starttime";s:8:"00:00:00";s:12:"firstendtime";s:8:"00:05:00";s:7:"endtime";s:8:"01:55:00";s:4:"sort";i:1;s:5:"cycle";i:300;s:7:"endsale";i:90;s:13:"inputcodetime";i:60;s:8:"droptime";i:30;s:6:"status";i:1;}i:2;a:9:{s:9:"starttime";s:8:"07:00:00";s:12:"firstendtime";s:8:"10:00:00";s:7:"endtime";s:8:"22:00:00";s:4:"sort";i:2;s:5:"cycle";i:600;s:7:"endsale";i:140;s:13:"inputcodetime";i:60;s:8:"droptime";i:30;s:6:"status";i:1;}i:3;a:9:{s:9:"starttime";s:8:"22:00:00";s:12:"firstendtime";s:8:"22:05:00";s:7:"endtime";s:8:"23:55:00";s:4:"sort";i:3;s:5:"cycle";i:300;s:7:"endsale";i:60;s:13:"inputcodetime";i:60;s:8:"droptime";i:30;s:6:"status";i:1;}}`
	data, err := PHPSerialize([]byte(str))
	if err != nil {
		panic(err)
	}

	dst := data.(map[string]interface{})
	rel := make([]map[string]interface{}, len(dst))
	sortKeys := make([]string, len(dst))
	i := 0
	for key, _ := range dst {
		pos := i
		for j := 0; j < i; j++ {
			val1, _ := strconv.Atoi(key)
			val2, _ := strconv.Atoi(sortKeys[j])
			if val1 < val2 {
				for k := i; k > j; k-- {
					sortKeys[k] = sortKeys[k-1]
				}
				pos = j
				break
			}
		}
		sortKeys[pos] = key
		i++
	}
	fmt.Println(sortKeys)
	for i, key := range sortKeys {
		rel[i] = (dst[key]).(map[string]interface{})
	}
	fmt.Println(rel)
}

func TestTime(t *testing.T) {
	test := 7 * time.May
	fmt.Println(test.String())
	timestamp, _ := StringToTime("15:04:05", "22:00:00")
	fmt.Println(timestamp)
	fmt.Println(TimeToString("15:04:05", timestamp))
}
