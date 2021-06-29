package runes

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"sync"
)

// feeder is a generic implementation of the output interfaces of Feeder
type Feeder struct {
	sync.Mutex

	in  io.RuneReader
	out []rune
	sz  []int
	err error
}

// NewFeederBytes creates a new Feeder using an slice of bytes as input
func NewFeederBytes(b []byte) *Feeder {
	return NewFeeder(bytes.NewReader(b))
}

// NewFeederString creates a new Feeder using a string as input
func NewFeederString(s string) *Feeder {
	return NewFeeder(strings.NewReader(s))
}

// NewFeeder creates a new Feeder using a Reader as input
func NewFeeder(in io.Reader) *Feeder {
	rd, ok := in.(io.RuneReader)
	if !ok {
		rd = bufio.NewReader(in)
	}
	return &Feeder{in: rd}
}

// Skip drops n runes from the head of the buffer
func (f *Feeder) Skip(n int) (int, bool) {
	f.Lock()
	defer f.Unlock()

	if l := f.skip(n); l > 0 {
		return l, true
	} else {
		return 0, false
	}
}
func (f *Feeder) skip(n int) int {
	if l := len(f.out); l > n {
		f.out = f.out[n:]
		f.sz = f.sz[n:]
		return l - n
	} else {
		f.out = f.out[:0]
		f.sz = f.sz[:0]
		return 0
	}
}

// ReadRune returns the next rune
func (f *Feeder) ReadRune() (r rune, size int, err error) {
	f.Lock()
	defer f.Unlock()

	if f.atLeast(1) {
		r = f.out[0]
		size = f.sz[0]

		f.skip(1)
	}

	err = f.Err()
	return
}

// AtLeast blocks until there are at least n runes on the buffer, or an error or EOF has occurred
func (f *Feeder) AtLeast(n int) (out []rune, err error) {
	f.Lock()
	defer f.Unlock()

	if !f.atLeast(n) {
		err = f.err
	}

	if len(f.out) > 0 {
		out = f.out
	}

	return
}

func (f *Feeder) atLeast(n int) bool {
	for len(f.out) < n {
		r, size, err := f.in.ReadRune()
		if err != nil && f.err == nil {
			// store first error
			f.err = err
		}

		if size > 0 {
			f.out = append(f.out, r)
			f.sz = append(f.sz, size)
		} else if f.err != nil {
			break
		}
	}

	return len(f.out) >= n
}

// Currently buffered runes
func (f *Feeder) Buffered() []rune {
	return f.out
}

// Count of currently buffered runes
func (f *Feeder) Len() int {
	return len(f.out)
}

// Feeder has reached EOF
func (f *Feeder) EOF() bool {
	return f.err == io.EOF
}

// Feeder encountered an error
func (f *Feeder) Err() error {
	if f.err == io.EOF {
		return nil
	}
	return f.err
}
