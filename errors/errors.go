package errors

import (
	"os"
	"fmt"
)

func Fatal(err os.Error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		panic("panic")
	}
}

func Catch(f func(), g func(interface{})) {
	defer func() {
		if v := recover(); v != nil {
			g(v)
		}
	}()
	f()
}
