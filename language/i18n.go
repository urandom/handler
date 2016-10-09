// +build go1.7

package language

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/urandom/handler"
)

// Provider supplies the I18N handler with supported language codes.
type Provider interface {
	// Languages returns a list of supported language codes.
	Languages() []string
	// Default provides the default language code to use.
	Default() string
	// RequestLanguage has to extract an accepted language from the request.
	// Usually by parsing the 'Accept-Language' header value.
	RequestLanguage(r *http.Request) string
}

// I18NOpts represents the various options for the I18N handler
type I18NOpts struct {
	// Languages provides the handler with supported languages.
	Provider Provider
	// Session will be used to set the current language.
	Session handler.Session
	// UrlPrefix is the prefix that may be used in the url, before the language
	// code itself. For example, if the prefix is '/web', the final url will be
	// '/web/en'.
	UrlPrefix string
	// Logger is used to print out any error messages. If none is provided, no
	// error message will be printed.
	Logger handler.Logger
}

type contextKey string

// ContextValue is stored in the request context
type ContextValue struct {
	// Languages contains all supported languages.
	Languages []string
	// Current is the currently required language.
	Current string
}

// ContextKey is the key under which the the language list and current language
// will be be stored in the request context.
var ContextKey contextKey = "i18n-data"

// SessionKey is the key under which the current session will be stored in the
// session.
var SessionKey string = "language"

// I18N returns a handler that deals with detecting and setting a language for
// use in handler h. It uses a Provider to know what languages are supported,
// and which is the default, or fallback, language. If there is no provider, or
// there is only one supported language, the handler only sets the ContextValue
// in the request before calling the handler h. The current language is set in
// the url, redirecting to it first if it is not there yet. It also strips the
// language from the request path so that other handlers in the chain won't see
// a url they won't expect.
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
// is checked first for a stored language. If none is found, the request is
// passed to the provider, so that it might try to determine a language. It
// then tries the LANG and LC_MESSAGES environment variables. If no language
// has been selected, or the selected one isn't supported, the default language
// determined by the provider is used. With a valid language, a redirect is
// created with the language code added to the url.
func I18N(h http.Handler, o I18NOpts) http.Handler {
	if o.Provider == nil || len(o.Provider.Languages()) < 2 {
		// No point in doing anything if only 1 language is supported. Just
		// provide the empty data
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKey, ContextValue{})))
		})
	}

	if o.UrlPrefix == "" {
		o.UrlPrefix = "/"
	} else if o.UrlPrefix[len(o.UrlPrefix)-1] != '/' {
		o.UrlPrefix = o.UrlPrefix + "/"
	}

	if o.Logger == nil {
		o.Logger = handler.NopLogger{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prefer RequestURI, to preserve url encoded '/'
		uriParts := strings.SplitN(r.RequestURI, "?", 2)
		if uriParts[0] == "" {
			uriParts[0] = r.URL.Path
		}

		data := ContextValue{Languages: o.Provider.Languages()}

		// Check if the url already contains a language and set the context
		// data. Strip the language code from the url for subsequent handlers.
		for _, language := range data.Languages {
			// Redirect to a slash-ending url
			if uriParts[0] == o.UrlPrefix+language {
				url := uriParts[0] + "/"

				if len(uriParts) > 1 && uriParts[1] != "" {
					url += "?" + uriParts[1]
				}

				http.Redirect(w, r, url, http.StatusFound)

				return
			}

			if strings.HasPrefix(uriParts[0], o.UrlPrefix+language+"/") {
				// Strip language code
				r.URL.Path = o.UrlPrefix + r.URL.Path[len(o.UrlPrefix+language+"/"):]
				uriParts[0] = o.UrlPrefix + uriParts[0][len(o.UrlPrefix+language+"/"):]
				r.RequestURI = strings.Join(uriParts, "?")

				data.Current = language

				break
			}
		}

		// Redirect to one of the supported languages
		if data.Current == "" {
			fallback := fallbackLanguage(o.Session, r, o.Provider)
			index := strings.Index(fallback, "-")
			short := fallback
			if index > -1 {
				short = fallback[:index]
			}

			var language string

			// Check if the language is in the list of supported ones
			for _, l := range o.Provider.Languages() {
				if l == fallback || l == short {
					language = l
					break
				}
			}

			if language == "" {
				language = o.Provider.Default()
			}

			url := o.UrlPrefix + language + uriParts[0][len(o.UrlPrefix)-1:]
			if len(uriParts) > 1 && uriParts[1] != "" {
				url += "?" + uriParts[1]
			}

			http.Redirect(w, r, url, http.StatusFound)

			return
		}

		if o.Session != nil {
			if err := o.Session.Set(r, SessionKey, data.Current); err != nil {
				o.Logger.Print("i18n handler: " + err.Error())
			}
		}

		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ContextKey, data)))
	})
}

func fallbackLanguage(sess handler.Session, r *http.Request, provider Provider) string {
	if sess != nil {
		if val, err := sess.Get(r, SessionKey); err == nil && val != nil {
			if v, ok := val.(string); ok {
				return v
			}
		}
	}

	if l := provider.RequestLanguage(r); l != "" {
		return l
	}

	language := os.Getenv("LANG")

	if language == "" {
		language = os.Getenv("LC_MESSAGES")
	}

	if language == "" {
		language = provider.Default()
	} else {
		index := strings.LastIndex(language, ".")
		if index != -1 {
			language = language[:index]
		}

	}

	return strings.ToLower(strings.Replace(language, "_", "-", -1))
}
