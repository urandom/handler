package encoding_test

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urandom/handler/encoding"
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
			h := encoding.Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tc.content))
			}))

			r, _ := http.NewRequest("GET", "http://localhost:8080", nil)
			rec := httptest.NewRecorder()

			if tc.encode {
				r.Header.Add("Accept-Encoding", "gzip")
			}

			h.ServeHTTP(rec, r)

			if tc.encode {
				gz, err := gzip.NewReader(rec.Body)
				if err != nil {
					t.Fatalf("gzip reader create: %s", err)
				}

				b, err := ioutil.ReadAll(gz)
				if err != nil {
					t.Fatalf("gzip reader: %s", err)
				}

				gz.Close()

				if string(b) != tc.content {
					t.Fatalf("expected gzipped %v, got %v", tc.content, string(b))
				}
			} else {
				if rec.Body.String() != tc.content {
					t.Fatalf("expected uncompressed %s, got %s", tc.content, rec.Body.String())
				}
			}
		})
	}
}
