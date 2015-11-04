package jobs

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"anys/pkg/utils"
	"github.com/astaxie/beego/logs"
)

const (
	TIME_INFINITY = -1

	mask    = 0x000ffff
	JobDown = 1 << iota
	JobUp
	JobMulDown
	JobMulUp
	JobTimer
	JobTicker
	JobEveryTime

	StatusOk = 1 << iota
	StatusErr
	StatusRuning
	StatusExited
	StatusInitialized
	StatusActived
	StatusPending
	StatusTimeout
	StatusDied
)

type Job struct {
	e    Entity
	eng  *Engine
	name string
	args []string
	nice int

	workerIndex int
	heapIndex   int
	interval    time.Duration
	timeout     int64

	jobType    int
	status     int
	statusLock sync.Mutex

	stdin  *Input
	stderr *Output
	stdout *Output
	log    logs.LoggerInterface

	key     int64
	color   bool
	left    *Job
	right   *Job
	parent  *Job
	next    *Job
	sibling *Job
}

func isBlack(j *Job) bool {
	return (j == nil || j.color)
}

func isRed(j *Job) bool {
	return !isBlack(j)
}

func black(j *Job) {
	if j != nil {
		j.color = true
	}
}

func red(j *Job) {
	if j != nil {
		j.color = false
	}
}

func copyColor(dst *Job, src *Job) {
	if dst != nil {
		if src != nil {
			dst.color = src.color
		} else {
			dst.color = false
		}
	}
}

func NewJob(entity Entity, name string) *Job {
	return &Job{
		e:           entity,
		name:        name,
		stdin:       NewInput(os.Stdin),
		stdout:      NewOutput(os.Stdout),
		stderr:      NewOutput(os.Stderr),
		key:         utils.Key(name),
		heapIndex:   -1,
		workerIndex: -1,
	}
}

func (j *Job) SetLog(log logs.LoggerInterface) {
	j.log = log
}

func (j *Job) setEngine(eng *Engine) {
	j.eng = eng
}

func (j *Job) GetEngine() *Engine {
	return j.eng
}

func (j *Job) leftRotate(root **Job) *Job {
	if j.right == nil {
		return j
	}

	temp := j.right
	temp.parent = j.parent

	j.right = temp.left
	temp.left = j

	if j == *root {
		*root = temp
	} else if j == j.parent.left {
		j.parent.left = temp
	} else {
		j.parent.right = temp
	}

	j.parent = temp

	return j
}

func (j *Job) rightRotate(root **Job) *Job {
	if j.left == nil {
		return j
	}

	temp := j.left
	temp.parent = j.parent

	j.left = temp.right
	temp.right = j

	if j == *root {
		*root = temp
	} else if j == j.parent.left {
		j.parent.left = temp
	} else {
		j.parent.right = temp
	}

	j.parent = temp

	return j
}

func (j *Job) Find(k int64) *Job {
	if j == nil {
		return nil
	}

	if k < j.key {
		child := j.left
		return child.Find(k)
	} else if k > j.key {
		child := j.right
		return child.Find(k)
	} else {
		return j
	}
}

func (j *Job) SetNice(val int) *Job {
	j.nice = val

	return j
}

func (j *Job) SetTimeout(val time.Duration) *Job {
	j.interval = val

	return j
}

func (j *Job) Timer(val time.Duration) *Job {
	j.jobType |= JobTimer
	j.interval = val

	return j
}

func (j *Job) Ticker(val time.Duration) *Job {
	j.jobType |= JobTicker
	j.interval = val

	return j
}

func (j *Job) Until(future int64) error {
	now := time.Now().UnixNano()
	if future <= now {
		return fmt.Errorf("the future time must greater than now")
	}

	j.timeout = future
	return nil
}

func (j *Job) EveryTime(unit time.Duration, start string) *Job {
	j.jobType |= JobEveryTime
	j.interval = unit

	return j
}

func (j *Job) GetTimeout() int64 {
	now := time.Now()
	if j.interval == 0 {

		j.interval = 5 * time.Second
	}

	var interval int64
	interval = j.interval.Nanoseconds()
	if j.jobType&JobTicker > 0 {
		if j.timeout > 0 {
			if now.UnixNano() > j.timeout {
				interval = j.interval.Nanoseconds() - (now.UnixNano() - j.timeout)
			}
		}
	}

	if interval > 0 {
		return now.Add(time.Duration(interval) * time.Nanosecond).UnixNano()
	} else {
		return 0
	}
}

