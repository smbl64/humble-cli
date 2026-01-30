package api

import (
	"errors"
	"testing"
)

func TestNetworkError(t *testing.T) {
	baseErr := errors.New("connection timeout")
	err := NewNetworkError(baseErr)

	if err.Type != NetworkError {
		t.Errorf("NewNetworkError() Type = %v, want %v", err.Type, NetworkError)
	}

	if err.Message != "network error" {
		t.Errorf("NewNetworkError() Message = %q, want %q", err.Message, "network error")
	}

	if err.Err != baseErr {
		t.Errorf("NewNetworkError() Err = %v, want %v", err.Err, baseErr)
	}
}

func TestDeserializeError(t *testing.T) {
	baseErr := errors.New("invalid JSON")
	err := NewDeserializeError(baseErr)

	if err.Type != DeserializeError {
		t.Errorf("NewDeserializeError() Type = %v, want %v", err.Type, DeserializeError)
	}

	if err.Message != "cannot parse the response" {
		t.Errorf("NewDeserializeError() Message = %q, want %q", err.Message, "cannot parse the response")
	}

	if err.Err != baseErr {
		t.Errorf("NewDeserializeError() Err = %v, want %v", err.Err, baseErr)
	}
}

func TestBundleNotFoundError(t *testing.T) {
	err := NewBundleNotFoundError()

	if err.Type != BundleNotFound {
		t.Errorf("NewBundleNotFoundError() Type = %v, want %v", err.Type, BundleNotFound)
	}

	if err.Message != "cannot find any data" {
		t.Errorf("NewBundleNotFoundError() Message = %q, want %q", err.Message, "cannot find any data")
	}

	if err.Err != nil {
		t.Errorf("NewBundleNotFoundError() Err = %v, want nil", err.Err)
	}
}

func TestApiError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *ApiError
		want    string
	}{
		{
			name: "error with wrapped error",
			err: &ApiError{
				Type:    NetworkError,
				Message: "network error",
				Err:     errors.New("connection refused"),
			},
			want: "network error: connection refused",
		},
		{
			name: "error without wrapped error",
			err: &ApiError{
				Type:    BundleNotFound,
				Message: "cannot find any data",
				Err:     nil,
			},
			want: "cannot find any data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("ApiError.Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestApiError_Unwrap(t *testing.T) {
	tests := []struct {
		name    string
		err     *ApiError
		wantNil bool
	}{
		{
			name: "error with wrapped error",
			err: &ApiError{
				Type:    NetworkError,
				Message: "network error",
				Err:     errors.New("base error"),
			},
			wantNil: false,
		},
		{
			name: "error without wrapped error",
			err: &ApiError{
				Type:    BundleNotFound,
				Message: "cannot find any data",
				Err:     nil,
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Unwrap()
			if (got == nil) != tt.wantNil {
				t.Errorf("ApiError.Unwrap() = %v, wantNil %v", got, tt.wantNil)
			}

			if !tt.wantNil && got != tt.err.Err {
				t.Errorf("ApiError.Unwrap() = %v, want %v", got, tt.err.Err)
			}
		})
	}
}

func TestApiError_Unwrap_ChainSupport(t *testing.T) {
	// Test that errors.Is and errors.As work with ApiError
	baseErr := errors.New("base error")
	apiErr := NewNetworkError(baseErr)

	if !errors.Is(apiErr, baseErr) {
		t.Errorf("errors.Is(apiErr, baseErr) = false, want true")
	}

	var targetErr *ApiError
	if !errors.As(apiErr, &targetErr) {
		t.Errorf("errors.As(apiErr, &targetErr) = false, want true")
	}

	if targetErr.Type != NetworkError {
		t.Errorf("errors.As() Type = %v, want %v", targetErr.Type, NetworkError)
	}
}
