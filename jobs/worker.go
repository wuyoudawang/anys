package jobs

type Worker struct {
	eng     *Engine
	cq      *CycleQueue
	index   int
	destory bool
}

func NewWorker(i int, eng *Engine) *Worker {
	w := new(Worker)

	w.eng = eng
	w.cq = NewCycleQueueSize(minQueueSize)
	w.destory = false
	w.index = i

	return w
}

func (w *Worker) Run() {
	go func() {
		for {

			if w.destory && w.cq.Empty() {
				return
			}

			if w.cq.Empty() {
				w.eng.Rest(w.index)
			}

			job := w.cq.Pop()
			job.Run()

			if job.downContact&extends > 0 {
				job.Extends(job.next)
			}

		}
	}()
}

func (w *Worker) Destory() {
	w.destory = true
}
