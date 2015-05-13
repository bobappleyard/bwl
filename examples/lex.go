/*
	Demonstrating use of the lexer.

	Takes a list of regular expressions from the command line, and then
	processes stdin until eof or failure to match to any of the expressions.

	For every match, some information is printed out: the index of the
	expression in the list passed in, the position in the input for the
	match, and the text of the match
*/

package main

import (
	"errors"
	"fmt"
	"os"

	bwlerrors "github.com/bobappleyard/bwl/errors"
	"github.com/bobappleyard/bwl/lexer"
)

func main() {
	l := lexer.New()
	for i, x := range os.Args[1:] {
		l.ForceRegex(x, nil).SetFinal(i)
	}
	l.Start(os.Stdin)
	for !l.Eof() {
		f := l.Next()
		if f == -1 {
			bwlerrors.Fatal(errors.New("failed to match"))
		}
		fmt.Printf("%d (%2d): %#v\n", f, l.Pos(), l.String())
	}
}
