package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/urandom/handler"
)

// Authenticator is responsible for specifying whether a user-password tuple is
// valid.
type Authenticator interface {
	// Authenticate returns true if the supplied username and password are valid
	Authenticate(user, password string) bool
}

// The AuthenticatorFunc type is an adapter to allow using ordinary functions
// as authenticators.
type AuthenticatorFunc func(user, password string) bool

func (a AuthenticatorFunc) Authenticate(user, password string) bool {
	return a(user, password)
}

// TokenValidator checks if the given token is valid.
type TokenValidator interface {
	Validate(token string, claims jwt.Claims) bool
}

// The TokenValidatorFunc type is an adapter to allow using ordinary functions
// as token storage.
type TokenValidatorFunc func(token string, claims jwt.Claims) bool

func (t TokenValidatorFunc) Validate(token string, claims jwt.Claims) bool {
	return t(token, claims)
}

// TokenStorage stores the given token string along with its expiration time.
type TokenStorage interface {
	Store(token string, expiration time.Time) error
}

// The TokenStorageFunc type is an adapter to allow using ordinary functions
// as token storage.
type TokenStorageFunc func(token string, expiration time.Time) error

func (t TokenStorageFunc) Store(token string, expiration time.Time) error {
	return t(token, expiration)
}

// A TokenOpt is used to change the default behaviour of token handlers.
type TokenOpt struct {
	f func(o *options)
}

// Logger defines the logger to be used whenever detailed messages have to be
// printed out.
func Logger(l handler.Logger) TokenOpt {
	return TokenOpt{func(o *options) {
		o.logger = l
	}}
}

// Expiration sets the expiration time of the auth token
func Expiration(e time.Duration) TokenOpt {
	return TokenOpt{func(o *options) {
		o.expiration = e
	}}
}

// Claimer is responsible for transforming a standard claims object into a
// custom one.
func Claimer(c func(claims *jwt.StandardClaims) jwt.Claims) TokenOpt {
	return TokenOpt{func(o *options) {
		o.claimer = c
	}}
}

// Issuer sets the issuer in the standart claims object.
func Issuer(issuer string) TokenOpt {
	return TokenOpt{func(o *options) {
		o.issuer = issuer
	}}
}

// User sets the query key from which to obtain the user.
func User(user string) TokenOpt {
	return TokenOpt{func(o *options) {
		o.user = user
	}}
}

// Password sets the query key from which to obtain the password.
func Password(password string) TokenOpt {
	return TokenOpt{func(o *options) {
		o.password = password
	}}
}

// Extractor extracts a token from a request
func Extractor(e request.Extractor) TokenOpt {
	return TokenOpt{func(o *options) {
		o.extractor = e
	}}
}

type options struct {
	logger     handler.Logger
	claimer    func(*jwt.StandardClaims) jwt.Claims
	expiration time.Duration
	issuer     string
	user       string
	password   string
	extractor  request.Extractor
}

type headerExtractor struct{}

type tokenKey string

const key tokenKey = "token-key"
const claimsKey tokenKey = "claims-key"

