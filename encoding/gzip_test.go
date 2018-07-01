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
		code    int
	}{
		{content: "Test 1", encode: false},
		{content: "Test 2", encode: true},
		{content: "Test 2", encode: true, code: http.StatusForbidden},
		{content: "Test 4 something long", encode: false},
		{content: "Test 3 something long", encode: true, code: http.StatusBadRequest},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			if tc.code == 0 {
				tc.code = http.StatusOK
			}
			h := encoding.Gzip(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
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

				if rec.Code != tc.code {
					t.Errorf("expected gzipped code %v, got %v", tc.code, rec.Code)
				}

				if string(b) != tc.content {
					t.Fatalf("expected gzipped %v, got %v", tc.content, string(b))
				}
			} else {
				if rec.Code != tc.code {
					t.Errorf("expected uncompressed code %v, got %v", tc.code, rec.Code)
				}

				if rec.Body.String() != tc.content {
					t.Fatalf("expected uncompressed %s, got %s", tc.content, rec.Body.String())
				}
			}
		})
	}
}
