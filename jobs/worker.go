package jobs

const (
	free = iota
	normal
	busy
)

type Worker struct {
	eng    *Engine
	cq     *CycleQueue
	next   *Worker
	index  int
	stoper Stoper
}

func NewWorker(i int, eng *Engine) *Worker {
	w := new(Worker)

	w.eng = eng
	w.cq = NewCycleQueueSize(minQueueSize)
	w.index = i

	return w
}

func (w *Worker) BusyLevel() int {
	return 0
}

func (w *Worker) Add(job *Job) bool {
	return w.cq.Push(job)
}

func (w *Worker) Run() {
	for {

		if w.stoper.IsStop() && w.cq.Empty() {
			return
		}

		if w.cq.Empty() {
			w.eng.Rest(w.index)
		}

		job := w.cq.Pop()
		if !job.GetStatus(StatusInitialized) {
			if err, level := job.init(); err != nil {
				job.exception(level)
			}

			if job.GetStatus(StatusDied) {
				continue
			}
		}

		isJobTime := false
		if job.jobType&JobTicker > 0 {
			newJob, err := job.Clone()
			if err != nil {
				continue
			}

			newJob.Pending()
			isJobTime = true
		}

		if err, level := job.run(); err != nil {
			job.exception(level)
		}

		if job.jobType&JobDown > 0 {
			job.Extends(job.next)
		}

		if job.jobType&JobTimer > 0 {
			job.Pending()
			isJobTime = true
		}

		if isJobTime {
			continue
		}

		job.exit()

	}
}

func (w *Worker) destory() {
	w.stoper.Stop()
}
