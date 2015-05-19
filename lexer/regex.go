package lexer

import (
	"bytes"
	"container/list"
	"errors"
	"strings"

	bwlerrors "github.com/bobappleyard/bwl/errors"
)

type RegexSet map[rune]string

var defaultMeta = RegexSet{
	'w': "a-zA-Z0-9_",
	's': " \t\n\r",
	'a': "a-zA-Z",
	'd': "0-9",
	'W': "^a-zA-Z0-9_",
	'S': "^ \t\n\r",
	'A': "^a-zA-Z",
	'D': "^0-9",
}

func ExtendSet(base, ext RegexSet) RegexSet {
	res := make(RegexSet)
	if base == nil {
		base = defaultMeta
	}
	for k, v := range base {
		res[k] = v
	}
	for k, v := range ext {
		res[k] = v
	}
	return res
}

func (self *Lexer) Regex(re string, m RegexSet) (*BasicState, error) {
	return self.root.AddRegex(re, m)
}

func (self *Lexer) ForceRegex(re string, m RegexSet) *BasicState {
	res, err := self.Regex(re, m)
	if err != nil {
		println(re, err.Error())
	}
	bwlerrors.Fatal(err)
	return res
}

func (self *Lexer) Regexes(m, regexes RegexSet) {
	for i, x := range regexes {
		self.ForceRegex(x, m).SetFinal(int(i))
	}
}

type Regex struct {
	l *Lexer
}

func NewRegex(re string, m RegexSet) *Regex {
	l := New()
	l.ForceRegex(re, m).SetFinal(0)
	l.ForceRegex(".", nil).SetFinal(1)
	return &Regex{l}
}

func (self *Regex) Match(s string) bool {
	self.l.StartString(s)
	if self.l.Next() == 0 {
		return self.l.Len() == len(s)
	}
	return false
}

func (self *Regex) Matches(s string) []string {
	res := make([]string, 0)
	self.l.StartString(s)
	for !self.l.Eof() {
		if self.l.Next() == 0 {
			res = append(res, self.l.String())
		}
	}
	return res
}

func (self *Regex) Replace(s string, f func(string) string) string {
	res := make([]string, 0)
	buf := bytes.Runes([]byte(s))
	last := 0
	self.l.StartString(s)
	for !self.l.Eof() {
		if self.l.Next() == 0 {
			res = append(res, string(buf[last:self.l.Pos()]))
			res = append(res, f(self.l.String()))
			last = self.l.Pos() + self.l.Len()
		}
	}
	res = append(res, string(buf[last:]))
	return strings.Join(res, "")
}

func Match(re, s string) bool {
	expr := NewRegex(re, nil)
	return expr.Match(s)
}

func Matches(re, s string) []string {
	expr := NewRegex(re, nil)
	return expr.Matches(s)
}

func Replace(re, s string, f func(string) string) string {
	expr := NewRegex(re, nil)
	return expr.Replace(s, f)
}

type regexPos struct {
	start, end *BasicState
}

func (self *BasicState) AddRegex(re string, m RegexSet) (*BasicState, error) {
	if m == nil {
		m = defaultMeta
	}
	// this is just the sort of horror show that lexical analysis avoids

	// stack machine
	start := self
	end := NewState()
	stack := list.New()

	// state flags
	expr, esc, cs := false, false, false
	setstr := ""

	// go into a subexpression
	push := func() {
		rp := &regexPos{start, end}
		stack.PushBack(rp)
		end = NewState()
		start.AddEmptyTransition(end)
	}
	// come out of a subexpression
	pop := func() {
		v := stack.Back()
		stack.Remove(v)
		rp := v.Value.(*regexPos)
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
		// escaped characters
		if esc {
			esc = false
			// inside a charset jobby
			if cs {
				setstr += string(c)
				continue
			}
			// check out the metachar action
			if meta, ok := m[c]; ok {
				move()
				chars, err := Charset(meta, end)
				if err != nil {
					return nil, err
				}
				start.AddEmptyTransition(chars)
				expr = true
				continue
			}
			// nothing else going on? well you escaped it for a reason
			goto add
		}
		// charsets
		if cs {
			if c == '\\' {
				esc = true
			} else if c == ']' {
				chars, err := Charset(setstr, end)
				if err != nil {
					return nil, err
				}
				start.AddEmptyTransition(chars)
				setstr = ""
				cs = false
				expr = true
			} else {
				setstr += string(c)
			}
			continue
		}
		// everything else
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
				return nil, errors.New("trying to close unopened charset")
			}
		// grouping
		case '(':
			move()
			push()
			expr = false
		case ')':
			if stack.Len() <= 1 {
				return nil, errors.New("trying to close unopened subexpr")
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
			return nil, errors.New("nothing to modify")
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

	// some final consistency checks
	if cs {
		return nil, errors.New("unclosed charset")
	}
	if esc {
		return nil, errors.New("invalid escape sequence")
	}
	if stack.Len() > 1 {
		return nil, errors.New("unclosed subexpr")
	}

	// close the implicit brackets
	pop()

	return end, nil
}
