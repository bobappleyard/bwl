package lexer

import (
	"io"
	"os"
	"container/vector"
	"strings"
	"bufio"
)

/* NFA operations */

func contains(s State, ss []State) bool {
	for _, x := range ss {
		if s == x {
			return true
		}
	}
	return false
}

func union(a []State, b []State) []State {
	res := make([]State, len(a) + len(b))
	l := copy(res, a)
	for _, x := range b {
		if !contains(x, res) {
			res[l] = x
			l++
		}
	}
	return res[0:l]
}

func close(from []State) []State {
	res := from
	done := false
	for !done {
		done = true
		for _, x := range res {
			ss := x.Close()
			if len(ss) != 0 {
				newres := union(res, ss)
				if len(newres) > len(res) {
					res = newres
					done = false
				}
			}
		}
	}
	return res
}

func move(from []State, c int) []State {
	to := []State {}
	for _, x := range from {
		ss := x.Move(c)
		if len(ss) != 0 {
			to = union(to, ss)
		}
	}
	return to
}

type finishState struct {
	id, pos int
}

func finished(set []State, pos int, fin *vector.Vector) {
	for _, x := range set {
		if f := x.Final(); f != FAIL {
			fin.Push(&finishState { f, pos })
		}
	}
}

/* Main interface */

const (
	_ = -iota
	FAIL
	EOF
)

type Lexer struct {
	root *BasicState
	src *bufio.Reader
	past *vector.IntVector
	pos, startPos, buf int
	eof bool
}

func New() *Lexer {
	res := new(Lexer)
	res.root = NewState()
	return res
}

func (self *Lexer) Root() *BasicState {
	return self.root
}

func (self *Lexer) Start(src io.Reader) {
	self.src = bufio.NewReader(src)
	self.past = new(vector.IntVector)
	self.buf = 0
	self.pos = 0
}

func (self *Lexer) StartString(src string) {
	self.Start(strings.NewReader(src))
}

func (self *Lexer) current() int {
	if self.buf == 0 {
		c, _, err := self.src.ReadRune()
		if err != nil {
			if err == os.EOF {
				self.eof = true
			}
			self.src = nil
			return FAIL
		}
		self.past.Push(c)
		self.buf = c
		return c
	}
	return self.buf
}

func (self *Lexer) Next() int {
	if self.src == nil {
		return FAIL
	}
	if self.eof {
		return EOF
	}
	pos := self.pos
	self.startPos = pos
	this := []State { self.root }
	fin := new(vector.Vector)
	for {
		// follow the empty transitions
		this = close(this)
		// check for finish states
		finished(this, pos, fin)
		// try to move
		c := self.current()
		if c == FAIL { break }
		next := move(this, c)
		if len(next) == 0 {
			break
		}
		// consume a char
		self.buf = 0
		pos++
		// proceed to the next state set
		this = next
	}
	res := FAIL
	for m := range fin.Iter() {
		match := m.(*finishState)
		if match.pos > self.pos ||
		  (match.pos == self.pos && match.id < res) {
			res = match.id
			self.pos = match.pos
		}
	}
	return res
}

func (self *Lexer) Eof() bool {
	return self.eof
}

func (self *Lexer) Position() int {
	return self.startPos
}

func (self *Lexer) Len() int {
	return self.pos - self.startPos
}

func (self *Lexer) Data() []int {
	return self.past.Data()[self.startPos:self.pos]
}

func (self *Lexer) String() string {
	return string(self.Data())
}


