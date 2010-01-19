package lexer

import (
	"os"
	"container/vector"
	"strings"
	"./errors"
)

type Metachars map[int] string

var defaultMeta = Metachars {
	'w': "(\\a|\\d|_)",
	's': "[ \t\n\r]",
	'a': "[a-zA-Z]",
	'd': "[0-9]",
	'W': "[^a-zA-Z0-9_]",
	'S': "[^ \t\n\r]",
	'A': "[^a-zA-Z]",
	'D': "[^0-9]",
}

func (self *Lexer) Regex(re string, m Metachars) (*BasicState, os.Error) {
	return AddRegex(self.root, re, m)
}

type Regex struct {
	l, lm *Lexer
}

func NewRegex(re string, m Metachars) *Regex {
	l, lm := New(), New()
	s, err := l.Regex(re, m)
	errors.Fatal(err)
	t, _ := lm.Regex(re, m)
	u, _ := lm.Regex(".", nil)
	s.SetFinal(0)
	t.SetFinal(0)
	u.SetFinal(1)
	return &Regex { l, lm }
}

func (self *Regex) Match(s string) bool {
	src := strings.Runes(s)
	l := self.l
	l.Start(src)
	if l.Next() == 0 {
		return self.l.Len() == len(src)
	}
	return false
}

func (self *Regex) Matches(s string) []string {
	res := new(vector.StringVector)
	l := self.lm
	l.StartString(s)
	for !l.Eof() {
		if l.Next() == 0 {
			res.Push(l.String())
		}
	}
	return res.Data()
}

func Match(re, s string) bool {
	expr := NewRegex(re, nil)
	return expr.Match(s)
}

func Matches(re, s string) []string {
	expr := NewRegex(re, nil)
	return expr.Matches(s)
}

type regexPos struct {
	start, end *BasicState
}

func AddRegex(start *BasicState, re string, m Metachars) (*BasicState, os.Error) {
	if m == nil {
		m = defaultMeta
	}
	// this is just the sort of horror show that lexical analysis avoids

	// stack machine
	end := NewState()
	stack := new(vector.Vector)
	
	// is it ok for a modifier to appear?
	expr := false
	// has the escape character just appeared?
	esc := false
	// charset stuff
	setstr := ""
	cs := false

	// go into a subexpression
	push := func() {
		rp := &regexPos { start, end }
		stack.Push(rp)
		end = NewState()
		start.AddEmptyTransition(end)
	}
	// come out of a subexpression
	pop := func() {
		rp := stack.Pop().(*regexPos)
		end.AddEmptyTransition(rp.end)
		start = rp.start
		end = rp.end
	}
	// move forward, for the purposes of concatenation
	move := func() {
		start = end
		end = NewState()
	}
	
	// the expression is inside an implicit ( ... )
	push()
	
	// parse the expression
	for _, c := range re {
		if esc {
			esc = false
			// check out the metachar action
			if meta, ok := m[c]; ok {
				move()
				s, err := AddRegex(start, meta, m)
				if err != nil {
					return nil, err
				}
				s.AddEmptyTransition(end)
				expr = true
				continue
			}
			// nothing going on? well you escaped it for a reason
			goto add
		}
		if cs && c != ']' {
			setstr += string(c)
			continue
		}
		switch c {
			// charsets
			case '.':
				move()
				start.AddEmptyTransition(Any(end))
				expr = true
			case '[':
				move()
				cs = true
			case ']':
				if !cs {
					return nil, os.ErrorString("trying to close unopened charset")
				}
				chars, err := Charset(setstr, end)
				if err != nil {
					return nil, err
				}
				start.AddEmptyTransition(chars)
				cs = false
				expr = true
			// grouping
			case '(':
				move()
				push()
				expr = false
			case ')':
				if stack.Len() <= 1 {
					return nil, os.ErrorString("trying to close unopened subexpr")
				}
				pop()
				expr = true
			// alternation
			case '|':
				pop()
				push()
				expr = false
			// modifiers
			case '?':
				start.AddEmptyTransition(end)
				goto check
			case '*':
				start.AddEmptyTransition(end)
				end.AddEmptyTransition(start)
				goto check
			case '+':
				end.AddEmptyTransition(start)
				goto check
			// escape character
			case '\\':
				esc = true
				expr = false
			// otherwise just add that char
			default:
				goto add
		}
		continue
		// make sure the modifier modified something
	check:
		if !expr { 
			return nil, os.ErrorString("nothing to modify") 
		}
		expr = false
		continue
		// add a character transition
	add:
		move()
		start.AddTransition(c, end)
		expr = true
		continue
	}
	
	if cs {
		return nil, os.ErrorString("unclosed charset")
	}
	
	if esc {
		return nil, os.ErrorString("invalid escape sequence")
	}
	
	if stack.Len() > 1 {
		return nil, os.ErrorString("unclosed subexpr")
	}
	
	// close the implicit brackets
	pop()
	
	return end, nil
}


