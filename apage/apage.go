package apage

import (
	http	"http";
	list 	"container/list";
	io		"io";
	hex 	"encoding/hex";
	md5 	"crypto/md5";
	os		"os";
	strings	"strings";
	fmt 	"fmt";
)

// based on messaging
type msg struct {
	t string;
	d interface{};
}

var send, receive chan *msg

// API

// change the number of items that appear in the apage cache
func SetCacheSize(s int) {
	send <- &msg { "set-cache", s };
}

// create a new anonymous page
func Handle(h http.Handler) string {
	send <- &msg { "add-handler", h };
	m := <-receive;
	return m.d.(string);
}

// shortcut for function pages
func Create(h func(c *http.Conn, r *http.Request)) string {
	return Handle(http.HandlerFunc(h));
}

// implementation

type item struct {
	a http.Handler;
	key string;
}

type cache struct {
	items *list.List;
	paths map[string] http.Handler;
}

func initCache() *cache {
	res := new(cache);
	res.items = list.New();
	res.paths = make(map[string] http.Handler);
	return res;
}

const CACHE_DEFAULT = 1024

func serveFail(c *http.Conn, r *http.Request) {
	io.WriteString(c, "fallen out of cache");
}

func init() {
	// messaging architecture
	server := make(map[string] func(data interface{}));
	send = make(chan *msg);
	receive = make(chan *msg);
	go func() {
		for {
			m := <-send;
			f := server[m.t];
			if f != nil {
				f(m.d);
			}			
		}
	}();
	// init the cache
	cache_limit := CACHE_DEFAULT;
	cache := initCache();
	// message handlers
	server["set-cache"] = func(data interface{}) {
		cache_limit = data.(int);
	};
	server["add-handler"] = func(data interface{}) {
		// remove old item from the cache
		if cache.items.Len() == cache_limit {
			e := cache.items.Back();
			cache.paths[e.Value.(*item).key] = nil, false;
			cache.items.Remove(e);
		}
		// add new item to the cache
		cache.items.PushFront(data);
		// generate a key
		h := md5.New();
		t, n, _ := os.Time();
		h.Write(strings.Bytes(fmt.Sprintf("%v %v", t, n)));
		p := hex.EncodeToString(h.Sum());
		cache.paths[p] = data.(http.Handler);
		receive <- &msg { "", "/apage?p=" + p };
	};
	server["serve-page"] = func(data interface{}) {
		page, ok := cache.paths[data.(string)];
		if !ok {
			receive <- &msg { "", http.HandlerFunc(serveFail) };
		}
		receive <- &msg { "", page };
	};
	// connect all this to the http library
	http.Handle("/apage", http.HandlerFunc(func (c *http.Conn, r *http.Request) {
		send <- &msg { "serve-page", r.FormValue("p") };
		m := <-receive;
		h := m.d.(http.Handler);
		h.ServeHTTP(c, r);
	}));
}





