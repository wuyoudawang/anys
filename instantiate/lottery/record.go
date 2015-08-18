package lottery

type record struct {
	projectid int64
	amount    float64
	currPunts int64
}

func NewRecord(id int64, amount float64) *record {
	return &record{
		projectid: id,
		amount:    amount,
	}
}

func (r *record) getCurrPunts() int64 {
	return r.currPunts
}
