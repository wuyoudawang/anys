package jobs

import (
	"bufio"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	deltaJobs = 5
)

type Server interface {
	Serve()
	Stop() error
	BeforePendingJob(*Job) error
	AfterPendingJob(*Job) error
}

type Stoper struct {
	isStop uint32
}

func (s *Stoper) Stop() {
	atomic.StoreUint32(&s.isStop, 1)
}

func (s *Stoper) Run() {
	atomic.StoreUint32(&s.isStop, 0)
}

func (s *Stoper) IsStop() bool {
	val := atomic.LoadUint32(&s.isStop)
	return val > 0
}

type Restroom struct {
	m sync.Mutex
	q *Queue
}

func (r *Restroom) Empty() bool {
	r.m.Lock()
	defer r.m.Unlock()

	return r.q.Empty()
}

func (r *Restroom) Pop() int {
	r.m.Lock()
	defer r.m.Unlock()

	return r.q.Pop()
}

func (r *Restroom) Push(index int) {
	r.m.Lock()
	r.q.Push(index)
	r.m.Unlock()
}

type Engine struct {
	c             *Container
	workers       []*Worker
	serversLock   sync.RWMutex
	servers       map[string]Server
	restroom      *Restroom
	list          *Worker
	wg            sync.WaitGroup
	ch            chan *Job
	maxGoroutines int
	isShutdown    bool
}

const (
	MAXGOROUTINE     = 4096
	defaultGoRoutine = 32
)

func NewEngine(n int) *Engine {
	e := new(Engine)

	e.c = NewContainer()
	e.restroom = &Restroom{}
	e.restroom.q = NewQueue(n)
	e.list = nil
	e.ch = make(chan *Job, defaultGoRoutine*16)
	e.servers = make(map[string]Server)

	e.workers = make([]*Worker, n)
	for i := 0; i < n; i++ {
		worker := NewWorker(i, e)
		e.workers[i] = worker
		e.restroom.Push(i)
	}
	e.wg.Add(n)

	return e
}

func (e *Engine) NewJob(entity Entity, name string) *Job {
	job := NewJob(entity, name)
	job.eng = e

	return job
}

func (e *Engine) Rest(index int) {
	e.restroom.Push(index)
}

func (e *Engine) Active(job *Job) {
	e.AddJob(job)
}

func (e *Engine) AddJob(job *Job) {

	if e.isShutdown {
		return
	}

	job.openStatus(StatusActived)
	if !job.IsRunning() &&
		!job.IsExited() {

		e.ch <- job
	}
}

func (e *Engine) Job(name string, args ...string) *Job {
	job := e.c.Find(name)
	if job == nil {
		return nil
	}

	job.args = args

	return job
}

func (e *Engine) start() {
	n := 1
	block := false
	var job *Job

	for {

		if n >= deltaJobs {
			n = 1
			block = true
		}

		if block {

			e.processPosted()
			job = <-e.ch
			goto do
		} else {

			select {
			case job = <-e.ch:
				goto do
			default:
				block = true
			}

			continue // this chan is empty,so direct to go to process posted list and block the goroutine
		}

	do:
		n++

		if job.IsActived() {

			if job.timeout == TIME_INFINITY {
				e.c.Post(job)
			} else {
				e.schedule(job)
			}

		} else if !job.GetStatus(StatusPending) {
			e.c.Pending(job)
		}

		block = false

	}

}

func (e *Engine) processPosted() {
	for {
		elem := e.c.posted.minHeapPop()
		if elem == nil {
			return
		}

		job := elem.(*Job)
		e.schedule(job)
	}
}

func (e *Engine) working(worker *Worker) {
	var pos **Worker
	pos = &(e.list)

	for w := e.list; w != nil; w = w.next {
		pos = &w
	}

	*pos = worker

}

func (e *Engine) schedule(job *Job) {

	index := e.restroom.Pop()

	if index != -1 {

		worker := e.workers[index]
		worker.Add(job)
		e.working(worker)
	} else {

		joined := false
		for worker := e.list; worker != nil; worker = worker.next {

			if worker.Add(job) {
				joined = true
				break
			}
		}

		if joined {
			return
		}

		size := len(e.workers)
		if size < e.getMaxGoroutines() {
			worker := NewWorker(index, e)
			e.workers = append(e.workers, worker)
			worker.Add(job)
			e.working(worker)
		}

		// job.log()
	}

}

func (e *Engine) SetMaxGoroutines(val int) {
	if val > MAXGOROUTINE {
		val = MAXGOROUTINE
	}

	e.maxGoroutines = val
}

func (e *Engine) getMaxGoroutines() int {
	if e.maxGoroutines < defaultGoRoutine {
		e.maxGoroutines = defaultGoRoutine
	}

	return e.maxGoroutines
}

func (e *Engine) Serve() {

	for _, server := range e.servers {
		go server.Serve()
	}

	for _, worker := range e.workers {
		go worker.Run()
	}

	e.start()
}

func (e *Engine) RegisterServer(server Server, name string) error {
	_, exists := e.servers[name]
	if exists {
		return fmt.Errorf("can't rewrite the same server:'%s'", name)
	}

	e.serversLock.Lock()
	e.servers[name] = server
	e.serversLock.Unlock()
	return nil
}

func (e *Engine) Shutdown() {
	e.isShutdown = true
	for _, worker := range e.workers {
		worker.destory()
	}

	e.wg.Wait()
	close(e.ch)
}

func (e *Engine) IsShutdown() bool {
	return e.isShutdown
}

func (e *Engine) ParseJob(input string) (*Job, error) {
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanWords)

	var (
		cmd []string
	)

	for scanner.Scan() {
		word := scanner.Text()
		cmd = append(cmd, word)
	}

	if len(cmd) == 0 {
		return nil, fmt.Errorf("empty command: '%s'", input)
	}

	job := e.Job(cmd[0], cmd[1:]...)
	if job == nil {
		return nil, fmt.Errorf("invaild command")
	}

	return job, nil
}

func (e *Engine) Register(job *Job) {
	e.c.Register(job)
}

func (e *Engine) Pending(job *Job) error {
	if job.IsExited() {
		return fmt.Errorf("this job has already executed '%s'", job.name)
	}

	for _, server := range e.servers {
		if err := server.BeforePendingJob(job); err != nil {
			fmt.Println(err)
		}
	}

	err := e.c.Pending(job)

	for _, server := range e.servers {
		if err := server.AfterPendingJob(job); err != nil {
			fmt.Println(err)
		}
	}

	return err
}

func (e *Engine) MinTimeout() int64 {
	return e.c.MinTimeout()
}

func (e *Engine) ProcessExpireTimer(now int64) {
	e.c.ProcessExpireTimer(now)
}
