package handler

import (
	"fmt"
	"os"
)

// A Logger interface is used by handlers to when some kind of output needs to be provided
type Logger interface {
	Print(v ...interface{})
}

type NopLogger struct{}

func (l NopLogger) Print(v ...interface{}) {}

type OutLogger struct{}

func (l OutLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stdout, v...)
}

type ErrLogger struct{}

func (l ErrLogger) Print(v ...interface{}) {
	fmt.Fprint(os.Stderr, v...)
}
