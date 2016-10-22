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

type options struct {
	logger     handler.Logger
	dateFormat string
	showStack  bool
}

// An Option is used to change the default behaviour of logging handlers.
type Option struct {
	f func(o *options)
}

// Logger defines the logger to be used whenever detailed messages have to be
// printed out.
func Logger(l handler.Logger) Option {
	return Option{func(o *options) {
		o.logger = l
	}}
}

// DateFormat is used to format the timestamp.
func DateFormat(f string) Option {
	return Option{func(o *options) {
		o.dateFormat = f
	}}
}

// Access returns a handler that writes an access log message to the provided
// logger, provided by the options, whenever the handler h is invoked. The log
// message is of the following format:
//
// IP - USER [DATETIME] "HTTP_METHOD URI" STATUS_CODE BODY_LENGTH "REFERER" USER_AGENT
//
// By default, all messages are printed to os.Stdout.
func Access(h http.Handler, opts ...Option) http.Handler {
	o := options{logger: handler.OutLogger(), dateFormat: AccessDateFormat}
	o.apply(opts)

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

		timestamp := time.Now().Format(o.dateFormat)
		code := wrapper.Code
		length := wrapper.Body.Len()

		o.logger.Print(fmt.Sprintf("%s - %s [%s] \"%s %s\" %d %d \"%s\" %s",
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

func (o *options) apply(opts []Option) {
	for _, op := range opts {
		op.f(o)
	}
}
