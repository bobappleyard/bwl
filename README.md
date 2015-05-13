This repository collects some helper libraries for constructing applications in Go.

They're all licensed under GPLv3 or above. Nyah!

Compiling etc
-------------

You can go-get for this.

    $ go get github.com/bobappleyard/bwl/apage

From here, you can check out the example programs.

    $ cd examples
    $ go build arc.go
    $ go build lex.go
    $ go build reg.go
    
You can then go on to try out three demo programs. `lex` creates a lexical analyser where each token type is specified by a regular expression passed in as an argument, and executes this analyser against stdin, printing a report to stdout. `reg` takes a single, regular expression, argument, and then prints all occurences of matches in stdin to stdout. `arc` creates a webserver, listening on port 12345, that has a page, /said, that performs the Arc Challenge. Except it doesn't at the moment, and I don't know why!

actor -- A Simple Serialisation Mechanism
=========================================

The actor package presents a way of making concurrent access to objects simpler. It receives thunks (functions that take no arguments, and can return anything), and then calls them, in sequence, all in a single Goroutine. This easily allows access from many different Goroutines to be serialised. 

Create an actor with `actor.New()`. Then use `Actor.Schedule(thk)` to push a thunk onto the queue. By default, the caller of `Schedule` will wait for the thunk to finish, to get some return value. If this is not desired (and there are many arguments against it), then fire off a Goroutine and forget about it.

This library was more elaborate in a previous life, but I decided to pare it down in order that it play nice with Go's OOP-ish stuff. It's now dashed simple, and also quite lovely. I want to hug it sometimes!

Example
-------

This uses an actor to manage access to a very simple bank database. Both synchronous and asynchronous uses of the Actor API are presented.

```go
type Bank struct {
	a *actor.Actor
	// no overdraft
	accounts map[string] uint
}

func (self *Bank) Deposit(name string, val uint) {
	go self.a.Schedule(func() interface{} {
		balance, ok := self.accounts[name]
		if !ok { balance = 0 }
		balance += val
		self.accounts[name] = balance
		return nil
	})
}

func (self *Bank) Withdraw(name string, val uint) uint {
	return self.a.Schedule(func() interface{} {
		balance, ok := self.accounts[name]
		if !ok { balance = 0 }
		if balance < val { val = balance }
		balance -= val
		self.accounts[name] = balance
		return val
	}).(int)
}
```

apage -- Anonymous Pages
========================

Go combines lexical scope with first-class functions. Put another way, Go supports "closures."

```go
func Accum(base int) func(int) int {
	return func(off int) int {
		base += off;
		return base;
	}
}
```

Go has a library for creating web applications, "http," where functions can be attached to paths on a web server, so as to render web pages. Wouldn't it be nice if we could create a web page that corresponded to an anonymous function?

