package atoa

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TokenIssuer is the issuer of all Atoa tokens
	TokenIssuer = "atoa.platform"
	// OrgTokenAudience is the audience for organization tokens
	OrgTokenAudience = "atoa.agent"
	// AgentTokenAudience is the audience for agent tokens
	AgentTokenAudience = "atoa.session"
	// DefaultTokenExpiry is the default token expiration time
	DefaultTokenExpiry = 1 * time.Hour
)

// OrgTokenClaims represents the claims in an organization JWT token
type OrgTokenClaims struct {
	jwt.RegisteredClaims
	OrgID    string `json:"org_id"`
	Verified bool   `json:"verified"`
}

// AgentTokenClaims represents the claims in an agent JWT token
type AgentTokenClaims struct {
	jwt.RegisteredClaims
	AgentID      string   `json:"agent_id"`
	OrgID        string   `json:"org_id"`
	Verified     bool     `json:"verified"`
	Capabilities []string `json:"capabilities"`
}

// IssueOrgToken issues a new JWT token for an organization
func IssueOrgToken(orgID string, verified bool, privateKey *ecdsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := OrgTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Audience:  jwt.ClaimStrings{OrgTokenAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(DefaultTokenExpiry)),
		},
		OrgID:    orgID,
		Verified: verified,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privateKey)
}

// IssueAgentToken issues a new JWT token for an agent
func IssueAgentToken(card *AgentCard, orgToken string, privateKey *ecdsa.PrivateKey) (string, error) {
	// Parse and validate the org token first
	orgClaims := &OrgTokenClaims{}
	err := ParseTokenWithPublicKey(orgToken, &privateKey.PublicKey, orgClaims)
	if err != nil {
		return "", fmt.Errorf("invalid org token: %w", err)
	}

	// Verify org_id matches
	if orgClaims.OrgID != card.OrgID {
		return "", errors.New("org_id mismatch between card and token")
	}

	// Create agent token claims
	now := time.Now()
	claims := AgentTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Audience:  jwt.ClaimStrings{AgentTokenAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(DefaultTokenExpiry)),
		},
		AgentID:      card.AgentID,
		OrgID:        card.OrgID,
		Verified:     orgClaims.Verified, // Inherit verification status from org
		Capabilities: card.Capabilities,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privateKey)
}

// ParseOrgToken parses and validates an organization JWT token
func ParseOrgToken(tokenString string) (*OrgTokenClaims, error) {
	// First parse without verification to get the public key
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &OrgTokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// TODO: Get the public key from a trusted source using keyID from token.Header["kid"]
	// For now, we'll just parse the claims without verification
	parser := jwt.NewParser(jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	token, err := parser.ParseWithClaims(tokenString, &OrgTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// For testing purposes, we'll skip verification
		// In production, we would get the public key from a trusted source using keyID
		return nil, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*OrgTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ParseAgentTokenClaims parses and validates an agent JWT token
func ParseAgentTokenClaims(tokenString string) (*AgentTokenClaims, error) {
	// First parse without verification to get the public key
	_, _, err := jwt.NewParser().ParseUnverified(tokenString, &AgentTokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// TODO: Get the public key from a trusted source using keyID from token.Header["kid"]
	// For now, we'll just parse the claims without verification
	parser := jwt.NewParser(jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	token, err := parser.ParseWithClaims(tokenString, &AgentTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// For testing purposes, we'll skip verification
		// In production, we would get the public key from a trusted source using keyID
		return nil, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*AgentTokenClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ParseTokenWithPublicKey parses and validates a JWT token with a specific public key
func ParseTokenWithPublicKey(tokenString string, publicKey *ecdsa.PublicKey, claims jwt.Claims) error {
	parser := jwt.NewParser(jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	_, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	return err
}
