package atoa

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSendMessage(t *testing.T) {
	// Initialize test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/messages" {
			t.Errorf("Expected path /messages, got %s", r.URL.Path)
		}

		// Check authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer valid-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse request body
		var msg A2AMessage
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid message format", http.StatusBadRequest)
			return
		}

		// Validate message fields
		if msg.SessionID == "" || msg.FromAgentID == "" || msg.ToAgentID == "" || msg.Type == "" || msg.Payload == nil {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test cases
	tests := []struct {
		name          string
		client        *AgentClient
		message       A2AMessage
		expectedError bool
	}{
		{
			name: "verified agent sends valid message",
			client: &AgentClient{
				BaseURL: server.URL,
				Token:   "valid-token",
				HTTP:    &http.Client{},
			},
			message: A2AMessage{
				SessionID:   "session-123",
				FromAgentID: "agent-1",
				ToAgentID:   "agent-2",
				Type:        "text",
				Payload:     json.RawMessage(`{"content": "Hello"}`),
				Timestamp:   time.Now(),
			},
			expectedError: false,
		},
		{
			name: "unverified agent sends message",
			client: &AgentClient{
				BaseURL: server.URL,
				Token:   "invalid-token",
				HTTP:    &http.Client{},
			},
			message: A2AMessage{
				SessionID:   "session-123",
				FromAgentID: "agent-1",
				ToAgentID:   "agent-2",
				Type:        "text",
				Payload:     json.RawMessage(`{"content": "Hello"}`),
				Timestamp:   time.Now(),
			},
			expectedError: true,
		},
		{
			name: "message with missing fields",
			client: &AgentClient{
				BaseURL: server.URL,
				Token:   "valid-token",
				HTTP:    &http.Client{},
			},
			message: A2AMessage{
				SessionID: "session-123",
				// Missing required fields
			},
			expectedError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.SendMessage(context.Background(), tt.message)
			if tt.expectedError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got %v", err)
			}
		})
	}
}
