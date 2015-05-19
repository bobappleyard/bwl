/*
	A demonstration of the regular expression facility in the lexer library.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bobappleyard/bwl/errors"
	"github.com/bobappleyard/bwl/lexer"
)

func main() {
	bs, err := ioutil.ReadAll(os.Stdin)
	errors.Fatal(err)
	for _, x := range lexer.Matches(os.Args[1], string(bs)) {
		fmt.Printf("%s\n", x)
	}
}