This is based on some of the ideas encapsulated in the PLT Web Server. Some of the benefits given there extend here as well. The central idea is that the stateful interactions between the server and the client (and, yes, I'm sorry to say, they do exist) operate correctly under actual use. The two main problems for the server are the back button and the use of multiple tabs/windows by the client.

Imagine for the moment that you are considering a train journey. You go to a website offering tickets. This journey might involve multiple stops, each of which could have several different pricing regimes (a not all that unlikely proposition). You choose one of the sets of options, and proceed to the  checkout. Before finalising payment, however, you might decide to try a  different set of choices. So you hit the back button a few times and then try to proceed in a different window. If the website is like most websites out there, this will fail outright, or could even lead you to booking both journeys -- arguably a worse outcome. This is because the state of the interaction between you and the website is handled through mechanisms that amount to global variables that get clobbered by your use of seemingly straightforward browser features.

PLT Web Server accomodates for this situation using continuations. They are a feature that is present in Scheme, but not in Go. How might Go handle such a situation?

Continuations can be expressed in a language that lacks them by using something called "continuation-passing style." That is, a function is called, and in that call another function, representing its continuation, is passed in as an argument. When the called function finishes, it calls the continuation function. A similar process may be expressed in the context of web applications. Instead of passing a function, a link may be embodied in the page sent to the client that corresponds to a function that is called when the link is followed. In effect, this is the continuation of the stateful interaction with the server.

An example, the Arc Challenge:

```go
func Said(c *http.Conn, r *http.Request) {
	io.WriteString(
		c,
		`<form method="post" action="` + 
		apage.Create(func(d *http.Conn, s *http.Request) {
			io.WriteString(
				d, 
				`<a href="` +
				apage.Create(func(e *http.Conn, t *http.Request) {
					io.WriteString(
						e,
						`you said: ` + s.FormValue("foo")
					);
				}) +
				`">click here</a>`
			);
		}) +
		`">` +
		`<input type="text" name="foo"></input>` +
		`<input type="submit">` +
		`</form>`
	);
}
```

Obviously not as nice as the Arc version, all that string writing stuff is a bit off, but you get the idea.

API
---

There are three functions in the API. SetCacheSize() tells the library how many anonymous page handlers to keep cached. Setting this to a large number will lead to increased memory usage, setting it to a small number may cause the library to "forget" page handlers before users might need them. Handle() attaches a http.Handler to the library. Create() does the same, but assumes it's a function and does the type conversion for you.


errors -- Error Handling
========================

This library currently contains a single function, Fatal(), that, when given an os.Error, crashes the program with a stack trace, along with the error's message, if that error is not nil. This makes writing toplevel programs slightly easier, along with functions that absolutely must succeed.

lexer -- A simple lexical analyser
==================================

The implementation is based on an NFA as described here:

http://swtch.com/~rsc/regexp/regexp1.html

It's slightly less efficient than the aforementioned, as it creates more states than are strictly necessary. This is because I skipped out some of the processing when parsing the regex that the above author took when building the state graph. That part should probably be rewritten anyway -- spaghetti is a generous description of its current status.

API
---

The library exports a struct type, Lexer, that has a range of methods for building an NFA and then matching against a buffer. Unlike most regular expression libraries, this is intended more as a tokeniser for a language parser. The main difference is that instead of simply matching or failing, the Lexer will return an integer corresponding to which token was matched. Further, submatches are not available. This facility can be added in at some point should the need arise. Backreferences are unlikely to make an appearance (they're nasty, anyway :P).

While it is possible to manipulate the state graph manually, using the regex parser provided is probably easier. Bind these expressions to token identifiers (integers), and then call match. If a token is matched, the corresponding identifier is returned. -1 is returned in the case of a failed match.

e.g.

```go
const (
	IDENT = iota
	NUMBER
	...
)

...

	l := new(lexer.Lexer);
	l.ForceRegex(`[a-zA-Z_][0-9a-zA-Z_]*`).SetFinal(IDENT)
	l.ForceRegex(`[0-9]+(\.[0-9]+)?`).SetFinal(NUMBER)
	
	...
		l.Start(src)
		switch l.Next() {
			case IDENT:
				return NewIdent(l.String()) // do something with the result
			case NUMBER:
				...
		    
			case -1:
				// handle failure
		}
