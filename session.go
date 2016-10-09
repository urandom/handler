package handler

import "net/http"

// Session allows storing arbitrary data in between requests.
type Session interface {
	Get(r *http.Request, key string) (interface{}, error)
	Set(r *http.Request, key string, value interface{}) error
}
