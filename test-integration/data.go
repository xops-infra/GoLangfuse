// Package integration_test provides utilities for creating test data structures
package integration_test

import (
	"time"

	"github.com/google/uuid"

	"github.com/bdpiprava/GoLangfuse/types"
)

// NewTestTraceEvent returns a sample TraceEvent for testing purposes
func NewTestTraceEvent() *types.TraceEvent {
	sessionID := uuid.New().String()
	userID := uuid.New().String()
	traceID := uuid.New()
	return &types.TraceEvent{
		ID:        &traceID,
		Name:      "SendTrace",
		SessionID: sessionID,
		UserID:    userID,
		Input:     "Test input for SendTrace",
		Output:    "Test output for SendTrace",
		Metadata: map[string]any{
			"key": "value",
		},
		Tags:   []string{"test", "integration"},
		Public: true,
	}
}

// BuildSession constructs a Session from a slice of TraceEvents.
func BuildSession(traceEvents ...*types.TraceEvent) *types.Session {
	if len(traceEvents) == 0 {
		return nil
	}

	sessionID := traceEvents[0].SessionID
	traces := make([]types.Trace, len(traceEvents))
	for i, traceEvent := range traceEvents {
		traces[i] = types.Trace{
			ID:        traceEvent.ID.String(),
			Name:      traceEvent.Name,
			UserID:    traceEvent.UserID,
			Metadata:  traceEvent.Metadata,
			ProjectID: "test-project",
			Public:    traceEvent.Public,
			Tags:      traceEvent.Tags,
			Input:     traceEvent.Input,
			Output:    traceEvent.Output,
			SessionID: sessionID,
		}
	}

	return &types.Session{
		ID:        sessionID,
		CreatedAt: time.Now(),
		ProjectID: "test-project",
		Traces:    traces,
	}
}
