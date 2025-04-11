package atoa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// OrgCard represents an organization's identity and verification status
type OrgCard struct {
	OrgID     string `json:"org_id"`
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	PublicKey string `json:"public_key"`
	Verified  bool   `json:"verified"`
}

// Validate checks if the OrgCard has all required fields and valid public key
func (oc *OrgCard) Validate() error {
	if oc.OrgID == "" {
		return errors.New("org_id is required")
	}
	if oc.Name == "" {
		return errors.New("name is required")
	}
	if oc.Domain == "" {
		return errors.New("domain is required")
	}
	if oc.PublicKey == "" {
		return errors.New("public_key is required")
	}

	// Validate public key format
	block, _ := pem.Decode([]byte(oc.PublicKey))
	if block == nil {
		return errors.New("invalid public key format")
	}
	if block.Type != "PUBLIC KEY" {
		return errors.New("public key must be in PEM format")
	}

	return nil
}

// SignChallenge signs the given challenge using the provided private key
func SignChallenge(challenge string, privateKey *ecdsa.PrivateKey) (string, error) {
	hash := sha256.Sum256([]byte(challenge))
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies a signature against a challenge using the public key
func VerifySignature(challenge, signature, publicKeyPEM string) (bool, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return false, errors.New("invalid public key format")
	}

	pubKey, err := parsePublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	hash := sha256.Sum256([]byte(challenge))
	return ecdsa.VerifyASN1(pubKey, hash[:], sig), nil
}

// parsePublicKey parses a DER-encoded public key
func parsePublicKey(der []byte) (*ecdsa.PublicKey, error) {
	pub, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an ECDSA key")
	}

	return ecdsaPub, nil
}
