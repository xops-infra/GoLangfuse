package langfuse

import (
	"sync"
	"time"
)

// HealthStatusValue represents the health status of the GoLangfuse client.
type HealthStatusValue string

type ComponentHealthValue string

const (
	cmpHealthHealthy  ComponentHealthValue = "healthy"
	cmpHealthCritical ComponentHealthValue = "critical"
	cmpHealthWarning  ComponentHealthValue = "warning"
	cmpHealthUnknown  ComponentHealthValue = "unknown"

	healthStatusHealthy   HealthStatusValue = "healthy"
	healthStatusDegraded  HealthStatusValue = "degraded"
	healthStatusStarting  HealthStatusValue = "starting"
	healthStatusUnhealthy HealthStatusValue = "unhealthy"
	healthStatusUnknown   HealthStatusValue = "unknown"

	// Metrics constants
	maxResponseTimeHistory   = 100  // Keep last 100 response times
	queueUtilizationCritical = 0.9  // 90% queue utilization threshold
	queueUtilizationWarning  = 0.7  // 70% queue utilization threshold
	errorRateCritical        = 0.1  // 10% error rate threshold
	errorRateWarning         = 0.05 // 5% error rate threshold
)

// Metrics contains comprehensive performance and operational metrics for the GoLangfuse client.
//
// This struct provides detailed insights into the client's operation including:
//   - Event processing statistics (processed, queued, failed)
//   - Batch processing metrics (batches processed and failed)
//   - HTTP request performance (total, success, failure counts and response times)
//   - Resource utilization (active processors, queue usage)
//   - Error tracking (last error time and message)
//
// All metrics are thread-safe and can be safely accessed concurrently.
// Metrics are automatically updated by the GoLangfuse client during operation
// and can be retrieved using the GetMetrics() method.
//
// Example usage:
//
//	client := langfuse.New(config)
//
//	// Get current metrics
//	metrics := client.GetMetrics()
//	fmt.Printf("Events processed: %d\n", metrics.EventsProcessed)
//	fmt.Printf("Success rate: %.2f%%\n",
//	    float64(metrics.HTTPRequestsSuccess)/float64(metrics.HTTPRequestsTotal)*100)
//
// JSON serialization is supported for integration with monitoring systems.
type Metrics struct {
	// EventsProcessed is the total number of events successfully processed
	// and sent to the Langfuse API since client initialization.
	EventsProcessed int64 `json:"events_processed"`

	// EventsQueued is the total number of events that have been queued
	// for processing. This includes events currently being processed.
	EventsQueued int64 `json:"events_queued"`

	// EventsFailed is the total number of events that failed to be processed
	// or sent to the API, even after retries.
	EventsFailed int64 `json:"events_failed"`

	// BatchesProcessed is the total number of event batches successfully
	// sent to the Langfuse API.
	BatchesProcessed int64 `json:"batches_processed"`

	// BatchesFailed is the total number of event batches that failed
	// to be sent to the API, even after retries.
	BatchesFailed int64 `json:"batches_failed"`

	// HTTPRequestsTotal is the total number of HTTP requests made to
	// the Langfuse API (includes both successful and failed requests).
	HTTPRequestsTotal int64 `json:"http_requests_total"`

	// HTTPRequestsSuccess is the number of HTTP requests that completed
	// successfully (2xx status codes).
	HTTPRequestsSuccess int64 `json:"http_requests_success"`

	// HTTPRequestsFailure is the number of HTTP requests that failed
	// (non-2xx status codes, network errors, timeouts).
	HTTPRequestsFailure int64 `json:"http_requests_failure"`

	// AverageResponseTime is the rolling average response time for HTTP
	// requests to the Langfuse API (based on last 100 requests).
	AverageResponseTime time.Duration `json:"average_response_time"`

	// TotalResponseTime is the cumulative response time for all HTTP
	// requests made to the API.
	TotalResponseTime time.Duration `json:"total_response_time"`

	// MaxResponseTime is the maximum response time observed for any
	// HTTP request to the API.
	MaxResponseTime time.Duration `json:"max_response_time"`

	// MinResponseTime is the minimum response time observed for any
	// HTTP request to the API.
	MinResponseTime time.Duration `json:"min_response_time"`

	// ActiveProcessors is the current number of active goroutines
	// processing events.
	ActiveProcessors int `json:"active_processors"`

	// QueueSize is the current number of events waiting in the queue
	// to be processed.
	QueueSize int `json:"queue_size"`

	// QueueCapacity is the maximum number of events that can be queued.
	// When this limit is reached, new events may be dropped or block.
	QueueCapacity int `json:"queue_capacity"`

	// StartTime is when the metrics collection began (typically when
	// the client was initialized).
	StartTime time.Time `json:"start_time"`

	// LastEventProcessedAt is the timestamp of when the last event
	// was successfully processed. Nil if no events have been processed.
	LastEventProcessedAt *time.Time `json:"last_event_processed_at"`

	// LastErrorAt is the timestamp of when the last error occurred.
	// Nil if no errors have occurred.
	LastErrorAt *time.Time `json:"last_error_at"`

	// LastError contains the error message from the most recent error.
	// Empty string if no errors have occurred.
	LastError string `json:"last_error,omitempty"`
}

