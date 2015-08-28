package jobs

type minHeap struct {
	h       []interface{}
	size    int
	maxLen  int
	compare func(a, b interface{}) int
}

func newMinHeap(length int, compare func(a, b interface{}) int) *minHeap {
	if compare == nil {
		return nil
	}

	m := new(minHeap)
	m.h = make([]interface{}, length)
	m.maxLen = length
	m.compare = compare

	return m
}

func (m *minHeap) minHeapTop() interface{} {
	return m.h[0]
}

func (m *minHeap) minHeapParent(i int) int {
	return i / 2
}

func (m *minHeap) minHeapLeft(i int) int {
	return i*2 + 1
}

func (m *minHeap) minHeapRight(i int) int {
	return (i + 1) * 2
}

func (m *minHeap) minHeapPush(elem interface{}) {
	if m.size >= m.maxLen {
		buf := make([]interface{}, m.maxLen*2)
		copy(buf, m.h[0:m.size])
		m.h = buf
		m.maxLen *= 2
	}

	m.h[m.size] = elem
	m.minHeapifyUp(m.size)
	m.size++
}

func (m *minHeap) minHeapPop() interface{} {
	elem := m.minHeapTop()
	m.h[0] = m.h[m.size-1]
	m.size--
	m.minHeapifyDown(0)
	return elem
}

func (m *minHeap) minHeapifyUp(i int) {
	if m.minHeapParent(i) == i {
		return
	}

	if m.compare(m.h[m.minHeapParent(i)], m.h[i]) > 0 {

		tmp := m.h[i]
		m.h[i] = m.h[m.minHeapParent(i)]
		m.h[m.minHeapParent(i)] = tmp
		m.minHeapifyUp(m.minHeapParent(i))
	}
}

func (m *minHeap) minHeapifyDown(i int) {
	smallest := i
	if m.minHeapLeft(i) < m.size &&
		m.compare(m.h[m.minHeapLeft(i)], m.h[smallest]) > 0 {

		smallest = m.minHeapLeft(i)
	}

	if m.minHeapRight(i) < m.size &&
		m.compare(m.h[m.minHeapRight(i)], m.h[smallest]) > 0 {

		smallest = m.minHeapRight(i)
	}

	if smallest != i {

		tmp := m.h[smallest]
		m.h[smallest] = m.h[i]
		m.h[i] = tmp
		m.minHeapifyDown(smallest)
	}
}
