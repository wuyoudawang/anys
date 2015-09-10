package jobs

import (
	"fmt"
	"testing"
	"time"
)

func TestCQ(t *testing.T) {
	job := new(Job)
	job.name = "test"
	cq := NewCycleQueueSize(3)

	go func() {
		for {
			j := cq.Pop()
			fmt.Println(j.name)
		}
	}()

	fmt.Println(len(cq.buf))
	for i := 0; i < 18; i++ {
		ok := cq.Push(job)
		fmt.Println(ok)
		time.Sleep(1 * time.Second)
	}
}

func TestHeap(t *testing.T) {
	heap := newMinHeap(
		16,
		func(a, b interface{}) int {
			jobA := a.(*Job)
			jobB := b.(*Job)

			return int(jobA.timeout - jobB.timeout)
		},
	)
	for i := 0; i < 4; i++ {
		job := new(Job)
		job.name = fmt.Sprintf("%d", i)
		job.timeout = time.Now().UnixNano() + int64(i)
		heap.minHeapPush(job)
	}
	fmt.Println(heap.minHeapPop())
	fmt.Println(heap.minHeapPop())
	fmt.Println(heap.minHeapPop())
	fmt.Println(heap.minHeapPop())
	fmt.Println(heap.h)

}
