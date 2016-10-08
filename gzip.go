package handler

import (
	"compress/gzip"
	"net/http"
	"strings"
)

func Gzip(h http.Handler) http.Handler {
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
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		gz.Close()

		w.WriteHeader(wrapper.Code)
	})
}
