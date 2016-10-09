package log

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/urandom/handler"
)

// AccessDateFormat is the default timestamp format for the access log messages.
const AccessDateFormat = "Jan 2, 2006 at 3:04pm (MST)"

type AccessOpts struct {
	// Logger will be used to print out an entry whenever a request is handled.
	// If none is provded, os.Stdout is used.
	Logger handler.Logger
	// DateFormat is used to format the timestamp. Defaults to AccessDateFormat.
	DateFormat string
}

// Access returns a handler that writes an access log message to the provided
// logger, provided by the options, whenever the handler h is invoked. The log
// message is of the following format:
//
// IP - USER [DATETIME] "HTTP_METHOD URI" STATUS_CODE BODY_LENGTH "REFERER" USER_AGENT
func Access(h http.Handler, o AccessOpts) http.Handler {
	if o.Logger == nil {
		o.Logger = handler.OutLogger()
	}
	if o.DateFormat == "" {
		o.DateFormat = AccessDateFormat
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wrapper := handler.NewResponseWrapper(w)

		uri := r.URL.RequestURI()
		remoteAddr := remoteAddr(r)
		remoteUser := remoteUser(r)
		method := r.Method
		referer := r.Header.Get("Referer")
		userAgent := r.Header.Get("User-Agent")

		h.ServeHTTP(wrapper, r)

		for k, v := range wrapper.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(wrapper.Code)
		w.Write(wrapper.Body.Bytes())

		timestamp := time.Now().Format(o.DateFormat)
		code := wrapper.Code
		length := wrapper.Body.Len()

		o.Logger.Print(fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" %s",
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

func remoteUser(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		fields := strings.Fields(auth)
		if len(fields) > 1 {
			decoded, err := base64.StdEncoding.DecodeString(fields[1])

			if err != nil {
				return ""
			}

			creds := strings.Split(string(decoded), ":")
			if len(creds) > 1 {
				return creds[0]
			}
		}
	}

	return ""
}
