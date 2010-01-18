package peg

/*
	PEG matching
*/

import (
	"bufio";
	"fmt";
	"io";
	"os";
	"path";
	v "container/vector";
	"strings";
)

/*
	The position structure stores information about the source text, by 
	moving through it.
*/

const (
	pos_running = iota;
	pos_end;
	pos_fail;
)

type Input interface {
	ReadRune() (int, int, os.Error);
}

type Position struct {
	fn string;
	src Input;
	char, state int;
	pos [2]int;
	next *Position;
}

func start(filename string, r Input) *Position {
	pre := new(Position);
	pre.fn = filename;
	pre.src = r;
	pre.pos[0] = -1;
	return pre.Next();
}

func (s *Position) String() string {
	var name string;
	if s.fn == "" {
		name = "unknown file";
	} else {
		name = s.fn;
	}
	return fmt.Sprintf("%s: [%d, %d]", name, s.pos[0], s.pos[1]);
}

func (s *Position) copy() *Position {
	return &Position { s.fn, s.src, s.char, s.state, s.pos, s.next }
}

func (s *Position) Next() (res *Position) {
	if s.next != nil {
		return s.next;
	}
	switch s.state {
		case pos_running:
			res = s.copy();
			rune, _, err := s.src.ReadRune();
			if err == os.EOF {
				res.char = 0;
				res.state = pos_end;
				// go past the end, to keep in with normal behaviour
				res.pos[0]++;
				return;
			}
			if err != nil {
				return s.Fail();
			}
			res.char = rune;
			if res.char == 10 {
				res.pos[0] = 0;
				res.pos[1]++;
			} else {
				res.pos[0]++;
			}
			s.next = res;
		case pos_end:
			res = s.Fail();
		case pos_fail:
			res = s;
	}
	return;
}

func (s *Position) Fail() (res *Position) {
	switch s.state {
		case pos_running, pos_end:
			res = s.copy();
			res.state = pos_fail;
		case pos_fail:
			res = s;
	}
	return;
}

func (s *Position) Failed() bool {
	return s.state == pos_fail;
}

func (s *Position) EOF() bool {
	return s.state == pos_end;
}

func (s *Position) Char() int {
	return s.char;
}

func (s *Position) Pos() [2]int {
	return s.pos;
}

/*
	The value a part of the source text is judged to have.
*/

type Data interface {}

/*
	An expression interface
	
	Expresssions match a position in the source, or not. The Match method
	should return the position in the code just after the end of the region
	matched.
	
	A Matcher is a function that can be Wrap'd to look like an expression
	object. This simplifies the creation of new peg forms where being able 
	to treat the expression as a slice or struct or what have you isn't
	necessary.
*/

type Expr interface {
	Match(m *Position) (*Position, Data);
}

type Matcher func(m *Position) (*Position, Data);

type closureExpr struct {
	f Matcher;
}

func (e *closureExpr) Match(m *Position) (*Position, Data) {
	return e.f(m);
}

func Wrap(m Matcher) Expr {
	return &closureExpr { m }
}

/*
	Primitive Expressions
*/

var Any = Wrap(func(m *Position) (*Position, Data) {
	return m.Next(), string(m.char);
})

var None = Wrap(func(m *Position) (*Position, Data) {
	return m.Fail(), nil;
})

var EOF = Wrap(func(m *Position) (*Position, Data) {
	if m.EOF() {
		return m, nil;
	}
	return m.Fail(), nil;
})

func Char(c int) Expr {
	return Wrap(func(m *Position) (*Position, Data) {
		if m.char == c {
			return m.Next(), string(c);
		}
		return m.Fail(), nil;
	});
}


func String(s string) Expr {
	return Wrap(func (m *Position) (*Position, Data) {
		cur := m;
		for _, c := range s {
			if cur.char == c {
				cur = cur.Next();
			} else {
				return m.Fail(), nil;
			}
		}
		return cur, s;
	});
}

/*
	Combining Expressions
*/

type And []Expr

func (e And) Match(m *Position) (*Position, Data) {
	cur := m;
	res := make([]Data, len(e));
	for i := 0; i < len(e); i++ {
		if cur.Failed() {
			return cur, nil;
		}
		cur, res[i] = e[i].Match(cur);
	}
	return cur, res;
}

type Or []Expr

func (e Or) Match(m *Position) (cur *Position, res Data) {
	for i := 0; i < len(e); i++ {
		cur, res = e[i].Match(m);
		if !cur.Failed() {
			return;
		}
	}
	return;
}

type Extensible struct {
	es *v.Vector;
}

func NewExtensible () *Extensible {
	return &Extensible { new(v.Vector) };
}

