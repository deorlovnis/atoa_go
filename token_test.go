package atoa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueOrgToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	token, err := IssueOrgToken("test-org", true, privateKey)
	if err != nil {
		t.Errorf("IssueOrgToken() error = %v", err)
	}

	claims := &OrgTokenClaims{}
	err = ParseTokenWithPublicKey(token, &privateKey.PublicKey, claims)
	if err != nil {
		t.Errorf("ParseTokenWithPublicKey() error = %v", err)
	}

	if claims.OrgID != "test-org" {
		t.Errorf("claims.OrgID = %v, want %v", claims.OrgID, "test-org")
	}
	if !claims.Verified {
		t.Errorf("claims.Verified = %v, want %v", claims.Verified, true)
	}
	if claims.Issuer != TokenIssuer {
		t.Errorf("claims.Issuer = %v, want %v", claims.Issuer, TokenIssuer)
	}
	if claims.Audience[0] != OrgTokenAudience {
		t.Errorf("claims.Audience = %v, want %v", claims.Audience, OrgTokenAudience)
	}
}

func TestIssueAgentToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// First issue an org token
	orgToken, err := IssueOrgToken("test-org", true, privateKey)
	if err != nil {
		t.Fatalf("failed to issue org token: %v", err)
	}

	// Create an agent card
	card := &AgentCard{
		AgentID:      "test-agent",
		OrgID:        "test-org",
		Capabilities: []string{"text", "form"},
	}

	// Issue an agent token
	token, err := IssueAgentToken(card, orgToken, privateKey)
	if err != nil {
		t.Errorf("IssueAgentToken() error = %v", err)
	}

	claims := &AgentTokenClaims{}
	err = ParseTokenWithPublicKey(token, &privateKey.PublicKey, claims)
	if err != nil {
		t.Errorf("ParseTokenWithPublicKey() error = %v", err)
	}

	if claims.AgentID != "test-agent" {
		t.Errorf("claims.AgentID = %v, want %v", claims.AgentID, "test-agent")
	}
	if claims.OrgID != "test-org" {
		t.Errorf("claims.OrgID = %v, want %v", claims.OrgID, "test-org")
	}
	if !claims.Verified {
		t.Errorf("claims.Verified = %v, want %v", claims.Verified, true)
	}
	if len(claims.Capabilities) != 2 {
		t.Errorf("len(claims.Capabilities) = %v, want %v", len(claims.Capabilities), 2)
	}
	if claims.Issuer != TokenIssuer {
		t.Errorf("claims.Issuer = %v, want %v", claims.Issuer, TokenIssuer)
	}
	if claims.Audience[0] != AgentTokenAudience {
		t.Errorf("claims.Audience = %v, want %v", claims.Audience, AgentTokenAudience)
	}
}

func TestIssueAgentToken_OrgMismatch(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Issue an org token for org-x
	orgToken, err := IssueOrgToken("org-x", true, privateKey)
	if err != nil {
		t.Fatalf("failed to issue org token: %v", err)
	}

	// Create an agent card for org-y
	card := &AgentCard{
		AgentID:      "test-agent",
		OrgID:        "org-y",
		Capabilities: []string{"text"},
	}

	// Try to issue an agent token
	_, err = IssueAgentToken(card, orgToken, privateKey)
	if err == nil {
		t.Error("IssueAgentToken() error = nil, want error")
	}
}

func TestIssueAgentToken_ExpiredOrgToken(t *testing.T) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	// Create an expired org token
	claims := OrgTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    TokenIssuer,
			Audience:  jwt.ClaimStrings{OrgTokenAudience},
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
		OrgID:    "test-org",
		Verified: true,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	orgToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	// Create an agent card
	card := &AgentCard{
		AgentID:      "test-agent",
		OrgID:        "test-org",
		Capabilities: []string{"text"},
	}

	// Try to issue an agent token
	_, err = IssueAgentToken(card, orgToken, privateKey)
	if err == nil {
		t.Error("IssueAgentToken() error = nil, want error")
	}
}
