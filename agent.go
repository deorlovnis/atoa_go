package atoa

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AgentCard represents an agent's identity and capabilities
type AgentCard struct {
	AgentID      string   `json:"agent_id"`
	OrgID        string   `json:"org_id"`
	Capabilities []string `json:"capabilities"`
	Endpoints    []string `json:"endpoints"`
	Verified     bool     `json:"verified"`
}

// Validate checks if the AgentCard has all required fields
func (ac *AgentCard) Validate() error {
	if ac.AgentID == "" {
		return errors.New("agent_id is required")
	}
	if ac.OrgID == "" {
		return errors.New("org_id is required")
	}
	if len(ac.Capabilities) == 0 {
		return errors.New("at least one capability is required")
	}
	return nil
}

// AgentToken represents the JWT token issued to an agent
type AgentToken struct {
	AgentID      string   `json:"agent_id"`
	OrgID        string   `json:"org_id"`
	Verified     bool     `json:"verified"`
	Capabilities []string `json:"capabilities"`
	Exp          int64    `json:"exp"`
	Iss          string   `json:"iss"`
	Aud          string   `json:"aud"`
}

// Validate checks if the AgentToken has all required fields and is not expired
func (at *AgentToken) Validate() error {
	if at.AgentID == "" {
		return errors.New("agent_id is required")
	}
	if at.OrgID == "" {
		return errors.New("org_id is required")
	}
	if at.Exp == 0 {
		return errors.New("expiration time is required")
	}
	if at.Iss == "" {
		return errors.New("issuer is required")
	}
	if at.Aud == "" {
		return errors.New("audience is required")
	}

	// Check if token is expired
	if time.Now().Unix() > at.Exp {
		return errors.New("token is expired")
	}

	return nil
}

// ParseAgentToken parses a JWT token string into an AgentToken
func ParseAgentToken(tokenString string) (*AgentToken, error) {
	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// TODO: Get the public key from the token header or a trusted source
		// For now, return nil as we're just parsing
		return nil, nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt())

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Convert claims to AgentToken
	agentToken := &AgentToken{
		AgentID:      getStringClaim(claims, "agent_id"),
		OrgID:        getStringClaim(claims, "org_id"),
		Verified:     getBoolClaim(claims, "verified"),
		Capabilities: getStringSliceClaim(claims, "capabilities"),
		Exp:          int64(getFloatClaim(claims, "exp")),
		Iss:          getStringClaim(claims, "iss"),
		Aud:          getStringClaim(claims, "aud"),
	}

	// Validate the token structure
	if err := agentToken.Validate(); err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return agentToken, nil
}

// Helper functions to safely extract claims
func getStringClaim(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key].(string); ok {
		return val
	}
	return ""
}

func getBoolClaim(claims jwt.MapClaims, key string) bool {
	if val, ok := claims[key].(bool); ok {
		return val
	}
	return false
}

func getFloatClaim(claims jwt.MapClaims, key string) float64 {
	if val, ok := claims[key].(float64); ok {
		return val
	}
	return 0
}

func getStringSliceClaim(claims jwt.MapClaims, key string) []string {
	if val, ok := claims[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}
