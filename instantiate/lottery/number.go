package lottery

import (
// "fmt"
// "anys/instantiate/lottery/model"
)

type Number struct {
	bets     string
	key      int
	index    int
	subtotal float64
	records  map[int64]*record
}

func (n *Number) Add(id int64, amount float64) {
	r, exists := n.records[id]
	if !exists {
		r = NewRecord(id, amount)
	} else {
		r.amount += amount
	}

	r.currPunts++

	n.subtotal += amount
	n.records[id] = r
}

func (n *Number) getRecordCurrPunts(id int64) int64 {
	r, exists := n.records[id]
	if !exists {
		return 0
	}

	return r.currPunts
}
