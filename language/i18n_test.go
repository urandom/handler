package language_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/urandom/handler/language"
)

func TestI18N(t *testing.T) {
	cases := []struct {
		langs  []string
		def    string
		prefix string
		real   string
		exp    string
		header string
	}{
		{[]string{"en", "de", "fr"}, "en", "", "/", "/", ""},
		{nil, "", "/foo", "/", "/foo/", ""},
		{[]string{}, "", "", "/", "/", ""},
		{[]string{"en", "de", "fr"}, "en", "/bar", "/", "/bar/", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "en", "/bar", "/en", "/bar/", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "en", "/bar", "/en/foo", "/bar/foo", "de, en-gb;q=0.8, en;q=0.7"},
		{[]string{"en", "de", "fr"}, "en", "", "/en", "/", "de, en-gb;q=0.8, en;q=0.7"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			p := provider{langs: tc.langs, def: tc.def}

			ts := httptest.NewServer(language.I18N(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tc.exp {
					t.Fatalf("expected url path %s, got %s", tc.exp, r.URL.Path)
				}
				if r.RequestURI != tc.exp {
					t.Fatalf("expected url path %s, got %s", tc.exp, r.URL.Path)
				}
			}), language.I18NOpts{Provider: p, UrlPrefix: tc.prefix}))
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

type provider struct {
	langs []string
	def   string
}

func (p provider) Languages() []string {
	return p.langs
}

func (p provider) Default() string {
	return p.def
}

func (p provider) RequestLanguage(r *http.Request) string {
	accepted := r.Header.Get("Accept-Language")

	fields := strings.Split(accepted, ",")

	if len(fields) > 0 {
		return fields[0]
	}

	return ""
}
