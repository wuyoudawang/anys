package jobs

import (
	"time"
)

type Job struct {
	e      entity
	key    int
	nice   int
	color  bool
	left   *Job
	right  *Job
	parent *Job
	C      chan interface{}
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

func (j *Job) Del(c *Container) *Job {
	if j == *root {
		c.root = nil
		c.Free(j)
	} else {
		node := j
		if j.right != nil {
			temp := node.right.min()
			temp.parent = node.parent
			if j == j.parent.left {
				j.parent.left = temp
			} else {
			}
		}

		if j.left != nil {
			node.left.parent = node.parent
			if node == node.parent.left {
				node.parent.left = node.left
			} else {
				node.parent.right = node.left
			}
		}

	}

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

func (j *Job) Pause() {}

func (j *Job) Run() {}

func (j *Job) Exit() {}

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
