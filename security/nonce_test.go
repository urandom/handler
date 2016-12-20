package security

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNonce(t *testing.T) {
	type args struct {
		h    http.Handler
		opts []Option
	}
	tests := []struct {
		name string
		args args
		want http.Handler
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := Nonce(tt.args.h, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. Nonce() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestNonceValueFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want NonceStatus
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if got := NonceValueFromRequest(tt.args.r); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. NonceValueFromRequest() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestStoreNonce(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		if err := StoreNonce(tt.args.w, tt.args.r); (err != nil) != tt.wantErr {
			t.Errorf("%q. StoreNonce() error = %v, wantErr %v", tt.name, err, tt.wantErr)
		}
	}
}

func TestNonceStatus_Valid(t *testing.T) {
	type fields struct {
		Status nonceStatus
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		s := NonceStatus{
			Status: tt.fields.Status,
		}
		if got := s.Valid(); got != tt.want {
			t.Errorf("%q. NonceStatus.Valid() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
