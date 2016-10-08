package handler_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urandom/handler"
)

func TestGzip(t *testing.T) {
	cases := []struct {
		content string
		encode  bool
	}{
		{"Test 1", false},
		{"Test 2", true},
		{"Test 4 something long", false},
		{"Test 3 something long", true},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			h := handler.Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.content))
			}))

			r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			rec := httptest.NewRecorder()

			if tc.encode {
				r.Header.Add("Accept-Encoding", "gzip")
			}

			h.ServeHTTP(rec, r)

			if tc.encode {
				var buf bytes.Buffer
				gz := gzip.NewWriter(&buf)

				if _, err := gz.Write([]byte(tc.content)); err != nil {
					t.Fatalf("gzip error: %s", err)
				}

				gz.Close()

				if !bytes.Equal(buf.Bytes(), rec.Body.Bytes()) {
					t.Fatalf("expected gzipped %v, got %v", buf.Bytes(), rec.Body.Bytes())
				}
			} else {
				if rec.Body.String() != tc.content {
					t.Fatalf("expected uncompressed %s, got %s", tc.content, rec.Body.String())
				}
			}
		})
	}
}
