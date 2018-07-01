package encoding

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/urandom/handler"
)

type options struct {
	logger handler.Logger
}

// An Option is used to change the default behaviour of the encoding handlers.
type Option struct {
	f func(o *options)
}

// Logger is used to print out any error messages during compression. If
// none is provided, no error message will be printed.
func Logger(l handler.Logger) Option {
	return Option{func(o *options) {
		o.logger = l
	}}
}

// Gzip returns a handler that will use gzip compression on the response body
// of handler h. Compression will only be applied if the request contains an
// 'Accept-Encoding' header that contains 'gzip'.
//
// By default, no messages are printed out.
func Gzip(h http.Handler, opts ...Option) http.Handler {
	o := options{logger: handler.OutLogger()}
	o.apply(opts)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		wrapper := handler.NewResponseWrapper(w)

		h.ServeHTTP(wrapper, r)

		for k, v := range wrapper.Header() {
			w.Header()[k] = v
		}
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")

		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", http.DetectContentType(wrapper.Body.Bytes()))
		}
		w.Header().Del("Content-Length")

		if wrapper.Code != http.StatusOK {
			w.WriteHeader(wrapper.Code)
		}

		gz := gzip.NewWriter(w)
		gz.Flush()

		if _, err := gz.Write(wrapper.Body.Bytes()); err != nil {
			o.logger.Print("gzip handler: " + err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		gz.Close()
	})
}

func (o *options) apply(opts []Option) {
	for _, op := range opts {
		op.f(o)
	}
}
