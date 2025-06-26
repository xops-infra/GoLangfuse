// Package langfuse provides structured error types for better error handling and observability integration.
//
// This package defines comprehensive error types that align with Langfuse API responses and provide
// structured error handling capabilities including error categorization, retry logic, and detailed
// context for debugging and monitoring.
//
// Error Categories:
//   - CONFIG: Configuration-related errors (missing keys, invalid URLs)
//   - VALIDATION: Input validation errors (invalid event data, malformed IDs)
//   - NETWORK: Network connectivity and timeout errors
//   - API: HTTP API errors from Langfuse service (4xx/5xx responses)
//   - PROCESSING: Internal processing errors (batch failures, service state)
//
// Usage:
//
//	// Create errors with context
//	err := ErrAPIUnauthorized.WithStatusCode(401).WithDetails(map[string]any{
//	    "endpoint": "/api/traces",
//	    "method": "POST",
//	})
//
//	// Check error properties
//	if langfuseErr, ok := err.(*LangfuseError); ok {
//	    if langfuseErr.IsRetryable() {
//	        // Implement retry logic
//	    }
//	}
//
//	// Wrap existing errors
//	wrappedErr := WrapError(originalErr, ErrNetworkTimeout)
package langfuse

import (
	"fmt"
	"net/http"
)

const (
	httpServerErrorStart = 500 // HTTP server error status codes start
)

// Error types matching Langfuse API responses
var (
	// Configuration errors
	ErrInvalidConfig    = &Error{Code: "INVALID_CONFIG", Message: "invalid langfuse configuration", Type: ErrorTypeConfig}
	ErrMissingURL       = &Error{Code: "MISSING_URL", Message: "langfuse URL is required", Type: ErrorTypeConfig}
	ErrMissingPublicKey = &Error{Code: "MISSING_PUBLIC_KEY", Message: "langfuse public key is required", Type: ErrorTypeConfig}
	ErrMissingSecretKey = &Error{Code: "MISSING_SECRET_KEY", Message: "langfuse secret key is required", Type: ErrorTypeConfig}

	// Validation errors
	ErrEventValidation  = &Error{Code: "EVENT_VALIDATION", Message: "event validation failed", Type: ErrorTypeValidation}
	ErrUnknownEventType = &Error{Code: "UNKNOWN_EVENT_TYPE", Message: "unknown event type", Type: ErrorTypeValidation}
	ErrInvalidEventID   = &Error{Code: "INVALID_EVENT_ID", Message: "invalid event ID", Type: ErrorTypeValidation}

	// Network errors
	ErrNetworkTimeout   = &Error{Code: "NETWORK_TIMEOUT", Message: "network request timed out", Type: ErrorTypeNetwork}
	ErrConnectionFailed = &Error{Code: "CONNECTION_FAILED", Message: "failed to connect to langfuse", Type: ErrorTypeNetwork}
	ErrRequestFailed    = &Error{Code: "REQUEST_FAILED", Message: "HTTP request failed", Type: ErrorTypeNetwork}

	// API errors
	ErrAPIUnauthorized = &Error{Code: "UNAUTHORIZED", Message: "unauthorized access to langfuse API", Type: ErrorTypeAPI}
	ErrAPIForbidden    = &Error{Code: "FORBIDDEN", Message: "forbidden access to langfuse resource", Type: ErrorTypeAPI}
	ErrAPINotFound     = &Error{Code: "NOT_FOUND", Message: "langfuse resource not found", Type: ErrorTypeAPI}
	ErrAPIRateLimit    = &Error{Code: "RATE_LIMIT", Message: "langfuse API rate limit exceeded", Type: ErrorTypeAPI}
	ErrAPIServerError  = &Error{Code: "SERVER_ERROR", Message: "langfuse server error", Type: ErrorTypeAPI}

	// Processing errors
	ErrBatchProcessing = &Error{Code: "BATCH_PROCESSING", Message: "batch processing failed", Type: ErrorTypeProcessing}
	ErrEventProcessing = &Error{Code: "EVENT_PROCESSING", Message: "event processing failed", Type: ErrorTypeProcessing}
	ErrServiceStopped  = &Error{Code: "SERVICE_STOPPED", Message: "langfuse service is stopped", Type: ErrorTypeProcessing}
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeConfig     ErrorType = "CONFIG"
	ErrorTypeValidation ErrorType = "VALIDATION"
	ErrorTypeNetwork    ErrorType = "NETWORK"
	ErrorTypeAPI        ErrorType = "API"
	ErrorTypeProcessing ErrorType = "PROCESSING"
)

// Error represents a structured error with context
type Error struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Type       ErrorType      `json:"type"`
	Details    map[string]any `json:"details,omitempty"`
	Cause      error          `json:"-"`
	StatusCode int            `json:"status_code,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithCause adds a cause to the error
func (e *Error) WithCause(cause error) *Error {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details map[string]any) *Error {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithStatusCode adds HTTP status code to the error
func (e *Error) WithStatusCode(statusCode int) *Error {
	newErr := *e
	newErr.StatusCode = statusCode
	return &newErr
}

// IsRetryable returns whether the error is retryable
func (e *Error) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork:
		return true
	case ErrorTypeAPI:
		// Only retry on server errors and rate limits
		return e.StatusCode >= httpServerErrorStart || e.StatusCode == http.StatusTooManyRequests
	default:
		return false
	}
}

// IsClientError returns whether the error is a client error (4xx)
func (e *Error) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns whether the error is a server error (5xx)
func (e *Error) IsServerError() bool {
	return e.StatusCode >= httpServerErrorStart
}

// NewHTTPError creates a new error from HTTP response
func NewHTTPError(statusCode int, message string) *Error {
	var baseErr *Error

	switch statusCode {
	case http.StatusUnauthorized:
		baseErr = ErrAPIUnauthorized
	case http.StatusForbidden:
		baseErr = ErrAPIForbidden
	case http.StatusNotFound:
		baseErr = ErrAPINotFound
	case http.StatusTooManyRequests:
		baseErr = ErrAPIRateLimit
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		baseErr = ErrAPIServerError
	default:
		if statusCode >= 400 && statusCode < 500 {
			baseErr = &Error{Code: "CLIENT_ERROR", Message: "client error", Type: ErrorTypeAPI}
		} else if statusCode >= httpServerErrorStart {
			baseErr = &Error{Code: "SERVER_ERROR", Message: "server error", Type: ErrorTypeAPI}
		} else {
			baseErr = &Error{Code: "HTTP_ERROR", Message: "HTTP error", Type: ErrorTypeNetwork}
		}
	}

	return baseErr.WithStatusCode(statusCode).WithDetails(map[string]any{
		"response_body": message,
	})
}

// NewValidationError creates a validation error with details
func NewValidationError(field string, value any, reason string) *Error {
	return ErrEventValidation.WithDetails(map[string]any{
		"field":  field,
		"value":  value,
		"reason": reason,
	})
}

// NewConfigError creates a configuration error with details
func NewConfigError(field string, reason string) *Error {
	return ErrInvalidConfig.WithDetails(map[string]any{
		"field":  field,
		"reason": reason,
	})
}

// WrapError wraps an existing error with Langfuse context
func WrapError(err error, baseErr *Error) *Error {
	if langfuseErr, ok := err.(*Error); ok {
		return langfuseErr
	}
	return baseErr.WithCause(err)
}
