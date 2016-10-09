package handler

import (
	"fmt"
	"os"
)

// A Logger interface is used by handlers to when some kind of output needs to be provided
type Logger interface {
	Print(v ...interface{})
}

// NopLogger is a Logger implementation that does nothing
type NopLogger struct{}

func (l NopLogger) Print(v ...interface{}) {}

// OutLogger is a Logger implementation that outputs to os.Stdout
type OutLogger struct{}

func (l OutLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stdout, v...)
}

// ErrLogger is a Logger implementation that outputs to os.Stderr
type ErrLogger struct{}

func (l ErrLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stderr, v...)
}
