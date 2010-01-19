package peg

/*
	PEG matching
*/

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"container/vector"
	"strings"
)

/*
	The position interface defines moving through a source
*/

type Value interface {}

type Position interface {
	Next() Position
	Fail() Position
	Failed() bool
	Eof() bool
	Data() Value
}

// "inherit" from this and get default behaviour for everything but Next()
type PosDefaults struct {}

// this is used by the default implementation to implement failures
type failure struct {
	PosDefaults
	reason string
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

func (self *failure) String() string {
	return reason
}


func (self *PosDefaults) Fail(reason string) Position {
	res := new(failure)
	res.reason = reason
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

var EofObject = func() Position {
	res := new(eofObj)
	return res
}()

func (self *EofObj) Next() Value {
	return self.Fail()
}

func (self *EofObj) Eof() bool {
	return true
}



/*
	An expression interface
	
	Expresssions match a position in the source, or not. The Match method
	should return the position in the input just after the end of the region
	matched.
	
	A Matcher is a function that looks like an expression object. This 
	simplifies the creation of new peg forms where being able to treat the 
	expression as a slice or struct or what have you isn't desired.
*/

type Expr interface {
	Match(m Position) (Position, Value)
}

type Matcher func(m Position) (Position, Value)

func (e Matcher) Match(m Position) (Position, Value) {
	return e(m)
}

/*
	Primitive Expressions
*/

var Any = Matcher(func(m Position) (Position, Value) {
	return m.Next(), m.Data()
})

var None = Matcher(func(m Position) (Position, Value) {
	return m.Fail(), nil
})

var Eof = Matcher(func(m Position) (Position, Value) {
	if m.Eof() {
		return m, nil
	}
	return m.Fail(), nil
})

func Terminal(c int) Expr {
	return Matcher(func(m Position) (Position, Value) {
		if m.char == c {
			return m.Next(), m.Data()
		}
		return m.Fail(), nil
	})
}

/*
	Combining Expressions
*/

type And []Expr

func (e And) Match(m Position) (Position, Value) {
	res := make([]Value, len(e))
	for i, x := range e {
		if m.Failed() {
			return m, nil
		}
		m, res[i] = x.Match(cur)
	}
	return m, res
}

type Or []Expr

func (e Or) Match(m Position) (cur Position, res Value) {
	for i, x := range e {
		cur, res = x.Match(m)
		if !cur.Failed() {
			return
		}
	}
	return
}

type ExtensibleExpr struct {
	es *vector.Vector
}

func Extensible () *ExtensibleExpr {
	return &ExtensibleExpr { new(vector.Vector) }
}

func (self *Extensible) Add(e Expr) {
	self.es.Push(e)
}

func (self *Extensible) Match(m Position) (Position, Value) {
	for _, e := range self.es.Data() {
		n, res := e.(Expr).Match(m)
		if !n.Failed() {
			return n, res
		}
	}
	return m.Fail(), nil
}

/*
	Expression quantifiers
*/

type quantifiedExpr struct {
	e Expr
	min, max int
}

func (e *quantifiedExpr) Match(m Position) (Position, Value) {
	var item Value
	cur := m
	res := new(v.Vector)
	// guaranteed minimum
	for i := 0; i < e.min; i++ {
		cur, item = e.e.Match(cur)
		if cur.Failed() {
			return cur, nil
		}
		res.Push(item)
	}
	last := cur
	// optional (up to a maximum)
	for i := e.min; e.max == -1 || i < e.max; i++ {
		cur, item = e.e.Match(last)
		if cur.Failed() {
			return last, res
		}
		res.Push(item)
		last = cur
	}
	return cur, res.Data()
}

func Quantify(e Expr, min, max int) Expr {
	return &quantifiedExpr { e, min, max }
}

func Option(e Expr) Expr {
	return Modify(e, 0, 1)
}

func Repeat(e Expr) Expr {
	return Modify(e, 0, -1)
}

func Multi(e Expr) Expr {
	return Modify(e, 1, -1)
}

/*
	Lookahead
*/

func Ensure(e Expr) Expr {
	return Matcher(func (m Position) (Position, Value) {
		n, _ := e.Match(m)
		if n.Failed() {
			return n, nil
		}
		return m, nil
	})
}

func Prevent(e Expr) Expr {
	return Matcher(func (m Position) (Position, Value) {
		n, _ := e.Match(m)
		if n.Failed() {
			return m, nil
		}
		return m.Fail(), nil
	})
}

func RepeatUntil(e, end Expr) Expr {
	return Repeat(Select(And { Prevent(end), e }, 1))
}

/*
	Recursion support
	
	Note that one must guard against left recursion.
*/

type RecursiveExpr struct {
	e Expr
	set bool
}

func Recursive() *RecursiveExpr {
	return new(RecursiveExpr)
}

func (e *RecursiveExpr) Match(m Position) (Position, Value) {
	if e.e == nil {
		return m.Fail(), nil
	}
	return e.e.Match(m)
}

func (e *RecursiveExpr) Set(val Expr) {
	if !e.set {
		e.e = val
	}
	e.set = true
}

/*
	Processing data
*/

func Bind(e Expr, f func(Value) Value) Expr {
	return Matcher(func (m Position) (Position, Value) {
		n, x := e.Match(m)
		if n.Failed() {
			return n, nil
		}
		return n, f(x)
	})
}

func Fold(e Expr, acc Value, f func(x, acc Value) Value) Expr {
	return Bind(e, func(v Value) Value {
		for _, x := range v.([]Value) {
			acc = f(x, acc)
		}
		return acc
	})
}

func Join(e Expr, sep string) Expr {
	return Fold(e, "", func(x_, acc_ Value) Value {
		x, acc := x_.(string), acc_.(string)
		if acc == "" {
			return x
		}
		return acc + sep + x
	})
}

func Merge(e Expr) Expr {
	return Join(e, "")
}

func Select(e Expr, n int) Expr {
	return Bind(e, func(v Value) Value {
		return v.([]Value)[n]
	})
}

/*
	Giving it a nice interface
*/

func MatcherInput(r io.Reader) Input {
	return bufio.NewReader(r)
}

func ParseNamed(e Expr, fn string, src Input) (Position, Value) {
	return e.Match(start(fn, src))
}

func Parse(e Expr, src Input) (Position, Value) {
	return ParseNamed(e, "", src)
}

func ParseString(e Expr, src string) (Position, Value) {
	return Parse(e, MatcherInput(strings.NewReader(src)))
}

func ParseFile(e Expr, fn string) (Position, Value) {
	_, name := path.Split(fn)
	f, err := os.Open(fn, os.O_RDONLY, 0)
	if err != nil {
		return new(Position).Fail(), nil
	}
	return ParseNamed(e, name, MatcherInput(f))
}

