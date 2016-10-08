package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

var PanicDateFormat = "Jan 2, 2006 at 3:04pm (MST)"

type PanicLogger interface {
	Print(v ...interface{})
}

// Panic returns a handler that invokes the passed handler h, catching any
// panics. If one occurs, an HTTP 500 response is produced. If the logger l is
// not nil, it will be used to print out a detailed message, including the
// timestamp and stack trace. If showStack is true, the detailed message is
// also written to the ResponseWriter.
func Panic(l PanicLogger, showStack bool, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				timestamp := time.Now().Format(PanicDateFormat)
				message := fmt.Sprintf("%s - %s\n%s\n", timestamp, rec, stack)

				if l != nil {
					l.Print(message)
				}

				w.WriteHeader(http.StatusInternalServerError)

				if !showStack {
					message = "Internal Server Error"
				}

				w.Write([]byte(message))
			}
		}()

		h.ServeHTTP(w, r)
	})
}
