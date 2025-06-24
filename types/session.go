package types

import "time"

// Session represents a Langfuse session containing traces.
type Session struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	ProjectID string    `json:"projectId"`
	Traces    []Trace   `json:"traces"`
}

// Trace represents a Langfuse trace, which is a single event within a session.
type Trace struct {
	ID         string         `json:"id"`
	ExternalID interface{}    `json:"externalId"`
	Timestamp  time.Time      `json:"timestamp"`
	Name       string         `json:"name"`
	UserID     string         `json:"userId"`
	Metadata   map[string]any `json:"metadata"`
	Release    string         `json:"release"`
	Version    string         `json:"version"`
	ProjectID  string         `json:"projectId"`
	Public     bool           `json:"public"`
	Bookmarked bool           `json:"bookmarked"`
	Tags       []string       `json:"tags"`
	Input      any            `json:"input"`
	Output     any            `json:"output"`
	SessionID  string         `json:"sessionId"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}
