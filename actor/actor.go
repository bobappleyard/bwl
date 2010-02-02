package actor

import (
	"fmt"
	"os"
)

type msg struct {
	f func(interface{}) interface{}
	data interface{}
	recv chan interface{}
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

func (self *Actor) Add(h interface{}) (func(interface{}) interface{}) {
	switch handler := h.(type) {
		case func(interface{}):
			return func(d interface{}) interface{} {
				self.send <- &msg { 
					func (e interface{}) interface{} { 
						handler(e) 
						return nil 
					}, 
					d, 
					nil,
				}
				return nil
			}
		case func(interface{}) interface{}:
			return func(d interface{}) interface{} {
				recv := make(chan interface{})
				self.send <- &msg { handler, d, recv }
				return <- recv
			}
		default:
			fmt.Fprintf(os.Stderr, "not a valid function: %v", h)
			os.Exit(-1)
	}
	panic("unreachable")
}

func (self *Actor) AddMany(h []interface{}) ([]func(interface{}) interface{}) {
	res := make([]func(interface{}) interface{}, len(h))
	for i, x := range h {
		res[i] = self.Add(x);
	}
	return res;
}
