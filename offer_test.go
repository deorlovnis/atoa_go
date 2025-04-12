package atoa

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var testPrivateKey *ecdsa.PrivateKey

func init() {
	var err error
	testPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic("failed to generate test private key: " + err.Error())
	}
}

func TestListOffers(t *testing.T) {
	tests := []struct {
		name          string
		agentVerified bool
		tokenValid    bool
		wantErr       bool
	}{
		{
			name:          "verified agent with valid token",
			agentVerified: true,
			tokenValid:    true,
			wantErr:       false,
		},
		{
			name:          "unverified agent with valid token",
			agentVerified: false,
			tokenValid:    true,
			wantErr:       true,
		},
		{
			name:          "verified agent with invalid token",
			agentVerified: true,
			tokenValid:    false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/offers" {
					t.Errorf("expected path /offers, got %s", r.URL.Path)
				}

				// Check authorization header
				auth := r.Header.Get("Authorization")
				if tt.tokenValid {
					if auth != "Bearer valid-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if !tt.agentVerified {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				} else if auth != "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Return mock offers
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`[{
					"header": {
						"id": "offer-1",
						"title": "Test Offer",
						"description": "A test offer",
						"type": "service"
					},
					"metadata": {
						"created_at": "2024-03-20T12:00:00Z",
						"updated_at": "2024-03-20T12:00:00Z",
						"tags": ["test"],
						"version": "1.0"
					},
					"requirements": {
						"capabilities": ["text"],
						"min_version": "1.0"
					}
				}]`))
			}))
			defer ts.Close()

			// Create a test agent client
			client := &AgentClient{
				BaseURL: ts.URL,
				HTTP:    &http.Client{},
				Token:   "test-token",
			}

			if tt.tokenValid {
				client.Token = "valid-token"
			} else {
				client.Token = "invalid-token"
			}

			// Test ListOffers
			offers, err := client.ListOffers(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ListOffers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(offers) == 0 {
					t.Error("ListOffers() returned empty list")
				}
				for _, offer := range offers {
					if offer.Header.Title == "" {
						t.Error("ListOffers() returned offer with empty title")
					}
					if offer.Header.Type == "" {
						t.Error("ListOffers() returned offer with empty type")
					}
				}
			}
		})
	}
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name          string
		agentVerified bool
		tokenValid    bool
		offerID       string
		wantErr       bool
	}{
		{
			name:          "verified agent with valid token and offer",
			agentVerified: true,
			tokenValid:    true,
			offerID:       "valid-offer",
			wantErr:       false,
		},
		{
			name:          "unverified agent with valid token",
			agentVerified: false,
			tokenValid:    true,
			offerID:       "valid-offer",
			wantErr:       true,
		},
		{
			name:          "verified agent with invalid token",
			agentVerified: true,
			tokenValid:    false,
			offerID:       "valid-offer",
			wantErr:       true,
		},
		{
			name:          "verified agent with invalid offer",
			agentVerified: true,
			tokenValid:    true,
			offerID:       "invalid-offer",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST request, got %s", r.Method)
				}
				if r.URL.Path != "/sessions" {
					t.Errorf("expected path /sessions, got %s", r.URL.Path)
				}

				// Check authorization header
				auth := r.Header.Get("Authorization")
				if tt.tokenValid {
					if auth != "Bearer valid-token" {
						w.WriteHeader(http.StatusUnauthorized)
						return
					}
					if !tt.agentVerified {
						w.WriteHeader(http.StatusForbidden)
						return
					}
				} else if auth != "" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Check offer ID
				var req struct {
					OfferID string `json:"offer_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}
				if req.OfferID != tt.offerID || req.OfferID == "invalid-offer" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// Return mock session
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{
					"session_id": "session-1",
					"offer_id": "` + tt.offerID + `",
					"from_agent_id": "agent-1",
					"to_agent_id": "agent-2",
					"created_at": "2024-03-20T12:00:00Z",
					"expires_at": "2024-03-20T13:00:00Z",
					"status": "active"
				}`))
			}))
			defer ts.Close()

			// Create a test agent client
			client := &AgentClient{
				BaseURL: ts.URL,
				HTTP:    &http.Client{},
				Token:   "test-token",
			}

			if tt.tokenValid {
				client.Token = "valid-token"
			} else {
				client.Token = "invalid-token"
			}

			// Test CreateSession
			session, err := client.CreateSession(context.Background(), tt.offerID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if session == nil {
					t.Error("CreateSession() returned nil session")
				}
				if session.SessionID == "" {
					t.Error("CreateSession() returned session with empty ID")
				}
			}
		})
	}
}
