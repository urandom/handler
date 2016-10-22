package log

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/urandom/handler"
)

// PanicDateFormat is the default timestamp format for the panic messages.
const PanicDateFormat = "Jan 2, 2006 at 3:04pm (MST)"

var (
	// ShowStack will cause the stack to be printed out in the
	// http.ResponseWriter. Only used by the Panic handler.
	ShowStack Option = showStack
	showStack        = Option{func(o *options) {
		o.showStack = true
	}}
)

// Panic returns a handler that invokes the passed handler h, catching any
// panics. If one occurs, an HTTP 500 response is produced.
//
// By default, all messages are printed out to os.Stderr.
func Panic(h http.Handler, opts ...Option) http.Handler {
	o := options{logger: handler.ErrLogger(), dateFormat: PanicDateFormat}
	o.apply(opts)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				timestamp := time.Now().Format(o.dateFormat)
				message := fmt.Sprintf("%s - %s\n%s\n", timestamp, rec, stack)

				o.logger.Print(message)

				w.WriteHeader(http.StatusInternalServerError)

				if !o.showStack {
					message = "Internal Server Error"
				}

				w.Write([]byte(message))
			}
		}()

		h.ServeHTTP(w, r)
	})
}