func (j *Job) add(root **Job) *Job {
	parent := *root
	node := root
	for *node != nil {
		parent = *node
		if j.key < parent.key {
			node = &(parent.left)
		} else if j.key > parent.key {
			node = &(parent.right)
		} else {
			return j
		}
	}
	*node = j
	j.parent = parent

	for *node != *root && isRed(parent) {
		if parent.parent != nil && parent == parent.parent.left {
			if isRed(parent.parent.right) {
				red(parent.parent)
				black(parent.parent.right)
				black(parent)
				node = &(parent.parent)
				parent = (*node).parent
			} else {
				if *node == parent.right {
					node = &((*node).parent)
					parent = (*node).parent
					(*node).parent.leftRotate(root)
				}

				black((*node).parent)
				red((*node).parent.parent)
				(*node).parent.parent.rightRotate(root)
			}
		} else {
			if parent.parent != nil && isRed(parent.parent.left) {
				red(parent.parent)
				black(parent.parent.right)
				black(parent)
				node = &(parent.parent)
				parent = (*node).parent
			} else {
				if *node == parent.right {
					node = &((*node).parent)
					if parent.parent == nil {
						parent = nil
						continue
					} else {
						parent = (*node).parent
						(*node).parent.leftRotate(root)
					}

				}

				red(parent.parent)
				black(parent)
				if parent != nil {
					parent.rightRotate(root)
				}
			}
		}
	}

	black((*root))
	return j
}

func (j *Job) Del(root **Job) {

	var tmp, subst *Job

	if j.right == nil {
		tmp = j.left
		subst = j
	} else if j.left == nil {
		tmp = j.right
		subst = j
	} else {
		subst = j.min()

		if subst.left != nil {
			tmp = subst.left
		} else {
			tmp = subst.right
		}
	}

	if subst == *root {
		*root = tmp
		black(tmp)

		j.left = nil
		j.right = nil
		j.parent = nil
		j.key = 0

		return
	}

	isRed_ := isRed(subst)

	if subst == subst.parent.left {
		subst.parent.left = tmp
	} else {
		subst.parent.right = tmp
	}

	if subst == j {
		tmp.parent = subst.parent
	} else {

		if subst.parent == j {
			tmp.parent = subst
		} else {
			tmp.parent = subst.parent
		}

		subst.left = j.left
		subst.right = j.right
		subst.parent = j.parent
		copyColor(subst, j)

		if j == *root {
			*root = subst
		} else {
			if j == j.parent.left {
				j.parent.left = subst
			} else {
				j.parent.right = subst
			}
		}

		if subst.left != nil {
			subst.left.parent = subst
		}

		if subst.right != nil {
			subst.right.parent = subst
		}
	}

	j.left = nil
	j.right = nil
	j.parent = nil
	j.key = 0

	if isRed_ {
		return
	}

	for tmp != *root && isBlack(tmp) {

		if tmp == tmp.parent.left {
			w := tmp.parent.right

			if isRed(w) {
				black(w)
				tmp.parent.leftRotate(root)
				w = tmp.parent.right
			}

			if isBlack(w.left) && isBlack(w.right) {
				red(w)
				tmp = tmp.parent
			} else {
				if isBlack(w.right) {
					red(w)
					w.rightRotate(root)
					w = tmp.parent.right
				}

				copyColor(w, tmp.parent)
				black(tmp.parent)
				tmp.parent.leftRotate(root)
				tmp = *root
			}
		} else {
			w := tmp.parent.left

			if isRed(w) {
				black(w)
				red(tmp.parent)
				tmp.parent.rightRotate(root)
				w = tmp.parent.left
			}

			if isBlack(w.left) && isBlack(w.right) {
				red(w)
				tmp = tmp.parent
			} else {
				if isBlack(w.left) {
					black(w.right)
					red(w)
					w.leftRotate(root)
					w = tmp.parent.left
				}

				copyColor(w, tmp.parent)
				black(tmp.parent)
				black(w.left)
				tmp.parent.rightRotate(root)
				tmp = *root
			}
		}
	}

	black(tmp)

}

func (j *Job) openStatus(status int) {
	j.statusLock.Lock()
	defer j.statusLock.Unlock()

	j.status |= status
}

func (j *Job) closeStatus(status int) {
	j.statusLock.Lock()
	defer j.statusLock.Unlock()

	j.status &= ^status
}

func (j *Job) GetStatus(status int) bool {
	j.statusLock.Lock()
	defer j.statusLock.Unlock()

	return j.status&status > 0
}

func (j *Job) IsRunning() bool {
	return j.GetStatus(StatusRuning)
}

func (j *Job) IsExited() bool {
	return j.GetStatus(StatusExited)
}

func (j *Job) IsActived() bool {
	return j.GetStatus(StatusActived)
}

func (j *Job) min() *Job {
	temp := j
	for temp.left != nil {
		temp = temp.left
	}

	return temp
}

