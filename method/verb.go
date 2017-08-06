package method

import "net/http"

type Verb string

const (
	GET    Verb = "GET"
	POST        = "POST"
	PUT         = "PUT"
	DELETE      = "DELETE"
	HEAD        = "HEAD"
	PATCH       = "PATCH"
)

// HTTP returns a handler that will check each request's method against a
// predefined whitelist. If the request's method is not part of the list,
// the response will be a 400 Bad Request.
func HTTP(h http.Handler, verb Verb, verbs ...Verb) http.Handler {
	verbSet := map[Verb]struct{}{verb: struct{}{}}
	for _, v := range verbs {
		verbSet[v] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := verbSet[Verb(r.Method)]; ok {
			h.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

// Get returns a handler that allows only GET requests to pass.
func Get(h http.Handler) http.Handler {
	return HTTP(h, GET)
}

// Post returns a handler that allows only POST requests to pass.
func Post(h http.Handler) http.Handler {
	return HTTP(h, POST)
}

// Put returns a handler that allows only PUT requests to pass.
func Put(h http.Handler) http.Handler {
	return HTTP(h, PUT)
}

// Delete returns a handler that allows only DELETE requests to pass.
func Delete(h http.Handler) http.Handler {
	return HTTP(h, DELETE)
}
