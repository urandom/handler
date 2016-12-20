package security

import (
	"net/http"
	"reflect"
	"testing"
)

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
		t.Run(tt.name, func(t *testing.T) {
			if got := NonceValueFromRequest(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NonceValueFromRequest() = %v, want %v", got, tt.want)
			}
		})
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
		t.Run(tt.name, func(t *testing.T) {
			if err := StoreNonce(tt.args.w, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("StoreNonce() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
		{"valid", fields{NonceValid}, true},
		{"invalid", fields{NonceInvalid}, false},
		{"not there", fields{NonceNotRequested}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NonceStatus{
				Status: tt.fields.Status,
			}
			if got := s.Valid(); got != tt.want {
				t.Errorf("NonceStatus.Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		t.Run(tt.name, func(t *testing.T) {
			if got := Nonce(tt.args.h, tt.args.opts...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Nonce() = %v, want %v", got, tt.want)
			}
		})
	}
}
