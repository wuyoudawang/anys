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

		now, timeout := t.setTimer()

		if timeout == jobs.TIME_INFINITY {
			t.timer.Stop()
		} else if timeout <= now {

			t.eng.ProcessExpireTimer(now)
			continue
		}

		st := <-t.timer.C
		fmt.Println(st)
		t.eng.ProcessExpireTimer(st.UnixNano())

	}

}

func (t *Timer) setTimer() (int64, int64) {
	timeout := t.eng.MinTimeout()
	now := time.Now().UnixNano()

	t.mt.Lock()
	defer t.mt.Unlock()
	if timeout == jobs.TIME_INFINITY ||
		(t.lastMinTimeout != jobs.TIME_INFINITY && t.lastMinTimeout >= now && timeout >= t.lastMinTimeout) {

		return now, timeout
	} else if timeout > now {

		t.lastMinTimeout = timeout
		t.timer.Reset(time.Duration(timeout - now))
	}

	return now, timeout
}

func (t *Timer) BeforePendingJob(job *jobs.Job) error {
	return nil
}

func (t *Timer) AfterPendingJob(job *jobs.Job) error {
	t.setTimer()
	fmt.Println("no timer")
	return nil
}
