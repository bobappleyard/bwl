package errors

import (
	"os"
	"fmt"
)

func Fatal(err os.Error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		panic()
	}
}

