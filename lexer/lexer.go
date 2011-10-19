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

func finished(set []State, pos int, fin []int) {
	for _, x := range set {
		f := x.Final()
		if f != FAIL && (pos > fin[1] || (pos == fin[1] && f < fin[0])) {
			fin[0] = f
			fin[1] = pos
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
	buf *vector.IntVector
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

func (self *Lexer) Start(src io.Reader) {
	self.src = bufio.NewReader(src)
	self.buf = new(vector.IntVector)
	self.pos = 0
}

func (self *Lexer) StartString(src string) {
	self.Start(strings.NewReader(src))
}

func (self *Lexer) get(pos int) int {
	for pos >= self.buf.Len() {
		c, _, err := self.src.ReadRune()
		if err != nil {
			if err == os.EOF {
				self.eof = true
			}
			self.src = nil
			return FAIL
		}
		self.buf.Push(c)
	}
	return self.buf.At(pos)
}

func (self *Lexer) Next() int {
	if self.eof {
		return EOF
	}
	if self.src == nil {
		return FAIL
	}
	fin := []int { FAIL, -1 }
	{
		pos := self.pos
		self.startPos = pos
		this := []State { self.root }
		for {
			// follow the empty transitions
			this = close(this)
			// check for finish states
			finished(this, pos, fin)
			// try to move
			c := self.get(pos)
			if c == FAIL { break }
			next := move(this, c)
			if len(next) == 0 { break }
			// consume a char
			pos ++
			// proceed to the next state set
			this = next
		}
	}
	if fin[0] != FAIL {
		self.pos = fin[1]
	}
	return fin[0]
}

func (self *Lexer) Eof() bool {
	return self.eof
}

func (self *Lexer) Pos() int {
	return self.startPos
}

func (self *Lexer) Len() int {
	return self.pos - self.startPos
}

func (self *Lexer) Data() []int {
	return []int(*self.buf.Slice(self.startPos, self.pos))
}

func (self *Lexer) String() string {
	return string(self.Data())
}


