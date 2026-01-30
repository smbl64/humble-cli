package api

import "fmt"

// ErrorType represents the type of API error
type ErrorType int

const (
	NetworkError ErrorType = iota
	DeserializeError
	BundleNotFound
)

// ApiError represents an error from the Humble Bundle API
type ApiError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *ApiError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *ApiError) Unwrap() error {
	return e.Err
}

// NewNetworkError creates a new network error
func NewNetworkError(err error) *ApiError {
	return &ApiError{
		Type:    NetworkError,
		Message: "network error",
		Err:     err,
	}
}

// NewDeserializeError creates a new deserialization error
func NewDeserializeError(err error) *ApiError {
	return &ApiError{
		Type:    DeserializeError,
		Message: "cannot parse the response",
		Err:     err,
	}
}

// NewBundleNotFoundError creates a new bundle not found error
func NewBundleNotFoundError() *ApiError {
	return &ApiError{
		Type:    BundleNotFound,
		Message: "cannot find any data",
	}
}
