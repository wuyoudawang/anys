package lottery

import (
// "github.com/liuzhiyi/anys/instantiate/lottery/model"
)

type Table struct {
	lty       *Lottery
	container []map[int]*Number
	spans     int
}

func NewTable(lty *Lottery, n int) *Table {
	t := &Table{
		lty:   lty,
		spans: n,
	}

	for i := 0; i < n; i++ {
		t.container = append(t.container, make(map[int]*Number))
	}

	return t
}

func (t *Table) Clone(lty *Lottery) *Table {
	return NewTable(lty, len(t.container))
}

func (t *Table) Hash(key int) int {
	return key % t.spans
}

func (t *Table) Get(key int) *Number {
	i := t.Hash(key)
	return t.container[i][key]
}

func (t *Table) Add(n *Number) {
	t.container[n.index][n.key] = n
}

func (t *Table) AddRecord(key int, id int64, amount float64) {
	i := t.Hash(key)
	n, exist := t.container[i][key]
	if !exist {
		bets := t.lty.cvt.number(key)
		n = &Number{
			bets:  bets,
			key:   key,
			index: i,
		}

		n.records = make(map[int64]*record)

		t.Add(n)
	}

	n.Add(id, amount)
}

func (t *Table) reset() {
	for i, _ := range t.container {
		t.container[i] = make(map[int]*Number)
	}
}
