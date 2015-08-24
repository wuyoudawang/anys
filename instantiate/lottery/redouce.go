package lottery

import (
// "fmt"

// "anys/pkg/utils"
)

func (l *Lottery) Reduce(i ...int) {
	nums := l.cvt.getGenerateFn()(l.count, i...)
	key := 0
	total := 0.0
	smallestTotal := 0.0
	smallestNum := ""
	for _, num := range nums {

		key, _ = l.cvt.integer(num)
		total = l.GetTotalReward(key)

		if smallestTotal == 0 || smallestTotal > total {
			smallestTotal = total
			smallestNum = num
		}

		if total <= l.GetMaxReward() {
			l.addNum(num)
		}
	}

	if len(l.nums) == 0 {
		l.addNum(smallestNum)
	}
}

func (l *Lottery) GetTotalReward(key int) float64 {

	total := 0.0
	if key <= 0 {
		return 0.0
	}

	var n *Number
	for i := 1; i <= l.count; i++ {
		keys := l.cvt.getChildrenKey(key, i)

		for _, key := range keys {
			n = l.t.Get(key)
			if n != nil {
				total += n.subtotal
			}
		}
	}

	// if l.cvt.getLen(key) <= 2 {
	// 	return total
	// }

	// for i := 0; ; i++ {

	// 	if b := key & l.cvt.shiftLeft(mask, i+1); b > 0 {

	// 		k := key & int(utils.NOT(int64(l.cvt.shiftLeft(mask, i+1))))

	// 		if k > 0 {
	// 			total += l.getHigherTotalReward(k)
	// 		}
	// 	}

	// 	if key&int(l.cvt.shiftLeft(int(utils.NOT(int64(mask))), i+1)) == 0 {
	// 		break
	// 	}
	// }

	return total
}

func (l *Lottery) getSubestTotalReward(key int) float64 {

	total := 0.0
	for i := 1; key > 0; key = l.cvt.shiftRight(key, i) {
		if k := key & mask; k > 0 {

			k = l.cvt.shiftLeft(k, i)

			n := l.t.Get(k)
			if n != nil {
				total += n.subtotal
			}
		}
		i++
	}

	return total
}
