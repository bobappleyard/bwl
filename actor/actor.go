package actor

import (
	"fmt"
	"os"
)

type Value interface{}

type msg struct {
	f func(Value) Value
	data Value
	recv chan Value
}

// An actor is a collection of functions that all run in the same goroutine,
// but can be called to do work from any goroutine. The requests to the actor
// are serialised. 
type Actor struct {
	send chan *msg
}

func New() *Actor {
	res := &Actor { make(chan *msg, 20) }
	go func() {
		for {
			m := <-res.send
			r := m.f(m.data)
			if m.recv != nil {
				m.recv <- r
			}
		}
	}()
	return res
}

func (self *Actor) Add(h Value) (func(Value) Value) {
	switch handler := h.(type) {
		case func(Value):
			return func(d Value) Value {
				self.send <- &msg { 
					func (e Value) Value { 
						handler(e) 
						return nil 
					}, 
					d, 
					nil,
				}
				return nil
			}
		case func(Value) Value:
			return func(d Value) Value {
				recv := make(chan Value)
				self.send <- &msg { handler, d, recv }
				return <- recv
			}
		default:
			fmt.Fprintf(os.Stderr, "not a valid function: %v", h)
			os.Exit(-1)
	}
	panic("unreachable")
}

func (self *Actor) AddMany(h []Value) ([]func(Value) Value) {
	res := make([]func(Value) Value, len(h))
	for i, x := range h {
		res[i] = self.Add(x);
	}
	return res;
}
