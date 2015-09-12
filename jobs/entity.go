package jobs

type Entity interface {
	Init(job *Job) (error, int)    // initialize a job
	Run(job *Job) (error, int)     // run a job
	Exit(job *Job) (error, int)    // exit a job
	Clone() (Entity, error)        // exit a job
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
	init      func(job *Job) (error, int)
	run       func(job *Job) (error, int)
	exit      func(job *Job) (error, int)
	exception func(job *Job, level int)
}

func (d *Default) Init(job *Job) (error, int) {
	if d.init != nil {
		return d.init(job)
	}
	return nil, NoErrLvl
}

func (d *Default) Run(job *Job) (error, int) {
	if d.run != nil {
		return d.run(job)
	}
	return nil, NoErrLvl
}

func (d *Default) Exit(job *Job) (error, int) {
	if d.exit != nil {
		return d.exit(job)
	}
	return nil, NoErrLvl
}

func (d *Default) Clone() (Entity, error) {
	return d, nil
}

func (d *Default) Exception(job *Job, level int) {
	if d.exception != nil {
		d.exception(job, level)
	}
}
