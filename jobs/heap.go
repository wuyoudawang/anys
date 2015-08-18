package jobs

type minHeap struct {
	h      []*Job
	size   int
	maxLen int
}

func (m *minHeap) minHeapTop() *Job {
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

func (m *minHeap) minHeapPush(job *Job) {
	m.h[m.size] = job
	m.minHeapifyUp(m.size)
	m.size++
}

func (m *minHeap) minHeapPop() *Job {
	job := m.minHeapTop()
	m.h[0] = m.h[m.size-1]
	m.size--
	m.minHeapifyDown(0)
	return job
}

func (m *minHeap) minHeapifyUp(i int) {
	if m.minHeapParent(i) == i {
		return
	}

	if m.h[m.minHeapParent(i)].key > m.h[i].key {
		tmp := m.h[i]
		m.h[i] = h[m.minHeapParent(i)]
		m.h[m.minHeapParent(i)] = tmp
		m.minHeapifyUp(m.minHeapParent(i))
	}
}

func (m *minHeap) minHeapifyDown(i int) {
	smallest := i
	if m.minHeapLeft(i) < m.size && m.h[m.minHeapLeft(i)] < m.h[smallest] {
		smallest = m.minHeapLeft(i)
	}

	if m.minHeapRight(i) < m.size && m.h[m.minHeapRight(i)] < m.h[smallest] {
		smallest = m.minHeapRight(i)
	}

	if smallest != i {
		tmp := m.h[smallest]
		m.h[smallest] = m.h[i]
		m.h[i] = tmp
		m.minHeapifyDown(smallest)
	}
}
