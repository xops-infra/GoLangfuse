// Package config provides configuration management for the GoLangfuse client library.
//
// This package handles loading, validation, and management of configuration
// parameters required to connect to and interact with the Langfuse API.
// Configuration can be loaded from environment variables using the envconfig
// library with automatic validation.
//
// Example usage:
//
//	// Load configuration from environment variables
//	cfg, err := config.LoadLangfuseConfig()
//	if err != nil {
//	    log.Fatalf("Failed to load config: %v", err)
//	}
//
//	// Or create configuration manually
//	cfg := &config.Langfuse{
//	    URL:       "https://api.langfuse.com",
//	    PublicKey: "pk_your_public_key",
//	    SecretKey: "sk_your_secret_key",
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatalf("Invalid config: %v", err)
//	}
package config

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/kelseyhightower/envconfig"
)

// Langfuse contains all configuration parameters required to initialize
// and operate the GoLangfuse client.
//
// Configuration values are automatically loaded from environment variables
// when using LoadLangfuseConfig(). Each field has a corresponding environment
// variable with the LANGFUSE_ prefix.
//
// Authentication Configuration:
//   - URL: The Langfuse server endpoint (typically https://api.langfuse.com)
//   - PublicKey: Your project's public key from the Langfuse dashboard
//   - SecretKey: Your project's secret key from the Langfuse dashboard
//
// Performance Configuration:
//   - NumberOfEventProcessor: Number of concurrent goroutines processing events
//   - BatchSize: Maximum number of events to batch together
//   - BatchTimeout: Maximum time to wait before sending a partial batch
//
// HTTP Configuration:
//   - Timeout: HTTP request timeout for API calls
//   - MaxIdleConns: Maximum number of idle HTTP connections
//   - MaxIdleConnsPerHost: Maximum idle connections per host
//   - IdleConnTimeout: How long to keep idle connections open
//
// Reliability Configuration:
//   - MaxRetries: Maximum number of retry attempts for failed requests
//   - RetryDelay: Base delay between retry attempts (uses exponential backoff)
//
// Example environment variables:
//
//	LANGFUSE_URL=https://api.langfuse.com
//	LANGFUSE_PUBLIC_KEY=pk_your_public_key
//	LANGFUSE_SECRET_KEY=sk_your_secret_key
//	LANGFUSE_NUM_OF_EVENT_PROCESSOR=4
//	LANGFUSE_BATCH_SIZE=10
//	LANGFUSE_BATCH_TIMEOUT=5s
//	LANGFUSE_MAX_RETRIES=3
type Langfuse struct {
	// URL is the Langfuse server endpoint.
	// Required. Must be a valid URL (e.g., https://api.langfuse.com).
	// Environment variable: LANGFUSE_URL
	URL string `envconfig:"LANGFUSE_URL" valid:"url,required"`

	// PublicKey is the public key for your Langfuse project.
	// Required. Obtained from your Langfuse project settings.
	// Environment variable: LANGFUSE_PUBLIC_KEY
	PublicKey string `envconfig:"LANGFUSE_PUBLIC_KEY" valid:"required"`

	// SecretKey is the secret key for your Langfuse project.
	// Required. Obtained from your Langfuse project settings.
	// Keep this value secure and never expose it in client-side code.
	// Environment variable: LANGFUSE_SECRET_KEY
	SecretKey string `envconfig:"LANGFUSE_SECRET_KEY" valid:"required"`

	// NumberOfEventProcessor controls the number of concurrent goroutines
	// that process and send events to the Langfuse API.
	// Default: 1. Recommended: 2-8 depending on event volume.
	// Environment variable: LANGFUSE_NUM_OF_EVENT_PROCESSOR
	NumberOfEventProcessor int `envconfig:"LANGFUSE_NUM_OF_EVENT_PROCESSOR" default:"1"`

	// Timeout is the HTTP request timeout for API calls.
	// Default: 30s. Increase for slow network conditions.
	// Environment variable: LANGFUSE_TIMEOUT
	Timeout time.Duration `envconfig:"LANGFUSE_TIMEOUT" default:"30s"`

	// MaxIdleConns controls the maximum number of idle HTTP connections
	// across all hosts. Helps with connection reuse and performance.
	// Default: 100.
	// Environment variable: LANGFUSE_MAX_IDLE_CONNS
	MaxIdleConns int `envconfig:"LANGFUSE_MAX_IDLE_CONNS" default:"100"`

	// MaxIdleConnsPerHost controls the maximum number of idle HTTP connections
	// to keep per-host. Should be tuned based on your usage patterns.
	// Default: 10.
	// Environment variable: LANGFUSE_MAX_IDLE_CONNS_PER_HOST
	MaxIdleConnsPerHost int `envconfig:"LANGFUSE_MAX_IDLE_CONNS_PER_HOST" default:"10"`

	// IdleConnTimeout is the maximum amount of time an idle connection
	// will remain idle before closing itself.
	// Default: 90s.
	// Environment variable: LANGFUSE_IDLE_CONN_TIMEOUT
	IdleConnTimeout time.Duration `envconfig:"LANGFUSE_IDLE_CONN_TIMEOUT" default:"90s"`

	// MaxRetries is the maximum number of retry attempts for failed requests.
	// Uses exponential backoff with jitter between attempts.
	// Default: 3. Set to 0 to disable retries.
	// Environment variable: LANGFUSE_MAX_RETRIES
	MaxRetries int `envconfig:"LANGFUSE_MAX_RETRIES" default:"3"`

	// RetryDelay is the base delay between retry attempts.
	// Actual delay uses exponential backoff: RetryDelay * (2^attempt) + jitter.
	// Default: 1s.
	// Environment variable: LANGFUSE_RETRY_DELAY
	RetryDelay time.Duration `envconfig:"LANGFUSE_RETRY_DELAY" default:"1s"`

	// BatchSize is the maximum number of events to batch together
	// before sending to the API. Larger batches improve throughput
	// but increase memory usage and latency.
	// Default: 10. Recommended range: 5-50.
	// Environment variable: LANGFUSE_BATCH_SIZE
	BatchSize int `envconfig:"LANGFUSE_BATCH_SIZE" default:"10"`

	// BatchTimeout is the maximum time to wait before sending a partial batch.
	// This ensures events are sent even when BatchSize is not reached.
	// Default: 5s. Lower values reduce latency but may decrease throughput.
	// Environment variable: LANGFUSE_BATCH_TIMEOUT
	BatchTimeout time.Duration `envconfig:"LANGFUSE_BATCH_TIMEOUT" default:"5s"`
}