// HealthStatus provides a comprehensive health assessment of the GoLangfuse client.
//
// The health status includes overall status and component-specific health indicators
// to help diagnose issues and monitor the client's operational state.
//
// Health Status Values:
//   - "healthy": All components operating normally
//   - "degraded": Some issues detected but service is functional
//   - "unhealthy": Critical issues detected, service may not be functioning
//   - "starting": Client is initializing
//   - "unknown": Health status cannot be determined
//
// Component Health Values:
//   - "healthy": Component operating normally
//   - "warning": Component has minor issues
//   - "critical": Component has serious issues
//   - "unknown": Component status cannot be determined
//
// Example usage:
//
//	health := client.CheckHealth(ctx)
//	if health.Status != "healthy" {
//	    log.Printf("Client health: %s", health.Status)
//	    for _, err := range health.Errors {
//	        log.Printf("Error: %s", err)
//	    }
//	}
type HealthStatus struct {
	// Status is the overall health status of the client.
	// Possible values: "healthy", "degraded", "unhealthy", "starting", "unknown"
	Status HealthStatusValue `json:"status"`

	// Uptime is how long the client has been running since initialization.
	Uptime time.Duration `json:"uptime"`

	// QueueHealth indicates the health of the event queue.
	// Based on queue utilization and capacity.
	QueueHealth ComponentHealthValue `json:"queue_health"`

	// ProcessorHealth indicates the health of event processors.
	// Based on active processor count and processing activity.
	ProcessorHealth ComponentHealthValue `json:"processor_health"`

	// APIHealth indicates the health of API connectivity.
	// Based on HTTP request success/failure rates.
	APIHealth ComponentHealthValue `json:"api_health"`

	// LastHealthCheck is when this health status was last updated.
	LastHealthCheck time.Time `json:"last_health_check"`

	// Errors contains critical issues that require immediate attention.
	// These issues may cause service disruption.
	Errors []string `json:"errors,omitempty"`

	// Warnings contains non-critical issues that should be monitored.
	// These issues may indicate potential future problems.
	Warnings []string `json:"warnings,omitempty"`
}

// MetricsCollector handles comprehensive metrics collection and health monitoring
// for the GoLangfuse client.
//
// The collector is thread-safe and designed for concurrent access from multiple
// goroutines. It tracks various performance metrics and provides health assessment
// capabilities for monitoring and alerting.
//
// The collector automatically maintains rolling averages for response times and
// provides detailed health checks based on configurable thresholds.
//
// Example usage:
//
//	collector := NewMetricsCollector()
//
//	// Record metrics during operation
//	collector.IncrementEventsProcessed()
//	collector.RecordHTTPRequest(true, 150*time.Millisecond)
//
//	// Get current metrics
//	metrics := collector.GetMetrics()
//
//	// Check health status
//	health := collector.CheckHealth()
type MetricsCollector struct {
	metrics          *Metrics
	healthStatus     *HealthStatus
	mu               sync.RWMutex
	startTime        time.Time
	responseTimes    []time.Duration
	maxResponseTimes int
}