```

For some more examples of using the interface, see the regex.go file in the library.

Supported Language
------------------

The regex language is pretty standard, if a little basic

a       -- matches `a`

ab      -- matches `a` followed by `b`

a|b     -- matches `a` or `b`

(ab)|c  -- matches `a` followed by `b`, or `c`

a?      -- matches zero or one occurrences of `a`

a\*      -- matches zero or more occurrences of `a`

a+      -- matches one or more occurrences of `a`

[a-d]   -- matches `a`, `b`, `c` or `d`

[abcd]  -- same as above

[-ab]   -- matches `-`, `a` or `b`

[^a]    -- matches anything but `a`

\a      -- *escapes* `a`: if `a` is a metacharacter (see below) then it matches the appropriate expression. Otherwise, it matches `a`. This is useful for matching on characters with special meaning, e.g. `\?` matches `?` where ordinarily an error would be thrown.

Metacharacters
--------------

Metacharacters are characters that stand for regular expressions. When the regular expression 

peg -- A Parser Library
=======================

PEG parsers are all the rage these days. Go makes writing them particularly easy. There are, principally, two concepts introduced in this library. The first is the Position interface:

```go
type Position interface {
	Next() Position
	Fail() Position
	Pos() int
	Failed() bool
	Eof() bool
	Id() int
	Data() interface{}
}
```

This describes a position in the source. The two main methods are listed first: Next() moves through the source, Fail() aborts the parse. Id() is also important as it identifies the kind of data at that position (e.g. the character code or the token id).

The second is the Expr interface:

```go
type Expr interface {
	Match(m Position) (Position, interface{})
}
```

This describes a PEG expression. If it succeeds in matching at a particular position, it should return (that position).Next(). If it does not it should call Fail().

Expressions
-----------

The bulk of the library consists of different kinds of expressions.

Matcher -- a function that looks like an Expr. Useful for simple expressions that don't need to be represented as structs or slices or anything like that.

Any         -- always matches.

None        -- never matches.

Eof         -- matches against the end of the input.

Terminal    -- an integer type that matches against the Id() method of the Position passed in, and returns the Position's Data() if it matches. Terminals are an important building block to PEGs. For some more information on handling them, see below.

And         -- a slice type representing a series of expressions that must all match, in order, for the whole expression to match. The data part is a slice containing the data parts of each of the subexpressions.

Or          -- a slice type representing a choice of expressions. The first subexpression to match will be what the Or matches to at the given position.

Extensible  -- returns an ExtensibleExpr that behaves like Or when matching, but can have extra alternatives incorporated through calls to Add(). Note that "left recursion" should be guarded against. If you add an Extensible to itself, an infinite loop will occur during matching. This is also true if you add anything where the first item is the extensible.

Quantify    -- returns an expression object that repeats the matching of another expression a given number of times. Given are a maximum and minumum number of repetitions. The maximum being -1 corresponds to there being no maximum.

Option      -- is equivalent to Quantify(*expr*, 0, 1).

Repeat      -- is equivalent to Quantify(*expr*, 0, -1).

Multi       -- is equivalent to Quantify(*expr*, 1, -1).

Ensure      -- returns an expression that performs lookahead. That is, it matches on an expression passed in, but instead of returning the Position *after* the match, returns the Position before (that is, it returns the Position passed into Match()).

Prevent     -- returns an expression that behaves like Ensure, but only matches when the expression passed in does not match.

RepeatUntil -- returns an expression that continues matching one expression, until another expression matches, before returning. If the first expression does not match at any point, the entire expression will fail.

Bind        -- returns an expression object whereby the data returned on Match() is subject to processing.

Fold        -- returns an expression object that applies a function over every item in the data returned from a match, given that the data is a slice of interface{}s. Expression types that do this in the PEG library include And, Quantify, Option, Repeat, Multi and RepeatUntil. This result is accumulated into a single value.

Map         -- returns an expression object that applies a function over every item in the data returned from a match, given that the data is a slice of interface{}s. This result is accumulated into a slice of the same length.

Join        -- returns an expression object that assumes the data it is processing is a slice of strings. It collects those strings together into one string, with a separator between them.

Merge       -- equivalent to Join(*expr*, "")

Select      -- returns an expression object that assumes the data it is processing is a slice of interface{}s. Returns a given element of said slice upon matching.

Using Terminals
---------------

Terminals are the main primitive in the PEG library. The PEG is effectively processing a series of integers. We can return to the lexer example, and by changing the const declaration so that it reads:

```go
const (
	IDENT peg.Terminal = iota
	NUMBER
	...
)
```

Will allow tokens parsed by the lexer to be used by the PEG.

One could also, assuming the character stream input, create a matcher for a string by doing something like:

```go
func String(s string) peg.Expr {
	rs := strings.Runes(s)
	res := make(peg.And, len(rs))
	for i, x := range rs {
		res[i] = peg.Terminal(x)
	}
	return res
}
```

This is not included in the library because using the lexer for this sort of thing is preferred. If anyone cares, it could be put in, but unlike all the other expression types listed above, it matters what sort of input the PEG is processing.
