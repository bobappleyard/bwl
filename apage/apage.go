package apage

import (
	"http";
	"container/list";
	"io";
	"rand";
	"strconv";
)

// API
var setCache, addHandler func(value) value 

// Changes the maximum number of items in the handler cache.
func SetCacheSize(newSize int) {
	setCache(newSize);
}

// Attaches a handler to a page, and returns a link to it.
func Handle(h http.Handler) string {
	return addHandler(h).(string);
}

// A shortcut for function pages.
func Create(h func(c *http.Conn, r *http.Request)) string {
	return Handle(http.HandlerFunc(h));
}

// implementation

const CACHE_DEFAULT = 1024

func init() {
	// init the cache
	cache_limit := CACHE_DEFAULT;
	items := list.New();
	paths := make(map[int64] http.Handler);
	// start the server
	a := new(Actor);
	// register some message handlers
	setCache = a.CreateHandler(func (data value) {
		cache_limit = data.(int);
	});
	addHandler = a.CreateHandler(func (data value) value {
		// remove old items from the cache
		for items.Len() >= cache_limit {
			e := items.Back();
			paths[e.Value.(int64)] = nil, false;
			items.Remove(e);
		}
		// generate a key
		var key int64;
		done := false;
		for !done {
			key = rand.Int63();
			if _, ok := paths[key]; !ok {
				done = true;
			}
		}
		// add new item
		items.PushFront(key);
		paths[key] = data.(http.Handler);
		// and return the path
		return "/apage/" + strconv.Itoa64(key);
	});
	serveFail := func(c *http.Conn, r *http.Request) {
		io.WriteString(c, "error while trying to serve page");
	};
	servePage := a.CreateHandler(func (data value) value {
		page, ok := paths[data.(int64)];
		if !ok {
			return http.HandlerFunc(serveFail);
		}
		return page;
	});
	
	// connect all this to the http library
	http.Handle("/apage/", http.HandlerFunc(func (c *http.Conn, r *http.Request) {
		_, name := path.Split(r.URL.Path);
		id, err := strconv.Atoi64(name);
		if err != nil {
			serveFail(c, r);
		} else {
			h := servePage(id).(http.Handler);
			h.ServeHTTP(c, r);
		}
	}));
}





