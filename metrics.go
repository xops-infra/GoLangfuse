// Package langfuse provides performance monitoring and metrics collection
package langfuse

import (
	"context"
	"sync"
	"time"
)

// Metrics represents performance metrics for the Langfuse client
type Metrics struct {
	mu                    sync.RWMutex
	EventsProcessed       int64         `json:"events_processed"`
	EventsQueued          int64         `json:"events_queued"`
	EventsFailed          int64         `json:"events_failed"`
	BatchesProcessed      int64         `json:"batches_processed"`
	BatchesFailed         int64         `json:"batches_failed"`
	HTTPRequestsTotal     int64         `json:"http_requests_total"`
	HTTPRequestsSuccess   int64         `json:"http_requests_success"`
	HTTPRequestsFailure   int64         `json:"http_requests_failure"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	TotalResponseTime     time.Duration `json:"total_response_time"`
	MaxResponseTime       time.Duration `json:"max_response_time"`
	MinResponseTime       time.Duration `json:"min_response_time"`
	ActiveProcessors      int32         `json:"active_processors"`
	QueueSize             int           `json:"queue_size"`
	QueueCapacity         int           `json:"queue_capacity"`
	StartTime             time.Time     `json:"start_time"`
	LastEventProcessedAt  *time.Time    `json:"last_event_processed_at"`
	LastErrorAt           *time.Time    `json:"last_error_at"`
	LastError             string        `json:"last_error,omitempty"`
}

// HealthStatus represents the health status of the Langfuse service
type HealthStatus struct {
	Status           string        `json:"status"`
	Uptime           time.Duration `json:"uptime"`
	QueueHealth      string        `json:"queue_health"`
	ProcessorHealth  string        `json:"processor_health"`
	APIHealth        string        `json:"api_health"`
	LastHealthCheck  time.Time     `json:"last_health_check"`
	Errors           []string      `json:"errors,omitempty"`
	Warnings         []string      `json:"warnings,omitempty"`
}

// MetricsCollector handles metrics collection and health monitoring
type MetricsCollector struct {
	metrics          *Metrics
	healthStatus     *HealthStatus
	mu               sync.RWMutex
	startTime        time.Time
	responseTimes    []time.Duration
	maxResponseTimes int
}

// NewMetricsCollector creates a new metrics collector
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
		responseTimes:    make([]time.Duration, 0, 100), // Keep last 100 response times
		maxResponseTimes: 100,
	}
}

// IncrementEventsProcessed increments the processed events counter
func (mc *MetricsCollector) IncrementEventsProcessed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.EventsProcessed++
	now := time.Now().UTC()
	mc.metrics.LastEventProcessedAt = &now
}

// IncrementEventsQueued increments the queued events counter
func (mc *MetricsCollector) IncrementEventsQueued() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.EventsQueued++
}

// IncrementEventsFailed increments the failed events counter
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

// IncrementBatchesProcessed increments the processed batches counter
func (mc *MetricsCollector) IncrementBatchesProcessed() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.BatchesProcessed++
}

// IncrementBatchesFailed increments the failed batches counter
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

// RecordHTTPRequest records HTTP request metrics
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

// UpdateQueueMetrics updates queue-related metrics
func (mc *MetricsCollector) UpdateQueueMetrics(size, capacity int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.QueueSize = size
	mc.metrics.QueueCapacity = capacity
}

// UpdateActiveProcessors updates the active processors count
func (mc *MetricsCollector) UpdateActiveProcessors(count int32) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.metrics.ActiveProcessors = count
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return *mc.metrics
}

// CheckHealth performs health checks and returns status
func (mc *MetricsCollector) CheckHealth(ctx context.Context) HealthStatus {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	now := time.Now().UTC()
	health := HealthStatus{
		Status:          "healthy",
		Uptime:          now.Sub(mc.startTime),
		LastHealthCheck: now,
		Errors:          []string{},
		Warnings:        []string{},
	}
	
	// Check queue health
	queueUtilization := float64(mc.metrics.QueueSize) / float64(mc.metrics.QueueCapacity)
	switch {
	case queueUtilization > 0.9:
		health.QueueHealth = "critical"
		health.Errors = append(health.Errors, "Queue utilization critical (>90%)")
		health.Status = "unhealthy"
	case queueUtilization > 0.7:
		health.QueueHealth = "warning"
		health.Warnings = append(health.Warnings, "Queue utilization high (>70%)")
		if health.Status == "healthy" {
			health.Status = "degraded"
		}
	default:
		health.QueueHealth = "healthy"
	}
	
	// Check processor health
	if mc.metrics.ActiveProcessors == 0 {
		health.ProcessorHealth = "critical"
		health.Errors = append(health.Errors, "No active processors")
		health.Status = "unhealthy"
	} else {
		health.ProcessorHealth = "healthy"
	}
	
	// Check API health based on error rates
	if mc.metrics.HTTPRequestsTotal > 0 {
		errorRate := float64(mc.metrics.HTTPRequestsFailure) / float64(mc.metrics.HTTPRequestsTotal)
		switch {
		case errorRate > 0.1: // 10% error rate
			health.APIHealth = "critical"
			health.Errors = append(health.Errors, "High API error rate (>10%)")
			health.Status = "unhealthy"
		case errorRate > 0.05: // 5% error rate
			health.APIHealth = "warning"
			health.Warnings = append(health.Warnings, "Elevated API error rate (>5%)")
			if health.Status == "healthy" {
				health.Status = "degraded"
			}
		default:
			health.APIHealth = "healthy"
		}
	} else {
		health.APIHealth = "unknown"
	}
	
	// Check for recent errors
	if mc.metrics.LastErrorAt != nil && now.Sub(*mc.metrics.LastErrorAt) < 5*time.Minute {
		health.Warnings = append(health.Warnings, "Recent errors detected")
		if health.Status == "healthy" {
			health.Status = "degraded"
		}
	}
	
	mc.healthStatus = &health
	return health
}

// GetHealthStatus returns the last known health status
func (mc *MetricsCollector) GetHealthStatus() HealthStatus {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if mc.healthStatus == nil {
		return HealthStatus{Status: "unknown"}
	}
	return *mc.healthStatus
}

// Reset resets all metrics (useful for testing)
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