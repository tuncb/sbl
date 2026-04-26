package console

import (
	"fmt"
	"io"
)

const ErrorPrefix = "\x1b[31m[Error]\x1b[0m "

func Errorf(out io.Writer, format string, args ...any) {
	fmt.Fprintf(out, ErrorPrefix+format, args...)
}
