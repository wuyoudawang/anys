package jobs

import (
	"sync"
	"time"
)

type Engine struct {
	c          *Container
	works      []*Worker
	restroom   *Queue
	m          sync.Mutex
	wg         sync.WaitGroup
	ch         chan []*Job
	isShutdown bool
}

const (
	MAXGOROUTINE     = 4096
	defaultGoRoutine = 32
)

func NewEngine(n int) *Engine {
	e := new(Engine)

	e.works = make([]*Worker, n)
	for i := 0; i < n; i++ {
		worker := NewWorker(i, e)
		e.works[i] = worker
	}
	e.wg.Add(n)

	e.restroom = NewQueue(n)

	return e
}

func (e *Engine) Rest(index int) {
	e.restroom.Push(index)
}

func (e *Engine) AddJob(job *Job) {

	if e.isShutdown {
		return
	}

	if !job.isActive &&
		!job.isRunning &&
		!job.isExit {

		e.ch <- job
	}
}

func (e *Engine) start() {
	now := time.Now().Unix()
	last := time.Now().Unix()

	for _, job := range ch {
		now = time.Now().Unix()

		if job.isActive {
			e.c.Active(job)
		} else if !job.isRunning {
			e.c.Pending(job)
		}

		if now-last > 0 {
			e.c.ProcessExpireTimer(now)
		}

		last = now
	}
}

func (e *Engine) Serve() {
	go e.start()
}

func (e *Engine) Shutdown() {
	e.isShutdown = true
	for _, worker := range e.works {
		worker.Destory()
	}

	e.wg.Wait()
	close(e.ch)
}