// NewMetricsCollector creates a new MetricsCollector with initialized metrics and health status.
//
// The collector starts in "starting" status and initializes all counters to zero.
// Response time tracking is configured to maintain a rolling window of the last
// 100 response times for calculating averages.
//
// The collector is immediately ready for use and thread-safe for concurrent access.
//
// Returns a fully initialized MetricsCollector ready for metrics collection.
//
// Example:
//
//	collector := NewMetricsCollector()
//	// Collector is ready to use immediately
//	collector.IncrementEventsQueued()
func NewMetricsCollector() *MetricsCollector {
	now := time.Now().UTC()
	return &MetricsCollector{
		metrics: &Metrics{
			StartTime:       now,
			MinResponseTime: time.Hour, // Initialize with a high value
		},
		healthStatus: &HealthStatus{
			Status:          "starting",
			LastHealthCheck: now,
		},
		startTime:        now,
		responseTimes:    make([]time.Duration, 0, maxResponseTimeHistory), // Keep last 100 response times
		maxResponseTimes: maxResponseTimeHistory,
	}
}

// IncrementEventsProcessed increments the processed events counter and updates
// the last processed timestamp.
//
// This method should be called each time an event is successfully processed
// and sent to the Langfuse API. It's thread-safe and can be called concurrently.
//
// The method updates both the EventsProcessed counter and the LastEventProcessedAt
// timestamp to track processing activity.
func (mc *MetricsCollector) IncrementEventsProcessed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.EventsProcessed++
	now := time.Now().UTC()
	mc.metrics.LastEventProcessedAt = &now
}

// IncrementEventsQueued increments the queued events counter.
//
// This method should be called each time an event is added to the processing
// queue. It helps track the total volume of events flowing through the system.
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) IncrementEventsQueued() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.EventsQueued++
}

// IncrementEventsFailed increments the failed events counter and records error details.
//
// This method should be called when an event fails to be processed or sent to the
// API, even after retries. It updates error tracking metrics including the failure
// count, last error timestamp, and error message.
//
// Parameters:
//   - err: The error that caused the event to fail (can be nil)
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) IncrementEventsFailed(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.EventsFailed++
	now := time.Now().UTC()
	mc.metrics.LastErrorAt = &now
	if err != nil {
		mc.metrics.LastError = err.Error()
	}
}

// IncrementBatchesProcessed increments the processed batches counter.
//
// This method should be called each time a batch of events is successfully
// sent to the Langfuse API. It helps track batching efficiency and throughput.
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) IncrementBatchesProcessed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.BatchesProcessed++
}

// IncrementBatchesFailed increments the failed batches counter and records error details.
//
// This method should be called when a batch of events fails to be sent to the API,
// even after retries. It updates error tracking metrics for batch processing failures.
//
// Parameters:
//   - err: The error that caused the batch to fail (can be nil)
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) IncrementBatchesFailed(err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.BatchesFailed++
	now := time.Now().UTC()
	mc.metrics.LastErrorAt = &now
	if err != nil {
		mc.metrics.LastError = err.Error()
	}
}

// RecordHTTPRequest records comprehensive HTTP request metrics including success/failure
// counts and response time statistics.
//
// This method should be called after each HTTP request to the Langfuse API to track
// performance and reliability metrics. It updates request counters, response time
// statistics, and maintains a rolling average of recent response times.
//
// Response Time Tracking:
//   - Updates total, min, max, and average response times
//   - Maintains a rolling window of the last 100 response times for average calculation
//   - Automatically manages the response time history buffer
//
// Parameters:
//   - success: true if the HTTP request completed successfully (2xx status), false otherwise
//   - responseTime: the total time taken for the HTTP request
//
// Thread-safe for concurrent access.
//
// Example:
//
//	start := time.Now()
//	resp, err := http.Get(url)
//	duration := time.Since(start)
//	collector.RecordHTTPRequest(err == nil && resp.StatusCode < 300, duration)
func (mc *MetricsCollector) RecordHTTPRequest(success bool, responseTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.HTTPRequestsTotal++
	if success {
		mc.metrics.HTTPRequestsSuccess++
	} else {
		mc.metrics.HTTPRequestsFailure++
	}

	// Record response time
	mc.metrics.TotalResponseTime += responseTime

	if responseTime > mc.metrics.MaxResponseTime {
		mc.metrics.MaxResponseTime = responseTime
	}

	if responseTime < mc.metrics.MinResponseTime || mc.metrics.MinResponseTime == 0 {
		mc.metrics.MinResponseTime = responseTime
	}

	// Keep track of recent response times for average calculation
	mc.responseTimes = append(mc.responseTimes, responseTime)
	if len(mc.responseTimes) > mc.maxResponseTimes {
		mc.responseTimes = mc.responseTimes[1:]
	}

	// Calculate average response time
	if len(mc.responseTimes) > 0 {
		var total time.Duration
		for _, rt := range mc.responseTimes {
			total += rt
		}
		mc.metrics.AverageResponseTime = total / time.Duration(len(mc.responseTimes))
	}
}

