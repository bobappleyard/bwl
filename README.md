This repository collects some helper libraries for constructing web applications in Go. Well, OK, there's only one at the moment, but I think it's a good one!

apage -- Anonymous Pages
========================

Go combines lexical scope with first-class functions. Put another way, Go supports "closures."

    func Accum(base int) func(int) int {
      return func(off int) int {
        base += off;
        return base;
      }
    }

Go has a library for creating web applications, "http," where functions can be attached to paths on a web server, so as to render web pages. Wouldn't it be nice if we could create a web page that corresponded to an anonymous function?

This is based on some of the ideas encapsulated in the PLT Web Server. Some of the benefits given there extend here as well. The central idea is that the stateful interactions between the server and the client (and, yes, I'm sorry to say, they do exist) operate correctly under actual use. The two main problems for the server are the back button and the use of multiple tabs/windows by the client.

Imagine for the moment that you are considering a train journey. You go to a website offering tickets. This journey might involve multiple stops, each of which could have several different pricing regimes (a not all that unlikely proposition). You choose one of the sets of options, and proceed to the  checkout. Before finalising payment, however, you might decide to try a  different set of choices. So you hit the back button a few times and then try to proceed in a different window. If the website is like most websites out there, this will fail outright, or could even lead you to booking both journeys -- arguably a worse outcome. This is because the state of the interaction between you and the website is handled through mechanisms that amount to global variables that get clobbered by your use of seemingly straightforward browser features.

PLT Web Server accomodates for this situation using continuations. They are a feature that is present in Scheme, but not in Go. How might Go handle such a situation?

Continuations can be expressed in a language that lacks them by using something called "continuation-passing style." That is, a function is called, and in that call another function, representing its continuation, is passed in as an argument. When the called function finishes, it calls the continuation function. A similar process may be expressed in the context of web applications. Instead of passing a function, a link may be embodied in the page sent to the client that corresponds to a function that is called when the link is followed. In effect, this is the continuation of the stateful interaction with the server.

An example, the Arc Challenge:

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

Obviously not as nice as the Arc version, all that string writing stuff is a bit off, but you get the idea.


lexer -- A simple lexical analyser
==================================

The implementation is based on an NFA as described here:

http://swtch.com/~rsc/regexp/regexp1.html

It's slightly less efficient than the aforementioned, as it creates more states than are strictly necessary. This is because I skipped out some of the processing when parsing the regex that the above author took when building the state graph. That part should probably be rewritten anyway -- spaghetti is a generous description of its current status.

I'll license it under GPLv3 or above. Nyah!

API
---

The library exports a struct type, Lexer, that has a range of methods for building an NFA and then matching against a buffer. Unlike most regular expression libraries, this is intended more as a tokeniser for a language parser. The main difference is that instead of simply matching or failing, the Lexer will return an integer corresponding to which token was matched. Further, submatches are not available. This facility can be added in at some point should the need arise. Backreferences are unlikely to make an appearance (they're nasty, anyway :P).

While it is possible to manipulate the state graph manually, using the regex parser provided is probably easier. Bind these expressions to token identifiers (integers), and then call match. If a token is matched, the corresponding identifier is returned. -1 is returned in the case of a failed match.

e.g.

	  const (
		  IDENT = iota;
		  NUMBER;
		  ...
	  )

	  ...

		  l := new(lexer.Lexer);
		  l.AddRegex(`[a-zA-Z_][0-9a-zA-Z_]*`, IDENT);
		  l.AddRegex(`[0-9]+(\.[0-9]+)?`, NUMBER);
		
	  ...
		  l.Start(src);
		  switch l.Next() {
		      case IDENT:
		          return NewIdent(l.String()); // do something with the result
		      case NUMBER:
		      ...
		      
		      case -1:
		        // handle failure
		  }

Supported Language
------------------

The regex language is pretty standard, if basic

a       -- matches `a`
ab      -- matches `a` followed by `b`
a|b     -- matches `a` or `b`
(ab)|c  -- matches `a` followed by `b`, or `c`
a?      -- matches zero or one occurrences of `a`
a*      -- matches zero or more occurrences of `a`
a+      -- matches one or more occurrences of `a`
[a-d]   -- matches `a`, `b`, `c` or `d`
[abcd]  -- same as above
[-ab]   -- matches `-`, `a` or `b`

That's it, so far.


