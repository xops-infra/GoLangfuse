// Package config provides application config
package config

import (
	"fmt"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/kelseyhightower/envconfig"
)

// Langfuse holds config for langfuse api calls
// URL langfuse server url
// PublicKey public key from langfuse project
// SecretKey secret key from langfuse project
type Langfuse struct {
	URL                    string        `envconfig:"LANGFUSE_URL" valid:"url,required"`
	PublicKey              string        `envconfig:"LANGFUSE_PUBLIC_KEY" valid:"required"`
	SecretKey              string        `envconfig:"LANGFUSE_SECRET_KEY" valid:"required"`
	NumberOfEventProcessor int           `envconfig:"LANGFUSE_NUM_OF_EVENT_PROCESSOR" default:"1"`
	Timeout                time.Duration `envconfig:"LANGFUSE_TIMEOUT" default:"30s"`
	MaxIdleConns           int           `envconfig:"LANGFUSE_MAX_IDLE_CONNS" default:"100"`
	MaxIdleConnsPerHost    int           `envconfig:"LANGFUSE_MAX_IDLE_CONNS_PER_HOST" default:"10"`
	IdleConnTimeout        time.Duration `envconfig:"LANGFUSE_IDLE_CONN_TIMEOUT" default:"90s"`
	MaxRetries             int           `envconfig:"LANGFUSE_MAX_RETRIES" default:"3"`
	RetryDelay             time.Duration `envconfig:"LANGFUSE_RETRY_DELAY" default:"1s"`
	BatchSize              int           `envconfig:"LANGFUSE_BATCH_SIZE" default:"10"`
	BatchTimeout           time.Duration `envconfig:"LANGFUSE_BATCH_TIMEOUT" default:"5s"`
}

// Validate validates the configuration
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

// LoadLangfuseConfig loads the Langfuse configuration from environment variables
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
