package lexer

import (
	"errors"
	"strings"
)

/* States */

// Encapsulates the actions of the Lexer in terms of states and transitions.
// Move chooses a transition to follow after consuming some input.
// Close follows all transitions that require no input.
// Final returns -1 if it is not a final state. If it is, Final returns the
// identifier associated with the state.
type State interface {
	Move(c rune) []State
	Close() []State
	Final() int
}

/* Yer basic transitiony state */

type BasicState struct {
	transitions map[rune]State
	empty       []State
	final       int
}

func NewState() *BasicState {
	return &BasicState{
		make(map[rune]State),
		make([]State, 0),
		-1,
	}
}

func (self *BasicState) Move(c rune) []State {
	s, ok := self.transitions[c]
	if ok {
		return []State{s}
	}
	return []State{}
}

func (self *BasicState) Close() []State {
	res := make([]State, len(self.empty))
	copy(res, self.empty)
	return res
}

func (self *BasicState) Final() int {
	return self.final
}

func (self *BasicState) AddTransition(c rune, s State) {
	self.transitions[c] = s
}

func (self *BasicState) AddEmptyTransition(s State) {
	self.empty = append(self.empty, s)
}

func (self *BasicState) SetFinal(f int) {
	self.final = f
}

/* More specialised stuff */

type SpecialState struct {
	next []State
}

func (self *SpecialState) SetNext(next State) {
	self.next = []State{next}
}

func (self *SpecialState) Move(c rune) []State {
	return self.next
}

func (self *SpecialState) Close() []State {
	return []State{}
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
	inv   bool
}

func Charset(spec string, next State) (State, error) {
	start := rune(0)
	inrange, inv := false, false
	chars := ""
	res := new(csState)
	res.SetNext(next)
	if spec == "" {
		return res, nil
	} // this will allow nothing past
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
				return nil, errors.New("invalid range specification")
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

func (self *csState) Move(c rune) []State {
	found := strings.Index(self.chars, string(c)) != -1
	if self.inv {
		found = !found
	}
	if found {
		return self.SpecialState.next
	}
	return []State{}
}