func (j *Job) Extends(job *Job) {
	job.log = j.log
	job.stdin.Redirect(j.stdout)

	j.next = job
}

func (j *Job) Pending() error {
	if j.eng == nil {
		return fmt.Errorf("job '%s' using an empty enginge", j.name)
	}

	return j.eng.Pending(j)
}

func (j *Job) init() (error, int) {
	defer j.openStatus(StatusInitialized)

	err, level := j.e.Init(j)
	if level == FatalErrLvl {
		j.openStatus(StatusDied)
	}

	return err, level
}

func (j *Job) run() (error, int) {
	j.openStatus(StatusRuning)
	defer j.closeStatus(StatusRuning)

	return j.e.Run(j)
}

func (j *Job) exception(level int) {
	j.e.Exception(j, level)
}

func (j *Job) exit() (error, int) {
	defer j.openStatus(StatusExited)

	return j.e.Exit(j)
}

func (j *Job) LogInfo() error {
	if j.log != nil {
		return j.log.WriteMsg("msg", NoErrLvl)
	}

	return nil
}

func (j *Job) Clone() (*Job, error) {
	entity, err := j.e.Clone()
	if err != nil {
		return nil, err
	}

	job := &Job{
		e:           entity,
		eng:         j.GetEngine(),
		name:        j.name,
		stdin:       j.stdin,
		stdout:      j.stdout,
		stderr:      j.stderr,
		key:         j.key,
		heapIndex:   -1,
		workerIndex: -1,
		args:        j.args,
		jobType:     j.jobType,
		timeout:     j.timeout,
		interval:    j.interval,
		log:         j.log,
	}
	job.openStatus(StatusInitialized)

	return job, nil
}

func (j *Job) CmdString() string {
	return fmt.Sprintf("%s %s", j.name, strings.Join(j.args, ", "))
}

func (j *Job) StatusString() string {
	var rel string

	if j.status == StatusOk {
		rel = "OK"
	} else {
		rel = "ERROR"
	}

	return fmt.Sprintf("%s(%d)", rel, j.status)
}

type Container struct {
	root     *Job
	regLock  sync.RWMutex
	posted   *minHeap
	timers   *minHeap
	timersMt sync.Mutex
}

func NewContainer() *Container {
	c := new(Container)
	posted := newMinHeap(
		MAXGOROUTINE,
		func(a, b interface{}) int {
			jobA := a.(*Job)
			jobB := b.(*Job)

			return jobA.nice - jobB.nice
		},
	)

	timers := newMinHeap(MAXGOROUTINE*2,
		func(a, b interface{}) int {
			jobA := a.(*Job)
			jobB := b.(*Job)

			return int(jobA.timeout - jobB.timeout)
		},
	)

	c.posted = posted
	c.timers = timers

	return c
}

func (c *Container) Register(job *Job) *Container {
	c.regLock.Lock()
	job.add(&c.root)
	c.regLock.Unlock()
	job.closeStatus(mask)

	return c
}

func (c *Container) Pending(job *Job) error {

	job.openStatus(StatusPending)
	timeout := job.GetTimeout()
	if timeout != TIME_INFINITY {
		if timeout == 0 {
			job.eng.Active(job)
			return nil
		} else {
			job.timeout = timeout
			c.timersMt.Lock()
			defer c.timersMt.Unlock()
			c.timers.minHeapPush(job)
			return nil
		}
	}

	return fmt.Errorf("timeout must is positive")
}

func (c *Container) Find(name string) *Job {
	c.regLock.RLock()
	defer c.regLock.RUnlock()
	key := utils.Key(name)
	return c.root.Find(key)
}

func (c *Container) Post(job *Job) *Container {
	c.posted.minHeapPush(job)

	return c
}

func (c *Container) ProcessExpireTimer(now int64) {
	c.timersMt.Lock()
	defer c.timersMt.Unlock()

	for !c.timers.empty() {

		item := c.timers.minHeapTop()
		job := item.(*Job)

		if job.timeout > now {
			break
		}

		job.openStatus(StatusTimeout)
		c.timers.minHeapPop()

		job.eng.AddJob(job)
	}
}

func (c *Container) MinTimeout() int64 {
	c.timersMt.Lock()
	defer c.timersMt.Unlock()

	if !c.timers.empty() {
		elem := c.timers.minHeapTop()
		job := elem.(*Job)
		return job.timeout
	}

	return TIME_INFINITY
}

func (c *Container) Active(job *Job) {

	if job.heapIndex >= 0 {
		c.timersMt.Lock()
		defer c.timersMt.Unlock()
		c.timers.minHeapRemove(job.heapIndex)
	}
	job.eng.AddJob(job)
}
