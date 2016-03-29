package log

import (
	"bytes"
	"fmt"
	"os"
	// "io"
	"io/ioutil"
	// "math/rand"
	"strings"
	"testing"

	"github.com/syndtr/goleveldb/leveldb/journal"
)

type dropper struct {
	t *testing.T
}

func (d dropper) Drop(n int, reason string) {
	fmt.Println(reason, n)
	d.t.Log(reason)
}

type jdropper struct {
	t *testing.T
}

func (d jdropper) Drop(err error) {
	d.t.Log(err)
}

func short(s string) string {
	if len(s) < 64 {
		return s
	}
	return fmt.Sprintf("%s...(skipping %d bytes)...%s", s[:20], len(s)-40, s[len(s)-20:])
}

// big returns a string of length n, composed of repetitions of partial.
func big(partial string, n int) string {
	return strings.Repeat(partial, n/len(partial)+1)[:n]
}

func TestEmpty(t *testing.T) {
	buf := new(bytes.Buffer)
	r := NewReader(buf, dropper{t}, true)
	record := []byte{}
	if err := r.ReadRecord(&record); err != nil {
		t.Fatal(err)
	}
}

func testGenerator(t *testing.T, reset func(), gen func() (string, bool)) {
	buf := new(bytes.Buffer)
	// buf, err := os.OpenFile(`E:\goWork\src\github.com/liuzhiyi/anys\cache\log\test.txt`, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0777)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer buf.Close()

	reset()
	w := NewWriter(buf)
	for {
		s, ok := gen()
		if !ok {
			break
		}
		err := w.AddRecord([]byte(s))
		if err != nil {
			t.Fatal(err)
		}
	}
	err := w.Flush()
	if err != nil {
		t.Fatal(err)
	}

	// checkByleveldb(t, reset, gen)
	reset()
	r := NewReader(buf, dropper{t}, true)
	for {
		s, ok := gen()
		if !ok {
			break
		}
		var x []byte
		err := r.ReadRecord(&x)
		if err != nil {
			t.Fatal(err)
		}
		if string(x) != s {
			t.Fatalf("got %q, want %q", short(string(x)), short(s))
		}
	}
}

func checkByleveldb(t *testing.T, reset func(), gen func() (string, bool)) {
	buf, err := os.OpenFile(`E:\goWork\src\github.com/liuzhiyi/anys\cache\log\test.txt`, os.O_RDONLY, 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer buf.Close()

	reset()
	r := journal.NewReader(buf, jdropper{t}, true, true)
	for {
		s, ok := gen()
		if !ok {
			break
		}
		rr, err := r.Next()
		if err != nil {
			t.Fatal(err, s)
		}
		x, err := ioutil.ReadAll(rr)
		if err != nil {
			t.Fatal(err, s)
		}
		if string(x) != s {
			t.Fatalf("got %q, want %q", short(string(x)), short(s))
		}
	}
}

func TestMany(t *testing.T) {
	const n = 1e5
	var i int
	reset := func() {
		i = 0
	}
	gen := func() (string, bool) {
		if i == n {
			return "", false
		}
		i++
		return fmt.Sprintf("%d.", i-1), true
	}
	testGenerator(t, reset, gen)
}
