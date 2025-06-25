// Package langfuse provides structured error types for better error handling
package langfuse

import (
	"fmt"
	"net/http"
)

// Error types matching Langfuse API responses
var (
	// Configuration errors
	ErrInvalidConfig     = &LangfuseError{Code: "INVALID_CONFIG", Message: "invalid langfuse configuration", Type: ErrorTypeConfig}
	ErrMissingURL        = &LangfuseError{Code: "MISSING_URL", Message: "langfuse URL is required", Type: ErrorTypeConfig}
	ErrMissingPublicKey  = &LangfuseError{Code: "MISSING_PUBLIC_KEY", Message: "langfuse public key is required", Type: ErrorTypeConfig}
	ErrMissingSecretKey  = &LangfuseError{Code: "MISSING_SECRET_KEY", Message: "langfuse secret key is required", Type: ErrorTypeConfig}

	// Validation errors
	ErrEventValidation   = &LangfuseError{Code: "EVENT_VALIDATION", Message: "event validation failed", Type: ErrorTypeValidation}
	ErrUnknownEventType  = &LangfuseError{Code: "UNKNOWN_EVENT_TYPE", Message: "unknown event type", Type: ErrorTypeValidation}
	ErrInvalidEventID    = &LangfuseError{Code: "INVALID_EVENT_ID", Message: "invalid event ID", Type: ErrorTypeValidation}

	// Network errors
	ErrNetworkTimeout    = &LangfuseError{Code: "NETWORK_TIMEOUT", Message: "network request timed out", Type: ErrorTypeNetwork}
	ErrConnectionFailed  = &LangfuseError{Code: "CONNECTION_FAILED", Message: "failed to connect to langfuse", Type: ErrorTypeNetwork}
	ErrRequestFailed     = &LangfuseError{Code: "REQUEST_FAILED", Message: "HTTP request failed", Type: ErrorTypeNetwork}

	// API errors
	ErrAPIUnauthorized   = &LangfuseError{Code: "UNAUTHORIZED", Message: "unauthorized access to langfuse API", Type: ErrorTypeAPI}
	ErrAPIForbidden      = &LangfuseError{Code: "FORBIDDEN", Message: "forbidden access to langfuse resource", Type: ErrorTypeAPI}
	ErrAPINotFound       = &LangfuseError{Code: "NOT_FOUND", Message: "langfuse resource not found", Type: ErrorTypeAPI}
	ErrAPIRateLimit      = &LangfuseError{Code: "RATE_LIMIT", Message: "langfuse API rate limit exceeded", Type: ErrorTypeAPI}
	ErrAPIServerError    = &LangfuseError{Code: "SERVER_ERROR", Message: "langfuse server error", Type: ErrorTypeAPI}

	// Processing errors
	ErrBatchProcessing   = &LangfuseError{Code: "BATCH_PROCESSING", Message: "batch processing failed", Type: ErrorTypeProcessing}
	ErrEventProcessing   = &LangfuseError{Code: "EVENT_PROCESSING", Message: "event processing failed", Type: ErrorTypeProcessing}
	ErrServiceStopped    = &LangfuseError{Code: "SERVICE_STOPPED", Message: "langfuse service is stopped", Type: ErrorTypeProcessing}
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

// LangfuseError represents a structured error with context
type LangfuseError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Type       ErrorType         `json:"type"`
	Details    map[string]any    `json:"details,omitempty"`
	Cause      error             `json:"-"`
	StatusCode int               `json:"status_code,omitempty"`
}

// Error implements the error interface
func (e *LangfuseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithCause adds a cause to the error
func (e *LangfuseError) WithCause(cause error) *LangfuseError {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// WithDetails adds details to the error
func (e *LangfuseError) WithDetails(details map[string]any) *LangfuseError {
	newErr := *e
	newErr.Details = details
	return &newErr
}

// WithStatusCode adds HTTP status code to the error
func (e *LangfuseError) WithStatusCode(statusCode int) *LangfuseError {
	newErr := *e
	newErr.StatusCode = statusCode
	return &newErr
}

// IsRetryable returns whether the error is retryable
func (e *LangfuseError) IsRetryable() bool {
	switch e.Type {
	case ErrorTypeNetwork:
		return true
	case ErrorTypeAPI:
		// Only retry on server errors and rate limits
		return e.StatusCode >= 500 || e.StatusCode == http.StatusTooManyRequests
	default:
		return false
	}
}

// IsClientError returns whether the error is a client error (4xx)
func (e *LangfuseError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns whether the error is a server error (5xx)
func (e *LangfuseError) IsServerError() bool {
	return e.StatusCode >= 500
}

// NewHTTPError creates a new error from HTTP response
func NewHTTPError(statusCode int, message string) *LangfuseError {
	var baseErr *LangfuseError
	
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
			baseErr = &LangfuseError{Code: "CLIENT_ERROR", Message: "client error", Type: ErrorTypeAPI}
		} else if statusCode >= 500 {
			baseErr = &LangfuseError{Code: "SERVER_ERROR", Message: "server error", Type: ErrorTypeAPI}
		} else {
			baseErr = &LangfuseError{Code: "HTTP_ERROR", Message: "HTTP error", Type: ErrorTypeNetwork}
		}
	}
	
	return baseErr.WithStatusCode(statusCode).WithDetails(map[string]any{
		"response_body": message,
	})
}

// NewValidationError creates a validation error with details
func NewValidationError(field string, value any, reason string) *LangfuseError {
	return ErrEventValidation.WithDetails(map[string]any{
		"field":  field,
		"value":  value,
		"reason": reason,
	})
}

// NewConfigError creates a configuration error with details
func NewConfigError(field string, reason string) *LangfuseError {
	return ErrInvalidConfig.WithDetails(map[string]any{
		"field":  field,
		"reason": reason,
	})
}

// WrapError wraps an existing error with Langfuse context
func WrapError(err error, baseErr *LangfuseError) *LangfuseError {
	if langfuseErr, ok := err.(*LangfuseError); ok {
		return langfuseErr
	}
	return baseErr.WithCause(err)
}