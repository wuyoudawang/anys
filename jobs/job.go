package jobs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"anys/pkg/utils"
	"github.com/astaxie/beego/logs"
)

const (
	StatusOk  = 0
	StatusErr = 1

	timeout_infinity = -1

	mask    = 0x0000000f
	extends = 1 << iota
	join
	mulExtends
	mulJoin
)

type Job struct {
	e           Entity
	name        string
	args        []string
	nice        int
	status      int
	downContact int
	workerIndex int
	timeout     time.Duration

	isTimeout bool
	isActive  bool
	isRunning bool
	isExit    bool

	stdin  *os.File
	stderr *os.File
	stdout *os.File
	log    logs.LoggerInterface
	c      chan interface{}

	key     int64
	color   bool
	left    *Job
	right   *Job
	parent  *Job
	next    *Job
	sibling *Job
}

func (j *Job) isBlack() bool {
	return (j == nil || j.color)
}

func (j *Job) isRed() bool {
	return (j != nil && !j.color)
}

func (j *Job) black() *Job {
	if j != nil {
		j.color = true
	}

	return j
}

func (j *Job) red() *Job {
	if j != nil {
		j.color = false
	}

	return j
}

func (j *Job) copyColor(job *Job) *Job {
	if j != nil {
		j.color = job.color
	}

	return j
}

func (j *Job) leftRotate(root **Job) *Job {
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
	j.timeout = val

	return j
}

func (j *Job) GetTimeout() time.Duration {
	if j.timeout == 0 {
		j.timeout = 5 * time.Second
	}

	return j.timeout
}

func (j *Job) Add(root **Job) *Job {
	parent := *root
	node := root
	for *node != nil {
		parent = *node
		if j.key < parent.key {
			*node = parent.left
		} else if j.key > parent.key {
			*node = parent.right
		} else {
			return j
		}
	}
	*node = j
	j.parent = parent

	for *node != *root && parent.isBlack() {
		if parent == parent.parent.left {
			if parent.parent.right.isRed() {
				parent.parent.red()
				parent.parent.right.black()
				parent.black()
				*node = parent.parent
				parent = (*node).parent
			} else {
				if *node == parent.right {
					(*node).parent.leftRotate(root)
					*node = (*node).parent
				}

				(*node).parent.black()
				(*node).parent.parent.red()
				(*node).parent.parent.rightRotate(root)
			}
		} else {
			if parent.parent.left.isRed() {
				parent.parent.red()
				parent.parent.right.black()
				parent.black()
				(*node) = parent.parent
				parent = (*node).parent
			} else {
				if *node == parent.right {
					(*node).parent.leftRotate(root)
					*node = (*node).parent
				}

				(*node).parent.black()
				(*node).parent.parent.red()
				(*node).parent.rightRotate(root)
			}
		}
	}

	(*root).black()
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
		tmp.black()

		j.left = nil
		j.right = nil
		j.parent = nil
		j.key = 0

		return
	}

	isRed := subst.isRed()

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
		subst.copyColor(j)

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

	if isRed {
		return
	}

	for tmp != *root && tmp.isBlack() {

		if tmp == tmp.parent.left {
			w := tmp.parent.right

			if w.isRed() {
				w.black()
				tmp.parent.leftRotate(root)
				w = tmp.parent.right
			}

			if w.left.isBlack() && w.right.isBlack() {
				w.red()
				tmp = tmp.parent
			} else {
				if w.right.isBlack() {
					w.red()
					w.rightRotate(root)
					w = tmp.parent.right
				}

				w.copyColor(tmp.parent)
				tmp.parent.black()
				tmp.parent.leftRotate(root)
				tmp = *root
			}
		} else {
			w := tmp.parent.left

			if w.isRed() {
				w.black()
				tmp.parent.red()
				tmp.parent.rightRotate(root)
				w = tmp.parent.left
			}

			if w.left.isBlack() && w.right.isBlack() {
				w.red()
				tmp = tmp.parent
			} else {
				if w.left.isBlack() {
					w.right.black()
					w.red()
					w.leftRotate(root)
					w = tmp.parent.left
				}

				w.copyColor(tmp.parent)
				tmp.parent.black()
				w.left.black()
				tmp.parent.rightRotate(root)
				tmp = *root
			}
		}
	}

	tmp.black()

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
	job.stdin = j.stdout

	j.next = job
}

func (j *Job) Pending() {}

func (j *Job) Run() {}

func (j *Job) Exit() {}

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
	root   *Job
	free   *Job
	post   *minHeap
	timers *minHeap
}

func NewContainer() *Container {
	c := new(Container)
	post := newMinHeap(
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

	c.post = post
	c.timers = timers

	return c
}

func (c *Container) Pending(job *Job) *Container {
	if job.GetTimeout() != timeout_infinity {
		c.timers.minHeapPush(job)
	}

	job.Add(&c.root)
	job.isActive = false
	job.isRunning = false
	job.isExit = false

	return c
}

func (c *Container) Find(name string) *Job {
	key := utils.Key(name)
	return c.root.Find(key)
}

func (c *Container) Active(job *Job) *Container {

	c.post.minHeapPush(job)
	job.isActive = true

	return c
}

func (c *Container) ProcessExpireTimer(now time.Duration) {
	for c.timers.minHeapTop() != nil {

		iteam := c.timers.minHeapTop()
		job := item.(*Job)

		if job.timeout > now {
			break
		}

		job.isTimeout = true
		c.timers.minHeapPop()
	}
}
