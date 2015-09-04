package jobs

type Entity interface {
	Init(job *Job) (error, int)    // initialize a job
	Run(job *Job) (error, int)     // run a job
	Exit(job *Job) (error, int)    // exit a job
	Exception(job *Job, level int) // a job in running produce a fatal error
}

// errors level
const (
	NoErrLvl = iota
	NoteErrLvl
	WarnErrLvl
	FatalErrLvl
)

type Default struct {
}

func (d *Default) Init(job *Job) (error, int) {
	return nil, NoErrLvl
}

func (d *Default) Run(job *Job) (error, int) {
	return nil, NoErrLvl
}

func (d *Default) Exit(job *Job) (error, int) {
	return nil, NoErrLvl
}

func (d *Default) Exception(job *Job, level int) {

}