// UpdateQueueMetrics updates queue-related metrics with current size and capacity.
//
// This method should be called periodically to track queue utilization, which is
// important for performance monitoring and health assessment. High queue utilization
// may indicate processing bottlenecks or insufficient processor capacity.
//
// Parameters:
//   - size: current number of events in the queue waiting to be processed
//   - capacity: maximum number of events the queue can hold
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) UpdateQueueMetrics(size, capacity int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.QueueSize = size
	mc.metrics.QueueCapacity = capacity
}

// UpdateActiveProcessors updates the count of active event processor goroutines.
//
// This method should be called when processor goroutines are started or stopped
// to maintain an accurate count of active processors. This information is used
// for health monitoring and capacity planning.
//
// Parameters:
//   - count: current number of active processor goroutines
//
// Thread-safe for concurrent access.
func (mc *MetricsCollector) UpdateActiveProcessors(count int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.ActiveProcessors = count
}

// GetMetrics returns a copy of the current metrics snapshot.
//
// This method provides a consistent view of all metrics at the time of the call.
// The returned Metrics struct is a copy, so it won't be affected by concurrent
// updates to the collector.
//
// Returns a complete Metrics struct containing all current metric values.
//
// Thread-safe for concurrent access (uses read lock for better performance).
//
// Example:
//
//	metrics := collector.GetMetrics()
//	fmt.Printf("Success rate: %.2f%%\n",
//	    float64(metrics.HTTPRequestsSuccess)/float64(metrics.HTTPRequestsTotal)*100)
//	fmt.Printf("Average response time: %v\n", metrics.AverageResponseTime)
func (mc *MetricsCollector) GetMetrics() Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return *mc.metrics
}

