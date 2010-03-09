/*
	The Arc Challenge. A version in Go.
*/

package main

import (
	"http"
	"fmt"
	"./apage"
)

func said(c *http.Conn, r *http.Request) {
	fmt.Fprintf(
		c,
		`<form method="POST" action="%s">
			<input type="text" name="msg"></input>
			<input type="submit"></input>
		</form>`,
		apage.Create(func(d *http.Conn, s *http.Request) {
			fmt.Fprintf(
				d, 
				"<a href=\"%s\">click here</a>",	
				apage.Create(func(e *http.Conn, t *http.Request) {
					fmt.Fprintf(e, "you said: %s", s.FormValue("msg"))
				}),
			)
		}),
	)
}

func main() {
	http.Handle("/said", http.HandlerFunc(said))
	http.ListenAndServe(":12345", nil)
}

