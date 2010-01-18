package errors

import (
	"os"
	"fmt"
)

func Fatal(err os.Error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(-1)
	}
}

