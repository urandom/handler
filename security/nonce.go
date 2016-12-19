package security

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/urandom/handler"
)

type nonceStatus int
type ctxKey string

const (
	// NonceNotRequested states that the request doesn't contain nonce data.
	NonceNotRequested nonceStatus = iota
	// NonceValid states that the request nonce exists and is valid.
	NonceValid
	// NonceInvalid states that the request nonce doesn't exist or has expired.
	NonceInvalid

	nonceValueKey  = ctxKey("nonce")
	nonceSetterKey = ctxKey("nonce-gen")
)

var (
	// TimeRandomGenerator creates string content for a nonce using
	// the current time and a random integer.
	TimeRandomGenerator Option = Option{func(o *options) {
		o.generator = timeRandomGenerator
	}}
	// HeaderGetter extracts an existing nonceis valid from a request's X-Nonce
	// header.
	HeaderGetter Option = Option{func(o *options) {
		o.getter = headerGetter
	}}

	HeaderSetter Option = Option{func(o *options) {
		o.setter = headerSetter
	}}
)

type NonceStatus struct {
	Status nonceStatus
}

type options struct {
	logger    handler.Logger
	generator func(io.Writer) error
	getter    func(*http.Request) string
	setter    func(string, http.ResponseWriter, *http.Request)
	age       time.Duration
}

// Logger defines the logger to be used whenever detailed messages have to be
// printed out.
func Logger(l handler.Logger) Option {
	return Option{func(o *options) {
		o.logger = l
	}}
}

// Age sets the maximum time duration a nonce can be valid
func Age(age time.Duration) Option {
	return Option{func(o *options) {
		o.age = age
	}}
}

// An Option is used to change the default behaviour of nonce handler.
type Option struct {
	f func(o *options)
}

type setter func(http.ResponseWriter, *http.Request) error
type nonceStore map[string]int64

func Nonce(h http.Handler, opts ...Option) http.Handler {
	o := options{
		logger:    handler.OutLogger(),
		generator: timeRandomGenerator,
		getter:    headerGetter,
		setter:    headerSetter,
		age:       45 * time.Second,
	}
	o.apply(opts)

	store := nonceStore{}
	opChan := make(chan func(nonceStore))

	go func() {
		for op := range opChan {
			op(store)
		}
	}()

	go func() {
		for {
			select {
			case <-time.After(5 * time.Minute):
				cleanup(o.age, opChan)
			}
		}
	}()

	setter := func(w http.ResponseWriter, r *http.Request) error {
		nonce, err := generateAndStore(o.age, o.generator, opChan)
		if err != nil {
			return err
		}

		o.setter(nonce, w, r)
		return nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		nonce := o.getter(r)
		if nonce != "" {
			if validateAndRemoveNonce(nonce, o.age, opChan) {
				ctx = context.WithValue(ctx, nonceValueKey, NonceStatus{NonceValid})
			} else {
				ctx = context.WithValue(ctx, nonceValueKey, NonceStatus{NonceInvalid})
			}
		}

		h.ServeHTTP(w, r.WithContext(context.WithValue(ctx, nonceSetterKey, setter)))
	})
}

func NonceValueFromRequest(r *http.Request) NonceStatus {
	if c := r.Context().Value(nonceValueKey); c != nil {
		if v, ok := c.(NonceStatus); ok {
			return v
		}
	}

	return NonceStatus{NonceNotRequested}
}

func StoreNonce(w http.ResponseWriter, r *http.Request) (err error) {
	if c := r.Context().Value(nonceSetterKey); c != nil {
		if setter, ok := c.(setter); ok {
			err = setter(w, r)
		}
	}

	return err
}

func (s NonceStatus) Valid() bool {
	return s.Status == NonceValid
}

func (o *options) apply(opts []Option) {
	for _, op := range opts {
		op.f(o)
	}
}

func timeRandomGenerator(w io.Writer) error {
	for _, s := range []string{
		strconv.FormatInt(time.Now().Unix(), 32),
		strconv.FormatInt(rand.Int63(), 32),
	} {
		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
	}

	return nil
}

func headerGetter(r *http.Request) string {
	return r.Header.Get("X-Nonce")
}

func headerSetter(nonce string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Nonce", nonce)
}

func validateAndRemoveNonce(nonce string, age time.Duration, opChan chan<- func(nonceStore)) bool {
	res := make(chan bool, 1)
	opChan <- func(store nonceStore) {
		now := time.Now().Unix()

		if t, ok := store[nonce]; ok {
			delete(store, nonce)
			if now-t <= int64(age) {
				res <- true
				return
			}
		}

		res <- false
	}

	return <-res
}

func generateAndStore(age time.Duration, generator func(w io.Writer) error, opChan chan<- func(nonceStore)) (string, error) {
	type result struct {
		nonce string
		err   error
	}
	res := make(chan result, 1)
	opChan <- func(store nonceStore) {
		h := md5.New()

		if err := generator(h); err != nil {
			res <- result{"", err}
			return
		}

		nonce := fmt.Sprintf("%x", h.Sum(nil))

		store[nonce] = time.Now().Unix()

		res <- result{nonce, nil}
	}

	r := <-res
	return r.nonce, r.err
}

func cleanup(age time.Duration, opChan chan<- func(nonceStore)) {
	opChan <- func(store nonceStore) {
		now := time.Now().Unix()

		for nonce, t := range store {
			if now-t > int64(age) {
				delete(store, nonce)
				return
			}
		}
	}
}
