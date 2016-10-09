package log_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/urandom/handler"
	"github.com/urandom/handler/log"
)

func TestPanic(t *testing.T) {
	cases := []struct {
		code      int
		showStack bool
		body      string
	}{
		{
			http.StatusInternalServerError,
			false,
			"Internal Server Error",
		},
		{
			http.StatusInternalServerError,
			true,
			fmt.Sprintf("%s - %s", time.Now().Format(log.PanicDateFormat), "Test"),
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			h := log.Panic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("Test")
			}), log.PanicOpts{handler.NopLogger{}, tc.showStack, ""})

			r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, r)

			if rec.Code != tc.code {
				t.Fatalf("got code %d, wanted %d", rec.Code, tc.code)
			}

			if tc.showStack {
				if !strings.HasPrefix(rec.Body.String(), tc.body) {
					t.Fatalf("got body %s, expected prefix %s", rec.Body.String(), tc.body)
				}
			} else {
				if rec.Body.String() != tc.body {
					t.Fatalf("got body %s, expected %s", rec.Body.String(), tc.body)
				}
			}
		})
	}

}
