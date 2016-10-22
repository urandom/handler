// +build go1.7

package lang

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/urandom/handler"
	xlang "golang.org/x/text/language"
)

type options struct {
	languages []xlang.Tag
	session   handler.Session
	urlPrefix string
	logger    handler.Logger
}

// An Option is used to change the default behaviour of the language handlers.
type Option struct {
	f func(o *options)
}

// Languages provides the handler with supported languages.
func Languages(tags []xlang.Tag) Option {
	return Option{func(o *options) {
		o.languages = tags
	}}
}

// Session will be used to set the current language.
func Session(s handler.Session) Option {
	return Option{func(o *options) {
		o.session = s
	}}
}

// URLPrefix is the prefix that may be used in the url, before the language
// code itself. For example, if the prefix is '/web', the final url will be
// '/web/en'.
func URLPrefix(p string) Option {
	return Option{func(o *options) {
		o.urlPrefix = p
	}}
}

// Logger is used to print out any error messages. If none is provided, no
// error message will be printed.
func Logger(l handler.Logger) Option {
	return Option{func(o *options) {
		o.logger = l
	}}
}

type contextKey string

// ContextValue is stored in the request context
type ContextValue struct {
	// Languages contains all supported languages.
	Languages []xlang.Tag
	// Current is the currently required language.
	Current xlang.Tag
}

// ContextKey is the key under which the the language list and current language
// will be be stored in the request context.
var ContextKey contextKey = "i18n-data"

// SessionKey is the key under which the current session will be stored in the
// session.
var SessionKey string = "language"

// I18N returns a handler that deals with detecting and setting a language for
// use in handler h. If there is only one supported language, the handler only
// sets the ContextValue in the request before calling the handler h. The
// current language is set in the url, redirecting to it first if it is not
// there yet. It also strips the language from the request path so that other
// handlers in the chain won't see a url they won't expect.
//
// The handler first checks whether the request url already contains any of the
// supported languages. If one matches, but ends with the language code and
// without a terminating slash, a redirect is sent. Example:
//
// '/en' -> '/en/'. Or if a UrlPRefix is set: '/prefix/en' -> '/prefix/en/'
//
// If the path contains the language and additional path data, that language is
// stored as the current language in the request context. It is also stored in
// the session, if such an interface is provided.
//
// If the url contains no language code, several methods are attempted to
// decide what the language should be. If a session interface is provided, it
// is checked first for a stored language. If none is found, the
// Accept-Language header is checked for a suitable choice. It then tries the
// LANG and LC_MESSAGES environment variables. If no language has been
// selected, or the selected one isn't supported, the first language in the
// supported slice is used. With a valid language, a redirect is created with
// the language code added to the url.
//
// By default, error messages will not be printed out.
func I18N(h http.Handler, opts ...Option) http.Handler {
	o := options{logger: handler.OutLogger()}
	o.apply(opts)

	if len(o.languages) < 2 {
		// No point in doing anything if only 1 language is supported. Just
		// provide the empty data
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKey, ContextValue{})))
		})
	}

	if o.urlPrefix == "" {
		o.urlPrefix = "/"
	} else if o.urlPrefix[len(o.urlPrefix)-1] != '/' {
		o.urlPrefix = o.urlPrefix + "/"
	}

	if o.logger == nil {
		o.logger = handler.NopLogger()
	}

	matcher := xlang.NewMatcher(o.languages)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prefer RequestURI, to preserve url encoded '/'
		uriParts := strings.SplitN(r.RequestURI, "?", 2)
		if uriParts[0] == "" {
			uriParts[0] = r.URL.Path
		}

		if !strings.HasPrefix(uriParts[0], o.urlPrefix) {
			h.ServeHTTP(w, r)
			return
		}

		sub := uriParts[0][len(o.urlPrefix):]
		slashIndex := strings.Index(sub, "/")

		var uriTag xlang.Tag

		if slashIndex == -1 {
			uriTag = xlang.Make(sub)
		} else {
			uriTag = xlang.Make(sub[:slashIndex])
		}

		tag, _, c := matcher.Match(uriTag)

		if c != xlang.Exact {
			// Redirect to one of the supported languages
			url := o.urlPrefix

			if c == xlang.No {
				tag, _, _ = matcher.Match(fallbackLanguage(o.session, r))
				url += tag.String() + "/" + sub
			} else {
				url += tag.String() + "/" + sub[slashIndex+1:]
			}

			if len(uriParts) > 1 && uriParts[1] != "" {
				url += "?" + uriParts[1]
			}

			http.Redirect(w, r, url, http.StatusFound)
		} else if slashIndex == -1 {
			// Redirect to a slash-ending url
			url := o.urlPrefix + tag.String() + "/"
			if slashIndex != -1 {
				url += sub[slashIndex+1:]
			}

			if len(uriParts) > 1 && uriParts[1] != "" {
				url += "?" + uriParts[1]
			}

			http.Redirect(w, r, url, http.StatusFound)

			return
		} else {
			data := ContextValue{Languages: o.languages}

			// Strip language code
			r.URL.Path = o.urlPrefix + r.URL.Path[len(o.urlPrefix+tag.String()+"/"):]
			uriParts[0] = o.urlPrefix + uriParts[0][len(o.urlPrefix+tag.String()+"/"):]
			r.RequestURI = strings.Join(uriParts, "?")

			data.Current = tag

			if o.session != nil {
				if err := o.session.Set(r, SessionKey, tag.String()); err != nil {
					o.logger.Print("i18n handler: " + err.Error())
				}
			}

			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKey, data)))
		}
	})
}

func fallbackLanguage(sess handler.Session, r *http.Request) xlang.Tag {
	if sess != nil {
		if val, err := sess.Get(r, SessionKey); err == nil && val != nil {
			if v, ok := val.(string); ok {
				return xlang.Make(v)
			}
		}
	}

	if tags, _, err := xlang.ParseAcceptLanguage(
		r.Header.Get("Accept-Language"),
	); err == nil && len(tags) > 0 {
		return tags[0]
	}

	language := os.Getenv("LANG")

	if language == "" {
		language = os.Getenv("LC_MESSAGES")
	}

	return xlang.Make(language)
}

func (o *options) apply(opts []Option) {
	for _, op := range opts {
		op.f(o)
	}
}
