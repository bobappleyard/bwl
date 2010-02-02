package peg

import (
	"./lexer"
)

/*
	The position interface describes moving through a source
*/

type Value interface {}

type Position interface {
	Next() Position
	Fail() Position
	Pos() int
	Failed() bool
	Eof() bool
	Id() int
	Data() Value
}

// "inherit" from this and get default behaviour for everything but Next() and
// Id()
type PosDefaults struct {
	pos int
}

// this is used by the default implementation to implement failures
type failure struct {
	PosDefaults
}

func (self *failure) Next() Position {
	return self
}

func (self *failure) Fail(reason string) Position {
	return self
}

func (self *failure) Failed() bool {
	return true
}

func (self *failure) Id() int {
	return -1
}

func (self *PosDefaults) Fail() Position {
	res := new(failure)
	failure.PosDefault.pos = self.pos
	return res
}

func (self *PosDefaults) Failed() bool {
	return false
}

func (self *PosDefaults) Eof() bool {
	return false
}

func (self *PosDefaults) Data() Value {
	return nil
}

type eofObj struct {
	PosDefaults
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

// A position object attached to a character stream
type charPos struct {

}

// A position object attached to a lexer
type lexPos struct {
	PosDefaults
	l *lexer.Lexer
	cur int
	data Value
	next Position
}

func newLex(l *lexer.Lexer) *lexPos {
	res := new(lexPos)
	res.l = l
	return res
}

func Token(l *lexer.Lexer) Position {
	return newLex(l).Next()
}

func (self *lexPos) Next() Position {
	if self.next != nil {
		return self.next
	}
	l := self.l
	if l.Eof() {
		return EofObject
	}
	res := new(l)
	res.cur = l.Next()
	res.data = l.String()
	res.PosDefaults.pos = self.PosDefaults.pos + l.Len()
	self.next = res
	return res
}

func (self *lexPos) Data() Value {
	return self.data
}


