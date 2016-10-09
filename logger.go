package handler

import (
	"fmt"
	"os"
)

// A Logger interface is used by handlers to when some kind of output needs to be provided
type Logger interface {
	Print(v ...interface{})
}

// NopLogger returns a Logger implementation that does nothing
func NopLogger() Logger {
	return nopLogger{}
}

// OutLogger returns a Logger implementation that outputs to os.Stdout
func OutLogger() Logger {
	return outLogger{}
}

// ErrLogger returns a Logger implementation that outputs to os.Stderr
func ErrLogger() Logger {
	return errLogger{}
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
