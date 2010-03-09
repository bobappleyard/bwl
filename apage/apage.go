package apage

import (
	"http"
	"container/list"
	"io"
	"rand"
	"strconv"
	"path"
	"./actor"
)

const CACHE_DEFAULT = 1024

type AnonymousPageServer struct {
	a *actor.Actor
	limit int
	prefix string
	items *list.List
	paths map[int64] http.Handler
}

func New(prefix string) *AnonymousPageServer {
	return &AnonymousPageServer {
		actor.New(),
		CACHE_DEFAULT,
		"/" + prefix + "/",
		list.New(),
		make(map[int64] http.Handler),
	}
}

// Changes the maximum number of items in the handler cache.
func (self *AnonymousPageServer) SetCacheSize(newSize int) {
	self.a.Schedule(func() interface{} {
		self.limit = newSize
		return nil
	})
}

// Attaches a handler to a page, and returns a link to it.
func (self *AnonymousPageServer) Handle(h http.Handler) string {
	return self.a.Schedule(func() interface{} {
		// clear cache
		for self.items.Len() >= self.limit {
			e := self.items.Back()
			self.paths[e.Value.(int64)] = nil, false
			self.items.Remove(e)
		}
		// generate a key
		var key int64
		done := false
		for !done {
			key = rand.Int63()
			if _, ok := self.paths[key]; !ok {
				done = true
			}
		}
		// add new item
		self.items.PushFront(key)
		self.paths[key] = h
		// and return the path
		return self.prefix + strconv.Itoa64(key)
	}).(string)
}

// A shortcut for function pages (which should be most of them)
func (self *AnonymousPageServer) Create(h func(c *http.Conn, r *http.Request)) string {
	return self.Handle(http.HandlerFunc(h))
}

func (self *AnonymousPageServer) getPage(id int64) http.Handler {
	return self.a.Schedule(func() interface{} {
		page, ok := self.paths[id]
		if !ok {
			return http.HandlerFunc(func(c *http.Conn, r *http.Request) {
				c.WriteHeader(http.StatusNotFound)
				io.WriteString(c, "page is not in cache")
			})
		}
		return page
	}).(http.Handler)
}

func (self *AnonymousPageServer) Attach(s *http.ServeMux) {
	s.Handle(self.prefix, http.HandlerFunc(func (c *http.Conn, r *http.Request) {
		_, name := path.Split(r.URL.Path)
		id, err := strconv.Atoi64(name)
		if err != nil {
			c.WriteHeader(http.StatusBadRequest)
			io.WriteString(c, "invalid page id")
		} else {
			self.getPage(id).ServeHTTP(c, r)
		}
	}))
}

var server = func() *AnonymousPageServer {
	res := New("apage")
	res.Attach(http.DefaultServeMux)
	return res
}()

func SetCacheSize(newSize int) {
	server.SetCacheSize(newSize)
}

func Handle(h http.Handler) string {
	return server.Handle(h)
}

func Create(h func(c *http.Conn, r *http.Request)) string {
	return server.Create(h)
}
