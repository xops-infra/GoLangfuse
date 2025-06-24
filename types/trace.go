package types

import (
	"github.com/google/uuid"
)

// TraceEvent model representing langfuse trace
// Fields:
//   - ID an uuid for trace, auto-generated if not provided
//   - Name trace name
//   - UserID a user id to map traces to individual users
//   - SessionID a session id to map traces to specific session
//   - Release map trace with release
//   - Version map trace with  version
//   - Metadata of the trace
//   - Tags attach tags to the trace
//   - Public trace visibility, public or private, defaults to privet
//   - Input an input to LLM
//   - Output an output from LLM
//   - Environment the environment in which the trace was created, e.g. "production", "staging", etc.
type TraceEvent struct {
	ID          *uuid.UUID     `json:"id"`
	Name        string         `json:"name"`
	UserID      string         `json:"userId,omitempty"`
	SessionID   string         `json:"sessionId,omitempty"`
	Release     string         `json:"release,omitempty"`
	Version     string         `json:"version,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Public      bool           `json:"public" default:"false"`
	Input       any            `json:"input,omitempty"`
	Output      any            `json:"output,omitempty"`
	Environment string         `json:"environment,omitempty"`
}

// GetID return an event ID
func (t *TraceEvent) GetID() *uuid.UUID {
	return t.ID
}

// SetID set event ID
func (t *TraceEvent) SetID(id *uuid.UUID) {
	t.ID = id
}
