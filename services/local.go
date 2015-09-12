package services

import (
	"fmt"
	"os"
	"os/signal"

	"anys/jobs"
)

type LocalServer struct {
	ch      chan os.Signal
	eng     *jobs.Engine
	stoper  jobs.Stoper
	handler func(sig os.Signal) error
}

func NewLocalServer(eng *jobs.Engine, handler func(sig os.Signal) error, sig ...os.Signal) *LocalServer {
	l := new(LocalServer)
	l.eng = eng
	l.handler = handler
	l.ch = make(chan os.Signal, 1)
	signal.Notify(l.ch, sig...)

	return l
}

func (l *LocalServer) Stop() error {
	l.stoper.Stop()
	signal.Stop(l.ch)

	return nil
}

func (l *LocalServer) Serve() {
	for {
		if l.stoper.IsStop() {
			break
		}

		sig := <-l.ch
		fmt.Println(l.eng.MinTimeout())
		err := l.handler(sig)
		if err != nil {
			fmt.Println(err, sig)
		}
	}
}

func (l *LocalServer) BeforePendingJob(job *jobs.Job) error {
	return nil
}

func (l *LocalServer) AfterPendingJob(job *jobs.Job) error {

	return nil
}
