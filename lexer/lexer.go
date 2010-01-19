package lexer

import (
	"strings"
	"container/vector"
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
				if len(newres) != len(res) {
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
	src []int
	pos, startPos int
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

func (self *Lexer) Start(src []int) {
	self.src = src
	self.pos = 0
}

func (self *Lexer) StartString(src string) {
	self.Start(strings.Runes(src))
}

func (self *Lexer) StartBytes(src []byte) {
	self.StartString(string(src))
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
		finished(this, pos, fin)
		// check for eof
		if pos >= len(self.src) {
			self.eof = true
			break
		}
		// try to move
		next := move(this, self.src[pos])
		if len(next) == 0 {
			break
		}
		// consume a char
		pos++
		// move to the next state set
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

func (self *Lexer) String() string {
	return string(self.src[self.startPos:self.pos])
}


