package atoa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Offer represents a service offer from an agent
type Offer struct {
	Header       OfferHeader       `json:"header"`
	Metadata     OfferMetadata     `json:"metadata"`
	Requirements OfferRequirements `json:"requirements"`
}

// OfferHeader contains the basic information about an offer
type OfferHeader struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// OfferMetadata contains additional information about the offer
type OfferMetadata struct {
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
	Tags      []string `json:"tags"`
	Version   string   `json:"version"`
}

// OfferRequirements specifies what is needed to use the offer
type OfferRequirements struct {
	Capabilities []string `json:"capabilities"`
	MinVersion   string   `json:"min_version"`
}

// Session represents an active communication session between agents
type Session struct {
	SessionID   string `json:"session_id"`
	OfferID     string `json:"offer_id"`
	FromAgentID string `json:"from_agent_id"`
	ToAgentID   string `json:"to_agent_id"`
	CreatedAt   string `json:"created_at"`
	ExpiresAt   string `json:"expires_at"`
	Status      string `json:"status"`
}

// ListOffers retrieves a list of available offers
func (c *AgentClient) ListOffers(ctx context.Context) ([]Offer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/offers", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var offers []Offer
	if err := json.NewDecoder(resp.Body).Decode(&offers); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return offers, nil
}

// CreateSession establishes a new session with an offer
func (c *AgentClient) CreateSession(ctx context.Context, offerID string) (*Session, error) {
	payload := struct {
		OfferID string `json:"offer_id"`
	}{
		OfferID: offerID,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/sessions", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Set authorization header
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &session, nil
}
