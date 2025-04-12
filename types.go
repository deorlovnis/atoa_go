package atoa

import (
	"encoding/json"
	"time"
)

// A2AMessage represents a message sent between agents in a session
type A2AMessage struct {
	SessionID   string          `json:"session_id"`
	FromAgentID string          `json:"from_agent_id"`
	ToAgentID   string          `json:"to_agent_id"`
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Timestamp   time.Time       `json:"timestamp"`
}