// Validate performs comprehensive validation of the Langfuse configuration.
//
// This method validates all configuration parameters to ensure they meet
// the requirements for successful operation of the GoLangfuse client.
//
// Validation includes:
//   - Required fields are not empty (URL, PublicKey, SecretKey)
//   - URL is a valid HTTP/HTTPS URL
//   - Numeric values are within acceptable ranges
//   - Duration values are positive
//
// Returns an error if any validation fails, with a descriptive message
// indicating which field(s) failed validation.
//
// Example:
//
//	cfg := &config.Langfuse{
//	    URL:       "https://api.langfuse.com",
//	    PublicKey: "pk_your_key",
//	    SecretKey: "sk_your_key",
//	}
//	if err := cfg.Validate(); err != nil {
//	    log.Fatalf("Configuration is invalid: %v", err)
//	}
func (c *Langfuse) Validate() error {
	if _, err := govalidator.ValidateStruct(c); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	if c.NumberOfEventProcessor <= 0 {
		return fmt.Errorf("number of event processors must be greater than 0")
	}

	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be greater than 0")
	}

	return nil
}

// LoadLangfuseConfig loads and validates Langfuse configuration from environment variables.
//
// This function automatically reads configuration values from environment variables
// with the LANGFUSE_ prefix and applies default values for optional parameters.
// After loading, it performs comprehensive validation to ensure all required
// fields are present and values are within acceptable ranges.
//
// Environment variables are processed using the envconfig library, which supports:
//   - Automatic type conversion (string, int, duration, etc.)
//   - Default values for optional fields
//   - Required field validation
//
// Required environment variables:
//   - LANGFUSE_URL: The Langfuse server endpoint
//   - LANGFUSE_PUBLIC_KEY: Your project's public key
//   - LANGFUSE_SECRET_KEY: Your project's secret key
//
// Optional environment variables (with defaults):
//   - LANGFUSE_NUM_OF_EVENT_PROCESSOR: Number of worker goroutines (default: 1)
//   - LANGFUSE_BATCH_SIZE: Events per batch (default: 10)
//   - LANGFUSE_BATCH_TIMEOUT: Max batch wait time (default: 5s)
//   - LANGFUSE_MAX_RETRIES: Retry attempts (default: 3)
//   - LANGFUSE_TIMEOUT: HTTP timeout (default: 30s)
//   - And others...
//
// Returns a validated Langfuse configuration ready for use, or an error
// if loading or validation fails.
//
// Example:
//
//	// Set environment variables first
//	os.Setenv("LANGFUSE_URL", "https://api.langfuse.com")
//	os.Setenv("LANGFUSE_PUBLIC_KEY", "pk_your_key")
//	os.Setenv("LANGFUSE_SECRET_KEY", "sk_your_key")
//
//	// Load configuration
//	cfg, err := config.LoadLangfuseConfig()
//	if err != nil {
//	    log.Fatalf("Failed to load configuration: %v", err)
//	}
//
//	// Configuration is ready to use
//	client := langfuse.New(cfg)
func LoadLangfuseConfig() (*Langfuse, error) {
	cfg := &Langfuse{}
	if err := envconfig.Process("langfuse", cfg); err != nil {
		return nil, fmt.Errorf("failed to load langfuse config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("langfuse config validation failed: %w", err)
	}

	return cfg, nil
}
