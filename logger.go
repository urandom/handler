package handler

import (
	"fmt"
	"os"
)

// A Logger interface is used by handlers to when some kind of output needs to be provided
type Logger interface {
	Print(v ...interface{})
}

type nopLogger struct{}

func (l nopLogger) Print(v ...interface{}) {}

type outLogger struct{}

func (l outLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stdout, v...)
}

type errLogger struct{}

func (l errLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stderr, v...)
}
