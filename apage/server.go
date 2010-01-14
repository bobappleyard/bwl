package apage

import (
	"fmt";
	"os";
)

type value interface{}

type msg struct {
	f func(value) value;
	data value;
	recv chan value;
}

type Actor struct {
	send chan *msg;
}

func (self *Actor) init() {
	self.send = make(chan *msg);
	go func() {
		for {
			m := <-self.send;
			r := m.f(m.data);
			if m.recv != nil {
				m.recv <- r;
			}
		}
	}();
}

func (self *Actor) CreateHandler(h value) (func(value) value) {
	if self.send == nil {
		self.init();
	}
	switch handler := h.(type) {
		case func(value):
			return func(d value) value {
				self.send <- &msg { 
					func (e value) value { 
						handler(e); 
						return nil; 
					}, 
					d, 
					nil 
				};
				return nil;
			}
		case func(value) value:
			return func(d value) value {
				recv := make(chan value);
				self.send <- &msg { handler, d, recv };
				return <- recv;
			}
		default:
			fmt.Fprintf(os.Stderr, "not a valid function: %v", h);
			os.Exit(-1);
	}
	panic("unreachable");
}

