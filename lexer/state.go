package lexer

import (
	"container/vector"
	"os"
	"strings"
)

/* States */

// Encapsulates the actions of the Lexer in terms of states and transitions.
// Move chooses a transition to follow after consuming some input.
// Close follows all transitions that require no input.
// Final returns -1 if it is not a final state. If it is, Final returns the 
// identifier associated with the state.
type State interface {
	Move(c int) []State
	Close() []State
	Final() int
}

/* Yer basic transitiony state */

type BasicState struct {
	transitions map[int] State
	empty *vector.Vector
	final int
}

func NewState() *BasicState {
	return &BasicState {
		make(map[int] State),
		new(vector.Vector),
		-1,
	}
}

func (self *BasicState) Move(c int) []State {
	s, ok := self.transitions[c]
	if ok {
		return []State { s }
	}
	return []State {}
}

func (self *BasicState) Close() []State {
	res := make([]State, self.empty.Len())
	for i, x := range *self.empty.Slice(0,self.empty.Len()) {
		res[i] = x.(State)
	}
	return res
}

func (self *BasicState) Final() int {
	return self.final
}

func (self *BasicState) AddTransition(c int, s State) {
	self.transitions[c] = s
}

func (self *BasicState) AddEmptyTransition(s State) {
	self.empty.Push(s)
}

func (self *BasicState) SetFinal(f int) {
	self.final = f
}

/* More specialised stuff */

type SpecialState struct {
	next []State
}

func (self *SpecialState) SetNext(next State) {
	self.next = []State { next }
}

func (self *SpecialState) Move(c int) []State {
	return self.next
}

func (self *SpecialState) Close() []State {
	return []State {}
}

func (self *SpecialState) Final() int {
	return -1
}

/* Anything */

func Any(next State) State {
	res := new(SpecialState)
	res.SetNext(next)
	return res
}

/* A state with charset stuff */

type csState struct {
	SpecialState
	chars string
	inv bool
}

func Charset(spec string, next State) (State, os.Error) {
	start := 0
	inrange, inv := false, false
	chars := ""
	res := new(csState)
	res.SetNext(next)
	if spec == "" { return res, nil } // this will allow nothing past
	if spec[0] == '^' {
		inv = true
		spec = spec[1:]
	}
	for _, x := range spec {
		switch {
			case x == '-':
				inrange = true
			case inrange:
				if start == 0 || x <= start {
					return nil, os.NewError("invalid range specification")
				}
				for i := start + 1; i <= x; i++ {
					chars += string(i)
				}
				inrange = false
			default:
				chars += string(x)
				start = x
		}
	}
	res.chars = chars
	res.inv = inv
	return res, nil
}

func (self *csState) Move(c int) []State {
	found := strings.Index(self.chars, string(c)) != -1
	if self.inv {
		found = !found
	}
	if found {
		return self.SpecialState.next
	}
	return []State {}
}


