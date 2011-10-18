package peg

/*
	PEG matching
*/

import (
	"container/vector"
)

/*
	An expression interface
	
	Expresssions match a position in the source, or not. The Match method
	should return the position in the input just after the end of the region
	matched, along with the data corresponding to the region matched. This 
	could be a string of the text region itself, or it could be an object
	representing the appropriate part of a syntax tree.
	
	A Matcher is a function that looks like an expression object. This 
	simplifies the creation of new peg forms where being able to treat the 
	expression as a slice or struct or what have you isn't desired.
*/
type Expr interface {
	Match(m Position) (Position, interface{})
}

type Matcher func(m Position) (Position, interface{})

func (self Matcher) Match(m Position) (Position, interface{}) {
	return self(m)
}

/*
	Primitive Expressions
*/

var Any = Matcher(func(m Position) (Position, interface{}) {
	return m.Next(), m.Data()
})

var None = Matcher(func(m Position) (Position, interface{}) {
	return m.Fail(), nil
})

var Eof = Matcher(func(m Position) (Position, interface{}) {
	if m.Eof() {
		return m, nil
	}
	return m.Fail(), nil
})

type Terminal int

func (self Terminal) Match(m Position) (Position, interface{}) {
	if m.Id() == int(self) {
		return m.Next(), m.Data()
	}
	return m.Fail(), nil
}

func QualifiedTerminal(t Terminal, s string) Expr {
	return Matcher(func (m Position) (Position, interface{}) {
		p, d := t.Match(m)
		if !p.Failed() && m.Data().(string) == s {
			return p, d
		}
		return m.Fail(), nil
	})
}

/*
	Combining Expressions
*/

type And []Expr

func (self And) Match(m Position) (Position, interface{}) {
	res := make([]interface{}, len(self))
	for i, x := range self {
		if m.Failed() {
			return m, nil
		}
		m, res[i] = x.Match(m)
	}
	return m, res
}

type Or []Expr

func (self Or) Match(m Position) (Position, interface{}) {
	for _, e := range self {
		n, res := e.(Expr).Match(m)
		if !n.Failed() {
			return n, res
		}
	}
	return m.Fail(), nil
}

type ExtensibleExpr struct {
	es *vector.Vector
	e Or
}

func Extensible () *ExtensibleExpr {
	return &ExtensibleExpr { new(vector.Vector), Or {} }
}

func (self *ExtensibleExpr) Add(e Expr) {
	self.es.Push(e)
}

func (self *ExtensibleExpr) Match(m Position) (Position, interface{}) {
	if len(self.e) != self.es.Len() {
		newe := make(Or, self.es.Len())
		for i, e := range *self.es.Slice(0,self.es.Len()) {
			newe[i] = e.(Expr)
		}
		self.e = newe
	}
	return self.e.Match(m)
}

/*
	Expression quantifiers
*/

type quantifiedExpr struct {
	e Expr
	min, max int
}

func (self *quantifiedExpr) Match(m Position) (Position, interface{}) {
	var item interface{}
	cur := m
	res := new(vector.Vector)
	// guaranteed minimum
	for i := 0; i < self.min; i++ {
		cur, item = self.e.Match(cur)
		if cur.Failed() {
			return cur, nil
		}
		res.Push(item)
	}
	last := cur
	// optional (up to a maximum)
	for i := self.min; self.max == -1 || i < self.max; i++ {
		cur, item = self.e.Match(last)
		if cur.Failed() {
			return last, []interface{}(*res.Slice(0,res.Len()))
		}
		res.Push(item)
		last = cur
	}
	return cur, []interface{}(*res.Slice(0,res.Len()))
}

func Quantify(e Expr, min, max int) Expr {
	return &quantifiedExpr { e, min, max }
}

func Option(e Expr) Expr {
	return Quantify(e, 0, 1)
}

func Repeat(e Expr) Expr {
	return Quantify(e, 0, -1)
}

func Multi(e Expr) Expr {
	return Quantify(e, 1, -1)
}

/*
	Lookahead
*/

func Ensure(e Expr) Expr {
	return Matcher(func (m Position) (Position, interface{}) {
		n, _ := e.Match(m)
		if n.Failed() {
			return n, nil
		}
		return m, nil
	})
}

func Prevent(e Expr) Expr {
	return Matcher(func (m Position) (Position, interface{}) {
		n, _ := e.Match(m)
		if n.Failed() {
			return m, nil
		}
		return m.Fail(), nil
	})
}

func RepeatUntil(e, end Expr) Expr {
	return Select(And { 
		Repeat(Select(And { Prevent(end), e }, 1)), 
		end,
	}, 0)
}

/*
	Processing data
*/

func Bind(e Expr, f func(interface{}) interface{}) Expr {
	return Matcher(func (m Position) (Position, interface{}) {
		n, x := e.Match(m)
		if n.Failed() {
			return n, nil
		}
		return n, f(x)
	})
}

func Fold(e Expr, acc interface{}, f func(x, acc interface{}) interface{}) Expr {
	return Bind(e, func(v interface{}) interface{} {
		for _, x := range v.([]interface{}) {
			acc = f(x, acc)
		}
		return acc
	})
}

func Map(e Expr, f func(interface{}) interface{}) Expr {
	return Bind(e, func(v interface{}) interface{} {
		xs := v.([]interface{})
		res := make([]interface{}, len(xs))
		for i, x := range xs {
			res[i] = f(x)
		}
		return res
	})
}

func Join(e Expr, sep string) Expr {
	return Fold(e, "", func(x_, acc_ interface{}) interface{} {
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
	return Bind(e, func(v interface{}) interface{} {
		return v.([]interface{})[n]
	})
}

