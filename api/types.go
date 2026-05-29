package api

import "time"

// APIResponse is the standard HTTP response envelope.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// WSEvent is a WebSocket push event envelope.
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocket event type constants.
const (
	EventLog      = "log"
	EventStatus   = "status"
	EventReadings = "readings"
)

// LogEntry is a structured log message for the UI console.
type LogEntry struct {
	Level   string `json:"level"` // info | warn | error
	Message string `json:"message"`
	Time    string `json:"time"`
}

// ReadingSnapshot is a single metric value pushed to the UI.
type ReadingSnapshot struct {
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
