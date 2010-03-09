/*
	A demonstration of the regular expression facility in the lexer library.
*/

package main

import (
	"os"
	"./lexer"
	"./errors"
	"io/ioutil"
	"fmt"
)

func main() {
	bs, err := ioutil.ReadAll(os.Stdin)
	errors.Fatal(err)
	for _, x := range lexer.Matches(os.Args[1], string(bs)) {
		fmt.Printf("%s\n", x)
	}
}


