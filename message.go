package atoa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Validate checks if all required fields are present in the message
func (m *A2AMessage) Validate() error {
	if m.SessionID == "" {
		return fmt.Errorf("session_id is required")
	}
	if m.FromAgentID == "" {
		return fmt.Errorf("from_agent_id is required")
	}
	if m.ToAgentID == "" {
		return fmt.Errorf("to_agent_id is required")
	}
	if m.Type == "" {
		return fmt.Errorf("type is required")
	}
	if m.Payload == nil {
		return fmt.Errorf("payload is required")
	}
	if m.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	return nil
}

// SendMessage sends an A2A message to a session
func (c *AgentClient) SendMessage(ctx context.Context, msg A2AMessage) error {
	// Validate message fields
	if err := msg.Validate(); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/messages", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Marshal message to JSON
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))

	// Send request
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
