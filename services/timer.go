package services

import (
	"fmt"
	"sync"
	"time"

	"anys/jobs"
)

type Timer struct {
	eng            *jobs.Engine
	timer          *time.Timer
	lastMinTimeout int64
	mt             sync.Mutex
	stoper         jobs.Stoper
}

func NewTimerServer(e *jobs.Engine) *Timer {
	t := &Timer{
		eng:            e,
		timer:          time.NewTimer(6 * time.Second),
		lastMinTimeout: jobs.TIME_INFINITY,
	}

	return t
}

func (t *Timer) Stop() error {
	t.stoper.Stop()
	return nil
}

func (t *Timer) Serve() {
	for {
		if t.stoper.IsStop() {
			break
		}

		t.setTimer()
		st := <-t.timer.C
		fmt.Println(st)
		t.eng.ProcessExpireTimer(st.UnixNano())

	}

}

func (t *Timer) setTimer() {
	timeout := t.eng.MinTimeout()
	t.mt.Lock()
	defer t.mt.Unlock()

	now := time.Now().UnixNano()
	if timeout == jobs.TIME_INFINITY {
		t.lastMinTimeout = jobs.TIME_INFINITY
		t.timer.Stop()
	} else if timeout > now {
		t.lastMinTimeout = timeout
		t.timer.Reset(time.Duration(timeout - now))
	} else {
		t.eng.ProcessExpireTimer(now)
	}
}

func (t *Timer) BeforePendingJob(job *jobs.Job) error {
	return nil
}

func (t *Timer) AfterPendingJob(job *jobs.Job) error {
	t.setTimer()

	return nil
}
