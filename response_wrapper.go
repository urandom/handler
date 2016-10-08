package handler

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
)

// ResponseWrapper is a simple wrapper around httptest.ResponseRecorder,
// implementing the http.Hijacker interface.
type ResponseWrapper struct {
	*httptest.ResponseRecorder

	// Hijacked will be set to true if the original http.ResponseWriter was
	// hijacked successfully.
	Hijacked bool

	writer http.ResponseWriter
}

// NewResponseWrapper creates a new wrapper. The passed http.ResponseWriter is
// used in case the wrapper needs to be hijacked.
func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{ResponseRecorder: httptest.NewRecorder(), writer: w}
}

// Hijack tries to use the original http.ResponseWriter for hijacking. If the
// original writer doesn't implement http.Hijacker, it returns an error.
func (w *ResponseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.writer.(http.Hijacker); ok {
		c, rw, err := hijacker.Hijack()

		if err == nil {
			w.Hijacked = true
		}

		return c, rw, err
	}

	return nil, nil, errors.New("Wrapped ResponseWriter is not a Hijacker")
}
