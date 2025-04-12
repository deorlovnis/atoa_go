package atoa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OrgClient handles organization registration and authentication
type OrgClient struct {
	BaseURL string
	HTTP    *http.Client
}

// NewOrgClient creates a new OrgClient with the given base URL
func NewOrgClient(baseURL string) *OrgClient {
	return &OrgClient{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}
}

// RegisterOrg registers a new organization and returns a challenge
func (c *OrgClient) RegisterOrg(card *OrgCard) (string, error) {
	if err := card.Validate(); err != nil {
		return "", fmt.Errorf("invalid org card: %w", err)
	}

	payload, err := json.Marshal(card)
	if err != nil {
		return "", fmt.Errorf("failed to marshal org card: %w", err)
	}

	resp, err := c.HTTP.Post(
		fmt.Sprintf("%s/orgs/register", c.BaseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to register org: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result struct {
		Challenge string `json:"challenge"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Challenge, nil
}

// RequestToken requests a JWT token after signing the challenge
func (c *OrgClient) RequestToken(orgID, challenge, signature string) (string, error) {
	payload := struct {
		OrgID     string `json:"org_id"`
		Challenge string `json:"challenge"`
		Signature string `json:"signature"`
	}{
		OrgID:     orgID,
		Challenge: challenge,
		Signature: signature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTP.Post(
		fmt.Sprintf("%s/orgs/token", c.BaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Token, nil
}

// AgentClient handles agent registration and authentication
type AgentClient struct {
	AgentCard AgentCard
	OrgToken  string
	Token     string
	BaseURL   string
	HTTP      *http.Client
}

// NewAgentClient creates a new AgentClient with the given base URL
func NewAgentClient(baseURL string) *AgentClient {
	return &AgentClient{
		BaseURL: baseURL,
		HTTP:    &http.Client{},
	}
}

// RegisterAgent registers a new agent and returns a JWT token
func (c *AgentClient) RegisterAgent(card *AgentCard, orgToken string) (string, error) {
	if err := card.Validate(); err != nil {
		return "", fmt.Errorf("invalid agent card: %w", err)
	}

	payload := struct {
		AgentCard *AgentCard `json:"agent_card"`
		OrgToken  string     `json:"org_token"`
	}{
		AgentCard: card,
		OrgToken:  orgToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTP.Post(
		fmt.Sprintf("%s/agents/token", c.BaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", fmt.Errorf("failed to register agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Token, nil
}

// JoinSession attempts to join a session using the agent's token
func (c *AgentClient) JoinSession(sessionID, agentToken string) error {
	payload := struct {
		SessionID string `json:"session_id"`
		Token     string `json:"token"`
	}{
		SessionID: sessionID,
		Token:     agentToken,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTP.Post(
		fmt.Sprintf("%s/sessions/join", c.BaseURL),
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to join session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("join failed with status %d", resp.StatusCode)
	}

	return nil
}
