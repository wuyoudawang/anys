package jobs

import (
	"fmt"
	"github.com/liuzhiyi/anys/pkg/utils"
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

func TestJobStore(t *testing.T) {
	container := NewContainer()
	container.Register(createTempJob("test"))
	container.Register(createTempJob("test1"))
	container.Register(createTempJob("test2"))
	container.Register(createTempJob("test3"))
	fmt.Println(utils.Key("test"), utils.Key("test1"), utils.Key("test2"))
	fmt.Println(container.Find("test1"))
	fmt.Println(container.Find("test3"))
}

func createTempJob(name string) *Job {
	return &Job{
		name: name,
		key:  utils.Key(name),
	}
}
