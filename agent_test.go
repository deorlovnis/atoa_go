package atoa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAgentCard_Validate(t *testing.T) {
	tests := []struct {
		name    string
		card    *AgentCard
		wantErr bool
	}{
		{
			name: "valid card",
			card: &AgentCard{
				AgentID:      "test-agent",
				OrgID:        "test-org",
				Capabilities: []string{"text"},
				Endpoints:    []string{"https://test.com"},
			},
			wantErr: false,
		},
		{
			name: "missing agent_id",
			card: &AgentCard{
				OrgID:        "test-org",
				Capabilities: []string{"text"},
			},
			wantErr: true,
		},
		{
			name: "missing org_id",
			card: &AgentCard{
				AgentID:      "test-agent",
				Capabilities: []string{"text"},
			},
			wantErr: true,
		},
		{
			name: "missing capabilities",
			card: &AgentCard{
				AgentID: "test-agent",
				OrgID:   "test-org",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AgentCard.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentClient_RegisterAgent(t *testing.T) {
	tests := []struct {
		name      string
		card      *AgentCard
		orgToken  string
		wantToken string
		wantErr   bool
	}{
		{
			name: "valid registration",
			card: &AgentCard{
				AgentID:      "test-agent",
				OrgID:        "test-org",
				Capabilities: []string{"text"},
			},
			orgToken:  "valid-token",
			wantToken: "agent-token",
			wantErr:   false,
		},
		{
			name: "expired org token",
			card: &AgentCard{
				AgentID:      "test-agent",
				OrgID:        "test-org",
				Capabilities: []string{"text"},
			},
			orgToken: "expired-token",
			wantErr:  true,
		},
		{
			name: "org mismatch",
			card: &AgentCard{
				AgentID:      "test-agent",
				OrgID:        "org-x",
				Capabilities: []string{"text"},
			},
			orgToken: "org-y-token",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/agents/token" {
					t.Errorf("expected path /agents/token, got %s", r.URL.Path)
				}

				var req struct {
					AgentCard *AgentCard `json:"agent_card"`
					OrgToken  string     `json:"org_token"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}

				// Simulate different error conditions
				if req.OrgToken == "expired-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if req.AgentCard.OrgID != "test-org" {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"token": "agent-token"}`))
			}))
			defer ts.Close()

			client := NewAgentClient(ts.URL)
			token, err := client.RegisterAgent(tt.card, tt.orgToken)

			if (err != nil) != tt.wantErr {
				t.Errorf("RegisterAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token != tt.wantToken {
				t.Errorf("RegisterAgent() token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestAgentClient_JoinSession(t *testing.T) {
	tests := []struct {
		name       string
		sessionID  string
		agentToken string
		wantErr    bool
	}{
		{
			name:       "valid join",
			sessionID:  "test-session",
			agentToken: "valid-token",
			wantErr:    false,
		},
		{
			name:       "expired token",
			sessionID:  "test-session",
			agentToken: "expired-token",
			wantErr:    true,
		},
		{
			name:       "revoked token",
			sessionID:  "test-session",
			agentToken: "revoked-token",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/sessions/join" {
					t.Errorf("expected path /sessions/join, got %s", r.URL.Path)
				}

				var req struct {
					SessionID string `json:"session_id"`
					Token     string `json:"token"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}

				// Simulate different error conditions
				if req.Token == "expired-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				if req.Token == "revoked-token" {
					w.WriteHeader(http.StatusForbidden)
					return
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			client := NewAgentClient(ts.URL)
			err := client.JoinSession(tt.sessionID, tt.agentToken)

			if (err != nil) != tt.wantErr {
				t.Errorf("JoinSession() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgentToken_Validate(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name    string
		token   *AgentToken
		wantErr bool
	}{
		{
			name: "valid token",
			token: &AgentToken{
				AgentID:      "test-agent",
				OrgID:        "test-org",
				Verified:     true,
				Capabilities: []string{"text"},
				Exp:          now + 3600,
				Iss:          "atoa.platform",
				Aud:          "atoa.agent",
			},
			wantErr: false,
		},
		{
			name: "missing agent_id",
			token: &AgentToken{
				OrgID:        "test-org",
				Verified:     true,
				Capabilities: []string{"text"},
				Exp:          now + 3600,
				Iss:          "atoa.platform",
				Aud:          "atoa.agent",
			},
			wantErr: true,
		},
		{
			name: "expired token",
			token: &AgentToken{
				AgentID:      "test-agent",
				OrgID:        "test-org",
				Verified:     true,
				Capabilities: []string{"text"},
				Exp:          now - 3600,
				Iss:          "atoa.platform",
				Aud:          "atoa.agent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.token.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("AgentToken.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