// TokenGenerator returns a handler that will read a username and password from
// a request form, create a jwt token if they are valid, and store the signed
// token in the request context for later consumption.
//
// If handler h is nil, the generated token will be written verbatim in the
// response.
func TokenGenerator(h http.Handler, auth Authenticator, secret []byte, opts ...TokenOpt) http.Handler {
	o := options{
		logger:     handler.NopLogger(),
		claimer:    func(c *jwt.StandardClaims) jwt.Claims { return c },
		expiration: time.Hour * 24 * 15,
		user:       "user",
		password:   "password",
	}

	o.apply(opts)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			o.logger.Print("Invalid request form: ", err)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		user := r.FormValue(o.user)
		if !auth.Authenticate(user, r.FormValue(o.password)) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expiration := time.Now().Add(o.expiration)
		t := jwt.NewWithClaims(jwt.SigningMethodHS512, o.claimer(&jwt.StandardClaims{
			Subject:   user,
			ExpiresAt: expiration.Unix(),
			Issuer:    o.issuer,
		}))

		if token, err := t.SignedString(secret); err == nil {
			if h == nil {
				w.Header().Add("Authorization", "Bearer "+token)
				w.Write([]byte(token))

				return
			}

			r = r.WithContext(context.WithValue(r.Context(), key, token))

			h.ServeHTTP(w, r)
		} else {
			o.logger.Print("Error authenticating user:", err)

			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

// TokenBlacklister returns a handler that will blacklist any token found in
// the request. Typically to be used around logout handlers.
//
// If handler h is nil, upon successful blacklist, 200 OK is returned.
func TokenBlacklister(h http.Handler, store TokenStorage, secret []byte, opts ...TokenOpt) http.Handler {
	o := options{
		logger:     handler.NopLogger(),
		claimer:    func(c *jwt.StandardClaims) jwt.Claims { return c },
		expiration: time.Hour * 24 * 15,
		extractor:  request.MultiExtractor{headerExtractor{}, request.ArgumentExtractor{"token"}},
	}

	o.apply(opts)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := o.claimer(&jwt.StandardClaims{})

		tokenStr, err := o.extractor.ExtractToken(r)
		if err != nil {
			o.logger.Print("Invalid request: ", err)

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return secret, nil
		})

		errCode := http.StatusInternalServerError
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// No reason to store an expired token
				if h == nil {
					w.WriteHeader(http.StatusOK)
				} else {
					h.ServeHTTP(w, r)
				}

				return
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				err = nil
			} else {
				errCode = http.StatusBadRequest
			}
		}

		if err == nil {
			var expiration time.Time
			if v, ok := token.Claims.(*jwt.StandardClaims); ok {
				expiration = time.Unix(v.ExpiresAt, 0)
			} else if v, ok := token.Claims.(jwt.MapClaims); ok {
				if exp, ok := v["exp"].(int64); ok {
					expiration = time.Unix(exp, 0)
				}
			}

			if expiration.IsZero() {
				expiration = time.Now().Add(o.expiration)
			}

			err = store.Store(tokenStr, expiration)
		}

		if err != nil {
			o.logger.Print("Error authenticating user: ", err)

			http.Error(w, err.Error(), errCode)
			return
		}

		if h == nil {
			w.WriteHeader(http.StatusOK)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

func RequireToken(h http.Handler, validator TokenValidator, secret []byte, opts ...TokenOpt) http.Handler {
	o := options{
		logger:     handler.NopLogger(),
		claimer:    func(c *jwt.StandardClaims) jwt.Claims { return c },
		expiration: time.Hour * 24 * 15,
		extractor:  request.MultiExtractor{headerExtractor{}, request.ArgumentExtractor{"token"}},
	}

	o.apply(opts)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowed := false
		claims := o.claimer(&jwt.StandardClaims{})

		token, err := request.ParseFromRequestWithClaims(r, o.extractor, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return secret, nil
		})

		if token != nil && token.Valid && token.Claims.Valid() == nil {
			// Check if the token hasn't been blacklisted, via the custom validator
			if tok, err := o.extractor.ExtractToken(r); err == nil && validator.Validate(tok, token.Claims) {
				r = r.WithContext(context.WithValue(r.Context(), claimsKey, token.Claims))
				allowed = true
			}
		} else if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		if allowed {
			h.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	})
}

// Token returns the token string stored in the request context, or an empty
// string.
func Token(r *http.Request) string {
	if token, ok := r.Context().Value(key).(string); ok {
		return token
	}

	return ""
}

// Claims returns the claims stored in the request
func Claims(r *http.Request) jwt.Claims {
	if claims, ok := r.Context().Value(claimsKey).(jwt.Claims); ok {
		return claims
	}

	return nil
}

func (o *options) apply(opts []TokenOpt) {
	for _, op := range opts {
		op.f(o)
	}
}

func (e headerExtractor) ExtractToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return header[7:], nil
	}

	return "", request.ErrNoTokenInRequest
}
