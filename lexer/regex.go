package lexer

import (
	"os"
	"fmt"
	v "container/vector"
)

type regexPos struct {
	start, end *BasicState
}

func (self *Lexer) AddRegex(re string, f int) os.Error {
	// this is just the sort of horror show that lexical analysis avoids

	// stack machine
	start := self.root
	end := NewState()
	stack := new(v.Vector)
	
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
					return os.ErrorString("trying to close unopened charset")
				}
				chars, err := Charset(setstr, end)
				if err != nil {
					return err
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
					return os.ErrorString("trying to close unopened subexpr")
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
			return os.ErrorString("nothing to modify") 
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
		return os.ErrorString("unclosed charset")
	}
	
	if esc {
		return os.ErrorString("invalid escape sequence")
	}
	
	if stack.Len() > 1 {
		return os.ErrorString("unclosed subexpr")
	}
	
	// close the implicit brackets
	pop()
	
	// mark the regex end point
	end.SetFinal(f)
	
	return nil
}

func NewRegex(re string) *Lexer {
	res := New()
	err := res.AddRegex(re, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	return res
}