func (self *Extensible) Add(e Expr) {
	self.es.Push(e);
}

func (self *Extensible) Match(m *Position) (*Position, Data) {
	for e := range self.es.Iter() {
		n, res := e.(Expr).Match(m);
		if !n.Failed() {
			return n, res;
		}
	}
	return m.Fail(), nil;
}

/*
	Expression modifiers
*/

type modifiedExpr struct {
	e Expr;
	min, max int;
}

func (e *modifiedExpr) Match(m *Position) (*Position, Data) {
	var item Data;
	cur := m;
	res := new(v.Vector);
	// guaranteed minimum
	for i := 0; i < e.min; i++ {
		cur, item = e.e.Match(cur);
		if cur.Failed() {
			return cur, nil;
		}
		res.Push(item);
	}
	last := cur;
	// optional (up to a maximum)
	for i := e.min; e.max == -1 || i < e.max; i++ {
		cur, item = e.e.Match(last);
		if cur.Failed() {
			return last, res;
		}
		res.Push(item);
		last = cur;
	}
	return cur, res.Data();
}

func Modify(e Expr, min, max int) Expr {
	return &modifiedExpr { e, min, max }
}

func Option(e Expr) Expr {
	return Modify(e, 0, 1);
}

func Repeat(e Expr) Expr {
	return Modify(e, 0, -1);
}

func Multi(e Expr) Expr {
	return Modify(e, 1, -1);
}

/*
	Lookahead
*/

func Ensure(e Expr) Expr {
	return Wrap(func (m *Position) (*Position, Data) {
		n, _ := e.Match(m);
		if n.Failed() {
			return n, nil;
		}
		return m, nil;
	});
}

func Prevent(e Expr) Expr {
	return Wrap(func (m *Position) (*Position, Data) {
		n, _ := e.Match(m);
		if n.Failed() {
			return m, nil;
		}
		return m.Fail(), nil;
	});
}

func RepeatUntil(e, end Expr) Expr {
	return Repeat(Select(And { Prevent(end), e }, 1));
}

/*
	Recursion support
	
	Note that one must guard against left recursion.
*/

type RecursiveExpr struct {
	e Expr;
	set bool;
}

func Recursive() *RecursiveExpr {
	return new(RecursiveExpr);
}

func (e *RecursiveExpr) Match(m *Position) (*Position, Data) {
	if e.e == nil {
		return m.Fail(), nil;
	}
	return e.e.Match(m);
}

func (e *RecursiveExpr) Set(val Expr) {
	if !e.set {
		e.e = val;
	}
	e.set = true;
}

/*
	Processing Data
*/

func Bind(e Expr, f func(Data) Data) Expr {
	return Wrap(func (m *Position) (*Position, Data) {
		n, x := e.Match(m);
		if n.Failed() {
			return n, nil;
		}
		return n, f(x);
	});
}

func Fold(e Expr, acc Data, f func(x, acc Data) Data) Expr {
	return Bind(e, func(v Data) Data {
		for _, x := range v.([]Data) {
			acc = f(x, acc);
		}
		return acc;
	});
}

func Join(e Expr, sep string) Expr {
	return Fold(e, "", func(x_, acc_ Data) Data {
		x := x_.(string);
		acc := acc_.(string);
		if acc == "" {
			return x;
		}
		return acc + sep + x;
	});
}

func Merge(e Expr) Expr {
	return Join(e, "");
}

func Select(e Expr, n int) Expr {
	return Bind(e, func(v Data) Data {
		return v.([]Data)[n];
	});
}

/*
	Utility expressions
*/

var LineEnd = Char(10)
var Whitespace = Or { Char(8), Char(9), Char(10), Char(13), Char(32) }
var Digit = Or { 
	Char('0'), Char('1'), Char('2'), Char('3'), Char('4'), 
	Char('5'), Char('6'), Char('7'), Char('8'), Char('9'), 
}

/*
	Giving it a nice interface
*/

func wrapInput(r io.Reader) Input {
	return bufio.NewReader(r);
}

func ParseNamed(e Expr, fn string, src Input) (*Position, Data) {
	return e.Match(start(fn, src));
}

func Parse(e Expr, src Input) (*Position, Data) {
	return ParseNamed(e, "", src);
}

func ParseString(e Expr, src string) (*Position, Data) {
	return Parse(e, wrapInput(strings.NewReader(src)));
}

func ParseFile(e Expr, fn string) (*Position, Data) {
	_, name := path.Split(fn);
	f, err := os.Open(fn, os.O_RDONLY, 0);
	if err != nil {
		return new(Position).Fail(), nil;
	}
	return ParseNamed(e, name, wrapInput(f));
}

