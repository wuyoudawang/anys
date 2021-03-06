package services

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/liuzhiyi/anys/jobs"
)

type testJob struct {
}

func (t *testJob) Init(job *jobs.Job) (error, int) {
	fmt.Println("initing")
	return nil, 0
}

func (t *testJob) Run(job *jobs.Job) (error, int) {
	fmt.Println("running")
	return nil, 0
}

func (t *testJob) Exit(job *jobs.Job) (error, int) {
	fmt.Println("exiting")
	return nil, 0
}

func (t *testJob) Exception(job *jobs.Job, status int) {
	fmt.Println("has error")
}

func (t *testJob) Clone() (jobs.Entity, error) {
	return t, nil
}

func TestTimer(t *testing.T) {
	eng := jobs.NewEngine(1)
	eng.RegisterServer(NewTimerServer(eng), "test")
	eng.RegisterServer(NewLocalServer(eng, processSig, os.Interrupt), "sfe")

	job := eng.NewJob(&testJob{}, "test")
	job.Ticker(4 * time.Second)
	eng.Pending(job)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	eng.Serve()
}

func processSig(sig os.Signal) error {
	fmt.Println(sig)
	return nil
}
