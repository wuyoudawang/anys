package lottery

import (
	"fmt"
	"testing"
)

var c = newConvert(1, 9)

func Test_Convert(t *testing.T) {

	fmt.Println(c.number(107796))
	fmt.Println(c.getLen(107796))
	// lotteryVal, _ := c.integer("0909")
	// fmt.Println(lotteryVal)
	// fmt.Println(c.number(lotteryVal))
	// fmt.Println(c.getGenerateFn()(5, 1, 1, 2))

}

func TestPermutation(t *testing.T) {
	permu := c.getPermutation([]string{"1", "2", "3", "4", "5", "6"}, 5)
	fmt.Println(len(permu), permu)
}

func TestCombination(t *testing.T) {
	com := c.getCombination([]string{"1", "2", "3", "4", "5", "6", "8"}, 4)
	fmt.Println(len(com), com)
}

func TestFomart(t *testing.T) {
	set := []string{"1", "2", "3", "4", "5", "6", "8"}
	set = c.formatString(set...)
	fmt.Println(set)
}

func TestRepeatNum(t *testing.T) {
	set := []string{"1", "2", "3"}
	rel := c.repeatNum(set, "4", 2)
	fmt.Println(rel)
}

func TestSumCom(t *testing.T) {
	rel := c.getSumCom(6, 4)
	fmt.Println(rel)
}

func TestSelection(t *testing.T) {
	set := []string{"1", "2"}
	rel := c.getSelection(set, 4)
	fmt.Println("selection", rel)
}
