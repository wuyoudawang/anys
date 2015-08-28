package jobs

import (
	"anys/config"
)

type Entity interface {
	Init(g *config.Config) error       // initialize a job
	Run(g *config.Config) (error, int) // run a job
	Exit(g *config.Config) error       // exit a job
	Exception(g *config.Config) error  // a job in running produce a fatal error
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

func (d *Default) Init(g *config.Config) error {
	return nil
}

func (d *Default) Run(g *config.Config) (error, int) {
	return nil, NoErrLvl
}

func (d *Default) Exit(g *config.Config) error {
	return nil
}

func (d *Default) Exception(g *config.Config) error {
	return nil
}
