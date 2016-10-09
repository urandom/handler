package log_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/urandom/handler/log"
)

func TestAccess(t *testing.T) {
	cases := []struct {
		uri     string
		user    string
		method  string
		code    int
		resp    string
		ip      string
		ref     string
		ua      string
		message string
	}{
		{"/", "", "GET", 200, "test1", "1.2.3.4", "ref1", "ua1", "%s - %s [%s] \"%s %s\" %d %d \"%s\" %s"},
		{"/posted", "frank", "POST", 304, "test2.0", "10.0.0.5", "ref2", "ua2", "%s - %s [%s] \"%s %s\" %d %d \"%s\" %s"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			l := &logger{}
			h := log.Access(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.code)
				w.Write([]byte(tc.resp))
			}), log.AccessOpts{Logger: l})

			r, _ := http.NewRequest(tc.method, "http://localhost:8080"+tc.uri, nil)
			rec := httptest.NewRecorder()

			r.Header.Add("Referer", tc.ref)
			r.Header.Add("User-Agent", tc.ua)
			if tc.user != "" {
				r.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(tc.user+":"+"foobar")))
			}
			r.RemoteAddr = tc.ip
			h.ServeHTTP(rec, r)

			m := fmt.Sprintf(tc.message, tc.ip, tc.user, time.Now().Format(log.AccessDateFormat), tc.method, tc.uri, tc.code, len(tc.resp), tc.ref, tc.ua)

			if m != l.message {
				t.Fatalf("expected %s, got %s", m, l.message)
			}
		})
	}
}

type logger struct {
	message string
}

func (l *logger) Print(v ...interface{}) {
	l.message = fmt.Sprint(v...)
}
