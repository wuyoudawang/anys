package lottery

import (
	"github.com/liuzhiyi/anys/instantiate/lottery/model"
	"github.com/liuzhiyi/anys/jobs"
)

type lotteryJob struct {
	lty *Lottery
}

func (lj *lotteryJob) Init(job *jobs.Job) (error, int) {
	return nil, 0
}

func (lj *lotteryJob) Run(job *jobs.Job) (error, int) {
	lj.lty.Process()
	lj.lty.Reset()
	return nil, 0
}

func (lj *lotteryJob) Exit(job *jobs.Job) (error, int) {

	return nil, 0
}

func (lj *lotteryJob) Clone() (jobs.Entity, error) {
	lty, err := lj.lty.Clone()
	return &lotteryJob{
		lty: lty,
	}, err
}

func (lj *lotteryJob) Exception(job *jobs.Job, status int) {

}

type issueJob struct {
	lotteryModel *model.Lottery
}

func (ij *issueJob) Init(job *jobs.Job) (error, int) {
	return nil, 0
}

func (ij *issueJob) Run(job *jobs.Job) (error, int) {
	ij.lotteryModel.AutoClearIssues()
	if err := ij.lotteryModel.AutoGenerateIssues(); err != nil {

	}
	return nil, 0
}

func (ij *issueJob) Exit(job *jobs.Job) (error, int) {

	return nil, 0
}

func (ij *issueJob) Clone() (jobs.Entity, error) {
	return ij, nil
}

func (lj *issueJob) Exception(job *jobs.Job, status int) {

}

type issueErrorJob struct {
	ltym *model.Lottery
}

func (iej *issueErrorJob) Init(job *jobs.Job) (error, int) {
	return nil, 0
}

func (iej *issueErrorJob) Run(job *jobs.Job) (error, int) {
	iej.ltym.ProcessIssueError()
	return nil, 0
}

func (iej *issueErrorJob) Exit(job *jobs.Job) (error, int) {

	return nil, 0
}

func (iej *issueErrorJob) Clone() (jobs.Entity, error) {
	return iej, nil
}

func (iej *issueErrorJob) Exception(job *jobs.Job, status int) {

}
