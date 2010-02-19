package peg

import (
	"io"
	"./lexer"
)

/*
	The position interface describes moving through a source
*/

type Position interface {
	Next() Position
	Fail() Position
	Pos() int
	Failed() bool
	Eof() bool
	Id() int
	Data() interface{}
}

// "inherit" from this and get default behaviour for everything but Next() and
// Id()
type PosDefaults struct {
	pos int
}

func (self *PosDefaults) Init(pos int) {
	self.pos = pos
}

func (self *PosDefaults) Fail() Position {
	res := new(failure)
	res.Init(self.pos)
	return res
}

func (self *PosDefaults) Pos() int {
	return self.pos
}

func (self *PosDefaults) Failed() bool {
	return false
}

func (self *PosDefaults) Eof() bool {
	return false
}

func (self *PosDefaults) Data() interface{} {
	return nil
}

type eofObj struct {
	PosDefaults
}

// this is used by the default implementation to implement failures
type failure struct {
	PosDefaults
}

func (self *failure) Next() Position {
	return self
}

func (self *failure) Fail() Position {
	return self
}

func (self *failure) Failed() bool {
	return true
}

func (self *failure) Id() int {
	return -1
}

func (self *failure) String() string {
	return string(self.pos)
}

// A position object representing the end of input
var EofObject = new(eofObj)

func (self *eofObj) Next() Position {
	return self.Fail()
}

func (self *eofObj) Eof() bool {
	return true
}

func (self *eofObj) Id() int {
	return -1
}

type lexPos struct {
	PosDefaults
	l *lexer.Lexer
	pass func(int) bool
	next Position
	id int
	data interface{}
}

func NewLex(in io.Reader, l *lexer.Lexer, pass func(int) bool) Position {
	res := new(lexPos)
	res.l = l
	l.Start(in)
	res.pass = pass
	return res.Next()
}

func (self *lexPos) Next() Position {
	if self.next != nil {
		return self.next
	}
	var n int
	for {
		n = self.l.Next()
		if n == lexer.EOF {
			return EofObject
		}
		if self.pass(n) { break }
	}
	next := new(lexPos)
	next.l = self.l
	next.pass = self.pass
	next.PosDefaults.pos = self.l.Position()
	next.id = n
	next.data = self.l.String()
	self.next = next
	return next
}

func (self *lexPos) Id() int {
	return self.id
}

func (self *lexPos) Data() interface{} {
	return self.data
}

