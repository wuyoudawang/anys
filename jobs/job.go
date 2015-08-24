package jobs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	StatusOk  = 0
	StatusErr = 1
)

type Job struct {
	e      entity
	name   string
	args   []string
	nice   int
	status int

	stdin  *os.File
	stderr *os.File
	stdout *os.File
	log    logs.LoggerInterface
	c      chan interface{}

	key    int
	color  bool
	left   *Job
	right  *Job
	parent *Job
}

func (j *Job) isBlack() bool {
	return (j == nil || j.color)
}

func (j *Job) isRed() bool {
	return !j.isBlack()
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

func (j *Job) Find(k int) *Job {
	if j == nil {
		return nil
	}

	if k < j.key {
		child := j.left
		return child.Find(name)
	} else if k > j.key {
		child := j.right
		return child.Find(name)
	} else {
		return j
	}
}

func (j *Job) Push() *Job {

}

func (j *Job) Add(root **Job, job *Job) *Job {
	parent := *root
	node := root
	for *node != nil {
		parent = *node
		if job.key < parent.key {
			node = parent.left
		} else if job.key > parent.key {
			node = parent.right
		} else {
			return j
		}
	}
	*node = job
	job.parent = parent

	for *node != root && parent.isBlack() {
		if parent == parent.parent.left {
			if parent.parent.right.isRed() {
				parent.parent.red()
				parent.parent.right.black()
				parent.black()
				*node = parent.parent
				parent = *node.parent
			} else {
				if *node == parent.right {
					*node.parent.leftRotate(root)
					*node = *node.parent
				}

				*node.parent.black()
				*node.parent.parent.red()
				*node.parent.parent.rightRotate(root)
			}
		} else {
			if parent.parent.left.isRed() {
				parent.parent.red()
				parent.parent.right.black()
				parent.black()
				*node = parent.parent
				parent = *node.parent
			} else {
				if *node == parent.right {
					*node.parent.leftRotate(root)
					*node = *node.parent
				}

				*node.parent.black()
				*node.parent.parent.red()
				*node.parent.rightRotate(root)
			}
		}
	}

	root.black()
	return j
}

func (j *Job) Del(root **Job) *Job {

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

	return j
}

func (j *Job) min() *Job {
	temp := j
	for temp.left != nil {
		temp = temp.left
	}

	return temp
}

func (j *Job) Pending() {}

func (j *Job) Run() {}

func (j *Job) Exit() {}

func (j *Job) CmdString() string {
	return fmt.Sprintf("%s %s", job.name, strings.Join(job.args, ", "))
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

func key(name string) int {
	k := 0
	for i, v := range name {
		k += i * int(v)
	}

	return k
}

type Container struct {
	root     *Job
	free     []*Job
	post     minHeap
	timers   minHeap
	postSize int
	postLen  int
}

func (c *Container) Pending(job *Job) *Container {
	c.post.minHeapPush(job)

	return c
}

func (c *Container) Run() {

}