// CheckHealth performs comprehensive health assessment and returns detailed health status.
//
// This method evaluates the health of various components and subsystems:
//   - Queue Health: Based on queue utilization (>90% critical, >70% warning)
//   - Processor Health: Based on active processor count (0 is critical)
//   - API Health: Based on HTTP request error rates (>10% critical, >5% warning)
//   - Recent Errors: Warnings for errors within the last 5 minutes
//
// Health Status Levels:
//   - "healthy": All components operating normally
//   - "degraded": Minor issues detected, service functional
//   - "unhealthy": Critical issues detected, service may not function properly
//
// The health check is performed synchronously and does not support cancellation.
//
// Returns a complete HealthStatus with overall status, component health,
// uptime, and lists of errors and warnings.
//
// Thread-safe for concurrent access.
//
// Example:
//
//	health := collector.CheckHealth()
//	if health.Status != "healthy" {
//	    log.Printf("Health issues detected: %s", health.Status)
//	    for _, err := range health.Errors {
//	        log.Printf("ERROR: %s", err)
//	    }
//	    for _, warn := range health.Warnings {
//	        log.Printf("WARNING: %s", warn)
//	    }
//	}
func (mc *MetricsCollector) CheckHealth() HealthStatus {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now().UTC()
	health := HealthStatus{
		Status:          healthStatusHealthy,
		Uptime:          now.Sub(mc.startTime),
		LastHealthCheck: now,
		Errors:          []string{},
		Warnings:        []string{},
	}

	// Check queue health
	queueUtilization := float64(mc.metrics.QueueSize) / float64(mc.metrics.QueueCapacity)

	switch {
	case queueUtilization > queueUtilizationCritical:
		health.QueueHealth = cmpHealthCritical
		health.Errors = append(health.Errors, "Queue utilization critical (>90%)")
		health.Status = healthStatusUnhealthy
	case queueUtilization > queueUtilizationWarning:
		health.QueueHealth = cmpHealthWarning
		health.Warnings = append(health.Warnings, "Queue utilization high (>70%)")
		if health.Status == healthStatusHealthy {
			health.Status = healthStatusDegraded
		}
	default:
		health.QueueHealth = cmpHealthHealthy
	}

	// Check processor health
	if mc.metrics.ActiveProcessors == 0 {
		health.ProcessorHealth = cmpHealthCritical
		health.Errors = append(health.Errors, "No active processors")
		health.Status = healthStatusUnhealthy
	} else {
		health.ProcessorHealth = cmpHealthHealthy
	}

	// Check API health based on error rates
	if mc.metrics.HTTPRequestsTotal > 0 {
		errorRate := float64(mc.metrics.HTTPRequestsFailure) / float64(mc.metrics.HTTPRequestsTotal)
		switch {
		case errorRate > errorRateCritical: // 10% error rate
			health.APIHealth = cmpHealthCritical
			health.Errors = append(health.Errors, "High API error rate (>10%)")
			health.Status = healthStatusUnhealthy
		case errorRate > errorRateWarning: // 5% error rate
			health.APIHealth = cmpHealthWarning
			health.Warnings = append(health.Warnings, "Elevated API error rate (>5%)")
			if health.Status == healthStatusHealthy {
				health.Status = healthStatusDegraded
			}
		default:
			health.APIHealth = cmpHealthHealthy
		}
	} else {
		health.APIHealth = cmpHealthUnknown
	}

	// Check for recent errors
	if mc.metrics.LastErrorAt != nil && now.Sub(*mc.metrics.LastErrorAt) < 5*time.Minute {
		health.Warnings = append(health.Warnings, "Recent errors detected")
		if health.Status == healthStatusHealthy {
			health.Status = healthStatusDegraded
		}
	}

	mc.healthStatus = &health
	return health
}

// GetHealthStatus returns the most recently computed health status.
//
// This method returns the cached health status from the last call to CheckHealth().
// If CheckHealth() has never been called, it returns a status of "unknown".
//
// This method is useful for retrieving health status without performing the
// computational overhead of a full health assessment. For the most current
// health status, use CheckHealth() instead.
//
// Returns the last known HealthStatus, or a status of "unknown" if no health
// check has been performed yet.
//
// Thread-safe for concurrent access (uses read lock for better performance).
//
// Example:
//
//	// Get cached health status (fast)
//	lastHealth := collector.GetHealthStatus()
//
//	// Or get current health status (more accurate but slower)
//	currentHealth := collector.CheckHealth()
func (mc *MetricsCollector) GetHealthStatus() HealthStatus {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	if mc.healthStatus == nil {
		return HealthStatus{Status: healthStatusUnknown}
	}
	return *mc.healthStatus
}

// Reset resets all metrics and health status to initial values.
//
// This method is primarily intended for testing purposes, allowing tests to
// start with a clean metrics state. In production, metrics should typically
// accumulate over the lifetime of the client.
//
// The reset operation:
//   - Clears all counters and statistics
//   - Resets start time to current time
//   - Sets health status back to "starting"
//   - Clears response time history
//   - Reinitializes the metrics collector to its initial state
//
// Thread-safe for concurrent access.
//
// Example:
//
//	// Reset metrics for a new test
//	collector.Reset()
//	// Collector is now in initial state
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now().UTC()
	mc.metrics = &Metrics{
		StartTime:       now,
		MinResponseTime: time.Hour,
	}
	mc.healthStatus = &HealthStatus{
		Status:          "starting",
		LastHealthCheck: now,
	}
	mc.startTime = now
	mc.responseTimes = make([]time.Duration, 0, mc.maxResponseTimes)
}
