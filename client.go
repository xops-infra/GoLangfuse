// Package langfuse provides http client for langfuse APIs
package langfuse

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	"github.com/bdpiprava/GoLangfuse/config"
	"github.com/bdpiprava/GoLangfuse/logger"
	"github.com/bdpiprava/GoLangfuse/types"
)

// Client a client interface for sending events to a Langfuse server
type Client interface {
	// Send sends ingestion event to langfuse using rest API
	Send(ctx context.Context, event types.LangfuseEvent) error
}

type client struct {
	client *http.Client
	config *config.Langfuse
}

// NewOptimizedHTTPClient creates an HTTP client optimized for Langfuse API calls
func NewOptimizedHTTPClient(cfg *config.Langfuse) *http.Client {
	return &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
			DisableKeepAlives:   false,
			DisableCompression:  false,
		},
	}
}

// NewClient initialise new langfuse api client
func NewClient(
	config *config.Langfuse,
	httpClient *http.Client,
) Client {
	return &client{
		client: httpClient,
		config: config,
	}
}

// Send sends ingestion event to langfuse using rest API
func (c client) Send(ctx context.Context, ingestionEvent types.LangfuseEvent) error {
	log := logger.FromContext(ctx)
	if strings.TrimSpace(c.config.URL) == "" {
		log.Warn("langfuse config is not provided. no action is taken")
		return fmt.Errorf("missing langfuse config")
	}

	eventType := getEventType(ingestionEvent)
	if eventType == "unknown" {
		log.Errorf("cannot process event of 'unknown' type")
		return errors.Errorf("cannot process event of 'unknown' type")
	}

	if _, err := govalidator.ValidateStruct(ingestionEvent); err != nil {
		log.WithError(err).Errorf("ingestion event validation failed")
		return errors.Wrapf(err, "ingestion event validation failed")
	}

	request := &ingestionRequest{
		Batch: []event{
			{
				ID:        ingestionEvent.GetID().String(),
				Body:      ingestionEvent,
				Type:      eventType,
				Timestamp: time.Now(),
			},
		},
	}

	resp, err := c.sendEventWithRetry(ctx, request)
	if err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		log.Errorf("request to langfues returned errors in response %v", resp.Errors)
		return errors.Errorf("request to langfues returned errors in response %v", resp.Errors)
	}

	return err
}

// sendEventWithRetry sends an ingestion event to langfuse with retry logic
func (c client) sendEventWithRetry(ctx context.Context, request *ingestionRequest) (*ingestionResponse, error) {
	var lastErr error
	
	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			// Calculate exponential backoff delay
			delay := time.Duration(math.Pow(2, float64(i-1))) * c.config.RetryDelay
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
		
		resp, err := c.sendEvent(ctx, request)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		log := logger.FromContext(ctx)
		
		// Don't retry on client errors (4xx)
		if httpErr, ok := err.(*HTTPError); ok && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
			log.WithError(err).Errorf("client error, not retrying: %v", err)
			break
		}
		
		if i < c.config.MaxRetries {
			log.WithError(err).Warnf("request failed, retrying (attempt %d/%d)", i+1, c.config.MaxRetries)
		}
	}
	
	return nil, fmt.Errorf("failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// sendEvent send and ingestion event to langfuse
func (c client) sendEvent(ctx context.Context, request *ingestionRequest) (*ingestionResponse, error) {
	log := logger.FromContext(ctx)
	apiPath, err := url.JoinPath(c.config.URL, "/api/public/ingestion")
	if err != nil {
		log.WithError(err).Errorf("failed to build langfuse url using %s and /api/public/ingestion", c.config.URL)
		return nil, errors.Wrapf(err, "failed to build langfuse url using %s and /api/public/ingestion", c.config.URL)
	}

	payload, err := json.Marshal(request)
	if err != nil {
		log.WithError(err).Error("failed to marshal request payload")
		return nil, errors.Wrapf(err, "failed to marshal request payload")
	}

	// Compress payload if it's large enough
	var body io.Reader
	var contentEncoding string
	if len(payload) > 1024 { // Compress if payload > 1KB
		var compressedBuf bytes.Buffer
		gzWriter := gzip.NewWriter(&compressedBuf)
		if _, err := gzWriter.Write(payload); err != nil {
			gzWriter.Close()
			return nil, errors.Wrapf(err, "failed to compress payload")
		}
		if err := gzWriter.Close(); err != nil {
			return nil, errors.Wrapf(err, "failed to close gzip writer")
		}
		body = &compressedBuf
		contentEncoding = "gzip"
	} else {
		body = bytes.NewBuffer(payload)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, apiPath, body)
	if err != nil {
		log.WithError(err).Error("failed to create langfuse request")
		return nil, errors.Wrapf(err, "failed to create langfuse request")
	}

	httpRequest.SetBasicAuth(c.config.PublicKey, c.config.SecretKey)
	httpRequest.Header.Set("Content-Type", "application/json")
	httpRequest.Header.Set("Accept-Encoding", "gzip")
	if contentEncoding != "" {
		httpRequest.Header.Set("Content-Encoding", contentEncoding)
	}

	resp, err := c.client.Do(httpRequest)
	if err != nil {
		log.WithError(err).Error("request to langfuse failed")
		return nil, errors.Wrapf(err, "request to langfuse failed")
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// Handle HTTP errors
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	// Handle compressed response
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			log.WithError(err).Error("failed to create gzip reader")
			return nil, errors.Wrapf(err, "failed to create gzip reader")
		}
		defer gzReader.Close()
		reader = gzReader
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil {
		log.WithError(err).Error("failed to read response")
		return nil, errors.Wrapf(err, "failed to read response")
	}

	var response ingestionResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		log.WithError(err).Error("failed to parse response")
		return nil, errors.Wrapf(err, "failed to parse response")
	}
	return &response, nil
}

func getEventType(ingestionEvent types.LangfuseEvent) string {
	switch ingestionEvent.(type) {
	case *types.TraceEvent:
		return "trace-create"
	case *types.GenerationEvent:
		return "generation-create"
	case *types.SpanEvent:
		return "span-create"
	case *types.ScoreEvent:
		return "score-create"
	}
	return "unknown"
}
