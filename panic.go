package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// PanicDateFormat is the default timestamp format for the panic messages.
const PanicDateFormat = "Jan 2, 2006 at 3:04pm (MST)"

// PanicOpts represents the various options for the Panic handler
type PanicOpts struct {
	// Logger will be used to print out detailed message whenever a panic is
	// recovered. Each message includes a stack trace and timestamp. If none is
	// provded, os.Stderr is used.
	Logger Logger
	// ShowStack will print the stack in the given http.ResponseWriter if true.
	ShowStack bool
	// DateFormat is used to format the timestamp. Defaults to PanicDateFormat.
	DateFormat string
}

// Panic returns a handler that invokes the passed handler h, catching any
// panics. If one occurs, an HTTP 500 response is produced.
func Panic(h http.Handler, o PanicOpts) http.Handler {
	if o.Logger == nil {
		o.Logger = errLogger{}
	}
	if o.DateFormat == "" {
		o.DateFormat = PanicDateFormat
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				timestamp := time.Now().Format(o.DateFormat)
				message := fmt.Sprintf("%s - %s\n%s\n", timestamp, rec, stack)

				o.Logger.Print(message)

				w.WriteHeader(http.StatusInternalServerError)

				if !o.ShowStack {
					message = "Internal Server Error"
				}

				w.Write([]byte(message))
			}
		}()

		h.ServeHTTP(w, r)
	})
}
