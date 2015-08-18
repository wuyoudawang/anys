package jobs

import (
	"anys/config"
)

type Entity interface {
	Init(g *config.G) error       // initialize a job
	Run(g *config.G) (error, int) // run a job
	Exit(g *config.G) error       // exit a job
	Exception(g *config.G) error  // a job in running produce a fatal error
}

// errors level
var (
	NoErrLvl = iota
	NoteErrLvl
	WarnErrLvl
	FatalErrLvl
)

type Default struct {
}

func (d *Default) Init(g *config.G) error {
	return nil
}

func (d *Default) Run(g *config.G) (error, int) {
	return nil, NoErrLvl
}

func (d *Default) Exit(g *config.G) error {
	return nil
}

func (d *Default) Exception(g *config.G) error {
	return nil
}
