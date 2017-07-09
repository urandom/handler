package auth_test

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/urandom/handler/auth"
)

var (
	secret     = []byte{'1', '2', '3', '4'}
	expiration = time.Hour * 24 * 15
)

func TestTokenGenerator(t *testing.T) {
	type args struct {
		auth auth.Authenticator
		opts []auth.TokenOpt
	}
	tests := []struct {
		name       string
		args       args
		ukey       string
		pkey       string
		code       int
		expiration time.Duration
	}{
		{"defaults", args{au("u1", "p1"), []auth.TokenOpt{}}, "", "", http.StatusOK, expiration},
		{"wrong user", args{au("u2", "p1"), []auth.TokenOpt{}}, "", "", http.StatusUnauthorized, expiration},
		{"wrong pass", args{au("u1", "p2"), []auth.TokenOpt{}}, "", "", http.StatusUnauthorized, expiration},
		{"user key", args{au("u1", "p1"), []auth.TokenOpt{auth.User("asd")}}, "asd", "", http.StatusOK, expiration},
		{"pass key", args{au("u1", "p1"), []auth.TokenOpt{auth.Password("asd")}}, "", "asd", http.StatusOK, expiration},
		{"expiration", args{au("u1", "p1"), []auth.TokenOpt{auth.Expiration(time.Minute)}}, "", "", http.StatusOK, time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := auth.TokenGenerator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(auth.Token(r)))
			}), tt.args.auth, secret, tt.args.opts...)

			userKey := "user"
			passKey := "password"

			if tt.ukey != "" {
				userKey = tt.ukey
			}

			if tt.pkey != "" {
				passKey = tt.pkey
			}

			r, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/?%s=u1&%s=p1", userKey, passKey), nil)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, r)

			if rec.Code != tt.code {
				t.Fatalf("expected %d, got %d", tt.code, rec.Code)
			}

			if tt.code == http.StatusOK {
				claims := &jwt.StandardClaims{}
				token, err := jwt.ParseWithClaims(rec.Body.String(), claimer(claims), claimsfunc)
				if err != nil {
					t.Fatalf("error parsing token %s: %v", rec.Body.String(), err)
				}

				if c, ok := token.Claims.(*jwt.StandardClaims); ok {
					if c.Subject != "u1" {
						t.Fatalf("expected u1, got %s", c.Subject)
					}

					exp := time.Now().Add(tt.expiration).Unix()
					if math.Abs(float64(exp-c.ExpiresAt)) > 10 {
						t.Fatalf("expected %d, got %d", exp, c.ExpiresAt)
					}
				} else {
					t.Fatalf("expected standard claims, got %T", token.Claims)
				}
			} else {
				if rec.Body.String() != "" {
					t.Fatalf("expected empty string, got %s", rec.Body.String())
				}
			}
		})
	}
}

func TestTokenBlacklister(t *testing.T) {
	type args struct {
		opts []auth.TokenOpt
	}
	tests := []struct {
		name       string
		args       args
		expiration time.Time
		code       int
	}{
		{"defaults", args{[]auth.TokenOpt{}}, time.Now(), http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := token(t, "user", tt.expiration)
			h := auth.TokenBlacklister(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.name))
			}), storage(token, tt.expiration), secret, tt.args.opts...)

			r, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/"), nil)
			r.Header.Add("Authorization", "Bearer "+token)
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, r)

			if rec.Code != tt.code {
				t.Fatalf("expected %d, got %d with body %s", tt.code, rec.Code, rec.Body)
			}

			if rec.Code == http.StatusOK {
				if rec.Body.String() != tt.name {
					t.Fatalf("expected %s, got %s", tt.name, rec.Body)
				}
			}
		})
	}
}

func TestRequireToken(t *testing.T) {
	type args struct {
		opts []auth.TokenOpt
	}
	tests := []struct {
		name      string
		args      args
		useToken  bool
		blacklist bool
		user      string
		code      int
	}{
		{"defaults", args{[]auth.TokenOpt{}}, true, false, "user", http.StatusOK},
		{"defaults", args{[]auth.TokenOpt{}}, false, false, "", http.StatusUnauthorized},
		{"defaults", args{[]auth.TokenOpt{}}, true, false, "removed", http.StatusUnauthorized},
		{"defaults", args{[]auth.TokenOpt{}}, true, true, "user", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := token(t, tt.user, time.Now().Add(time.Hour))
			blacklist := ""
			if tt.blacklist {
				blacklist = token
			}

			h := auth.RequireToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.name))
			}), validator(blacklist), secret, tt.args.opts...)

			r, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/"), nil)
			if tt.useToken {
				r.Header.Add("Authorization", "Bearer "+token)
			}
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, r)

			if rec.Code != tt.code {
				t.Fatalf("expected %d, got %d with body %s", tt.code, rec.Code, rec.Body)
			}

			if rec.Code == http.StatusOK {
				if rec.Body.String() != tt.name {
					t.Fatalf("expected %s, got %s", tt.name, rec.Body)
				}
			}
		})
	}
}

func au(u1, p1 string) auth.Authenticator {
	return auth.AuthenticatorFunc(func(user, password string) bool {
		return user == u1 && password == p1
	})
}

func claimsfunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return secret, nil
}

func claimer(c *jwt.StandardClaims) jwt.Claims {
	return c
}

func storage(t string, e time.Time) auth.TokenStorage {
	return auth.TokenStorageFunc(func(token string, expiration time.Time) error {
		if t != token {
			return errors.Errorf("token %s doesn't match %s", token, t)
		}
		if expiration.Sub(e) > time.Second {
			return errors.Errorf("expiration %s doesn't match %s with %d", expiration, e, expiration.Sub(e))
		}
		return nil
	})
}

func token(t *testing.T, user string, expiration time.Time) string {
	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, &jwt.StandardClaims{
		Subject:   user,
		ExpiresAt: expiration.Unix(),
		Issuer:    "issuer",
	})
	if token, err := tok.SignedString(secret); err == nil {
		return token
	} else {
		t.Fatal(err)
		return ""
	}
}

func validator(blacklisted string) auth.TokenValidator {
	return auth.TokenValidatorFunc(func(token string, claims jwt.Claims) bool {
		if std, ok := claims.(*jwt.StandardClaims); ok {
			if std.Subject == "removed" {
				return false
			}

			return token != blacklisted
		}

		return false
	})
}
