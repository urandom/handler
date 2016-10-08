package handler_test

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urandom/handler"
)

type hijackWriter struct {
	*httptest.ResponseRecorder
	hijacked bool
}

func TestResponseWrapper(t *testing.T) {
	hw := &hijackWriter{httptest.NewRecorder(), false}

	var wrapper http.ResponseWriter = handler.NewResponseWrapper(hw)

	if hj, ok := wrapper.(http.Hijacker); ok {
		if _, _, err := hj.Hijack(); err != nil {
			t.Fatalf("error wasn't expected: %s", err)
		}

		if !hw.hijacked {
			t.Fatalf("writer wasn't hijacked")
		}

		if !wrapper.(*handler.ResponseWrapper).Hijacked {
			t.Fatalf("wrapper isn't marked as hijacked")
		}
	} else {
		t.Fatalf("wrapper isn't a hijacker")
	}
}

func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, nil
}
