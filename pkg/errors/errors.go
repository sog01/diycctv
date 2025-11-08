package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error with HTTP status code information
type AppError struct {
	Code       string // Internal error code
	Message    string // Human-readable error message
	StatusCode int    // HTTP status code
	Err        error  // Original error if wrapping
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error if any
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError with default status code 500
func New(code string, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
	}
}

// NewWithStatus creates a new AppError with a specific status code
func NewWithStatus(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code string, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

// WrapWithStatus wraps an existing error with additional context and specific status code
func WrapWithStatus(err error, code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// Common errors with standard HTTP status codes
var (
	// 400 Bad Request
	ErrBadRequest = NewWithStatus("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	
	// 401 Unauthorized
	ErrUnauthorized = NewWithStatus("UNAUTHORIZED", "Unauthorized", http.StatusUnauthorized)
	
	// 403 Forbidden
	ErrForbidden = NewWithStatus("FORBIDDEN", "Forbidden", http.StatusForbidden)
	
	// 404 Not Found
	ErrNotFound = NewWithStatus("NOT_FOUND", "Resource not found", http.StatusNotFound)
	
	// 409 Conflict
	ErrConflict = NewWithStatus("CONFLICT", "Resource already exists", http.StatusConflict)
	
	// 422 Unprocessable Entity
	ErrValidation = NewWithStatus("VALIDATION_ERROR", "Validation error", http.StatusUnprocessableEntity)
	
	// 500 Internal Server Error
	ErrInternal = New("INTERNAL_ERROR", "Internal server error")
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetStatusCode returns the HTTP status code for the error
// If not an AppError, defaults to 500 Internal Server Error
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
