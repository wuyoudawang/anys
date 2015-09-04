package jobs

import (
	"io"
	"sync"
)

type Output struct {
	sync.Mutex
	dests []io.Writer
	tasks sync.WaitGroup
	used  bool
}

func NewOutput(w io.Writer) *Output {
	o := &Output{}
	o.SetStdout(w)

	return o
}

func (o *Output) Used() bool {
	o.Lock()
	defer o.Unlock()
	return o.used
}

func (o *Output) Add(dst io.Writer) {
	o.Lock()
	defer o.Unlock()
	o.dests = append(o.dests, dst)
}

func (o *Output) SetStdout(dst io.Writer) {
	o.Close()
	o.Lock()
	defer o.Unlock()
	o.dests = []io.Writer{dst}
}

func (o *Output) Stdout() io.Writer {
	return o.dests[0]
}

func (o *Output) Redirect() io.Reader {
	r, w := io.Pipe()
	o.SetStdout(w)
	return r
}

func (o *Output) Close() error {
	o.Lock()
	defer o.Unlock()
	var firstErr error
	for _, dst := range o.dests {
		if closer, ok := dst.(io.Closer); ok {
			err := closer.Close()
			if err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	o.tasks.Wait()
	o.dests = nil
	return firstErr
}

func (o *Output) Write(p []byte) (n int, err error) {
	o.Lock()
	defer o.Unlock()
	o.used = true
	var firstErr error
	for _, dst := range o.dests {
		_, err := dst.Write(p)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return len(p), firstErr
}

type Input struct {
	src io.Reader
	sync.Mutex
}

func NewInput(r io.Reader) *Input {
	return &Input{src: r}
}

func (i *Input) Read(p []byte) (n int, err error) {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()
	if i.src == nil {
		return 0, io.EOF
	}
	return i.src.Read(p)
}

func (i *Input) Close() error {
	if i.src != nil {
		if closer, ok := i.src.(io.Closer); ok {
			return closer.Close()
		}
	}
	return nil
}

func (i *Input) Redirect(src *Output) {
	i.Mutex.Lock()
	defer i.Mutex.Unlock()

	r := src.Redirect()
	i.src = r
}
