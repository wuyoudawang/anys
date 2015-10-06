package cache

type Batch struct {
	data     []byte
	sequence uint64
}

func (b *Batch) Put(key, value []byte) {

}

func (b *Batch) Delete(key []byte) {

}

func (b *Batch) Clear() {

}

func (b *Batch) insertMem() {

}
