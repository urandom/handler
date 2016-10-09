package lang_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"

	"github.com/urandom/handler/lang"
)

func TestI18N(t *testing.T) {
	cases := []struct {
		langs  []string
		prefix string
		real   string
		exp    string
		header string
	}{
		{[]string{"en", "de", "fr"}, "", "/", "/", ""},
		{nil, "/foo", "/", "/foo/", ""},
		{[]string{}, "", "/", "/", ""},
		{[]string{"en", "de", "fr"}, "/bar", "/", "/bar/", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "/bar", "/en", "/bar/", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "/bar", "/en/foo", "/bar/foo", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "", "/en", "/", "de, en-gb;q=0.8, en;q=0.7"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			langs := make([]language.Tag, len(tc.langs))
			for i, l := range tc.langs {
				langs[i] = language.Make(l)
			}
			ts := httptest.NewServer(lang.I18N(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.exp {
					t.Fatalf("expected url path %s, got %s", tc.exp, r.URL.Path)
				}
				if r.RequestURI != tc.exp {
					t.Fatalf("expected url path %s, got %s", tc.exp, r.URL.Path)
				}
			}), lang.I18NOpts{Languages: langs, UrlPrefix: tc.prefix}))
			defer ts.Close()

			r, _ := http.NewRequest("GET", ts.URL+tc.prefix+tc.real, nil)

			r.Header.Set("Accept-Language", tc.header)
			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Fatalf("test server request: %s", err)
			}

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("http status: %d", resp.StatusCode)
			}
		})
	}
}
