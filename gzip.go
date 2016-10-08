package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type GzipOpts struct {
	// Logger is used to print out any error messages during compression. If
	// none is provided, no error message will be printed.
	Logger Logger
}

// Gzip returns a handler that will use gzip compression on the response body
// of handler h. Compression will only be applied if the request contains an
// 'Accept-Encoding' header that contains 'gzip'.
func Gzip(h http.Handler, o GzipOpts) http.Handler {
	if o.Logger == nil {
		o.Logger = nopLogger{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		wrapper := NewResponseWrapper(w)

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

		gz := gzip.NewWriter(w)
		gz.Flush()

		if _, err := gz.Write(wrapper.Body.Bytes()); err != nil {
			o.Logger.Print("gzip handler: " + err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		gz.Close()

		w.WriteHeader(wrapper.Code)
	})
}
