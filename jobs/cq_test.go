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

func TestEngine() {}
