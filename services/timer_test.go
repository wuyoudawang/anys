package services

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"anys/jobs"
)

type testJob struct {
}

func (t *testJob) Init(job *jobs.Job) (error, int) {
	fmt.Println("initing")
	return nil, 0
}

func (t *testJob) Run(job *jobs.Job) (error, int) {
	fmt.Println("running")
	job.GetEngine().Pending(job)
	return nil, 0
}

func (t *testJob) Exit(job *jobs.Job) (error, int) {
	fmt.Println("exiting")
	return nil, 0
}

func (t *testJob) Exception(job *jobs.Job, status int) {
	fmt.Println("has error")
}

func TestTimer(t *testing.T) {
	eng := jobs.NewEngine(1)
	eng.RegisterServer(NewTimerServer(eng), "test")
	eng.Serve()

	job := eng.NewJob(&testJob{}, "test")
	job.SetTimeout(4 * time.Second)
	eng.Pending(job)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	err := http.ListenAndServe("127.0.0.1:80", nil)
	fmt.Println(err)
}
