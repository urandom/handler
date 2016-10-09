# handler [![Build Status](https://travis-ci.org/urandom/handler.png?branch=master)](https://travis-ci.org/urandom/handler) [![GoDoc](http://godoc.org/github.com/urandom/handler?status.png)](http://godoc.org/github.com/urandom/handler)
http.Handlers for Go

## Domains

Handlers are separated into their own domains. They are:

* [log](https://godoc.org/github.com/urandom/handler/log) - handlers used for logging purposes
  * Access - logs each request to the provided logger.
  * Panic - catches panics, logs the stack traces to a provided logger, and returns an Internal Server Error. Optionally prints the stack trace in the response body.
* [encoding](https://godoc.org/github.com/urandom/handler/encoding) - handlers dealing with encoding
  * Gzip - compresses the response body
* [lang](https://godoc.org/github.com/urandom/handler/lang) - handlers for language/translation support
  * I18N - deals with language handling, redirecting to a url with a supported language code. Provides the supported languages and current one in the request context.
  
## Example

### Using the [I18N](https://godoc.org/github.com/urandom/handler/lang#I18N) handler to process a requested language.

When a user visits '/', a language will be picked from the list of supported ones, based on what the browser has set in the Accepted-Language header, or fall back to the first one in the list. If, for example, the selected language was English, a redirect will be sent to '/en/', where the final handler extracts the selected language from the request context.

```go
package main

import (
	"net/http"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"

	"github.com/urandom/handler/lang"
)

func main() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Context().Value(lang.ContextKey).(lang.ContextValue)

		w.Write([]byte("Current language: " + display.Self.Name(val.Current)))
	})
  
	http.Handle("/", lang.I18N(handler, lang.I18NOpts{
		Languages: []language.Tag{language.German, language.French, language.English},
	}))
  
	http.ListenAndServe(":8080", nil)
}

```
