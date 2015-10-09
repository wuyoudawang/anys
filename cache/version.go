package cache

type VesionSet struct {
	dbname         string
	opt            *Options
	nextFileNumber uint64
	manifestNumber uint64
	lastSequence   uint64
	logNumber      uint64
	prevLogNumber  uint64
}

type Version struct {
	vset *VesionSet
	prev *Version
	next *Version
	refs int
}
