package handler

import "net/http"

// Session allows storing arbitrary data in between requests.
type Session interface {
	Get(r *http.Request, key string) (string, error)
	Set(r *http.Request, key string, value string) error
}
