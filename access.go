package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

var AccessDateFormat = "Jan 2, 2006 at 3:04pm (MST)"

// AccessLogger is used to print out access log messages.
type AccessLogger interface {
	Print(v ...interface{})
}

// Access returns a handler that writes an access log message to the provided
// logger whenever the handler h is invoked. The log message is of the
// following format:
//
// IP - USER [DATETIME] "HTTP_METHOD URI" STATUS_CODE BODY_LENGTH "REFERER" USER_AGENT
//
// If the logger l is nil, the function panics.
func Access(l AccessLogger, h http.Handler) http.Handler {
	if l == nil {
		// Panic on incorrect usage
		panic("logger is nil")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := NewResponseWrapper(w)

		uri := r.URL.RequestURI()
		remoteAddr := remoteAddr(r)
		remoteUser := ""
		method := r.Method
		referer := r.Header.Get("Referer")
		userAgent := r.Header.Get("User-Agent")

		h.ServeHTTP(wrapper, r)

		for k, v := range wrapper.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(wrapper.Code)
		w.Write(wrapper.Body.Bytes())

		timestamp := time.Now().Format(AccessDateFormat)
		code := wrapper.Code
		length := wrapper.Body.Len()

		l.Print(fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" %s",
			remoteAddr, remoteUser, timestamp, method, uri, code, length, referer, userAgent))
	})
}

func remoteAddr(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with ","
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		return parts[0]
	}
	return hdrRealIp
}

func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}
