/*
	The Arc Challenge. A version in Go.
*/

package main

import (
	"fmt"
	"net/http"

	"github.com/bobappleyard/bwl/apage"
)

func said(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(
		w,
		`<form method="POST" action="%s">
			<input type="text" name="msg"></input>
			<input type="submit"></input>
		</form>`,
		apage.Create(func(x http.ResponseWriter, s *http.Request) {
			fmt.Fprintf(
				x,
				"<a href=\"%s\">click here</a>",
				apage.Create(func(y http.ResponseWriter, t *http.Request) {
					fmt.Fprintf(y, "you said: %s", s.FormValue("msg"))
				}),
			)
		}),
	)
}

func main() {
	http.Handle("/said", http.HandlerFunc(said))
	http.ListenAndServe(":12345", nil)
}
