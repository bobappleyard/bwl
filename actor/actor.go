package actor

type msg struct {
	thk  func() interface{}
	recv chan interface{}
}

// An actor is a collection of functions that all run in the same goroutine,
// but can be called to do work from any goroutine. The requests to the actor
// are serialised.
type Actor struct {
	send chan *msg
}

func New() *Actor {
	res := &Actor{make(chan *msg, 20)}
	go func() {
		for m := range res.send {
			m.recv <- m.thk()
			close(m.recv)
		}
	}()
	return res
}

func (self *Actor) Schedule(thk func() interface{}) interface{} {
	recv := make(chan interface{})
	self.send <- &msg{thk, recv}
	return <-recv
}
