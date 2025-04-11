package atoa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrgCard_Validate(t *testing.T) {
	tests := []struct {
		name    string
		card    *OrgCard
		wantErr bool
	}{
		{
			name: "valid card",
			card: &OrgCard{
				OrgID:     "test-org",
				Name:      "Test Org",
				Domain:    "test.org",
				PublicKey: generateTestPublicKey(t),
			},
			wantErr: false,
		},
		{
			name: "missing org_id",
			card: &OrgCard{
				Name:      "Test Org",
				Domain:    "test.org",
				PublicKey: generateTestPublicKey(t),
			},
			wantErr: true,
		},
		{
			name: "invalid public key",
			card: &OrgCard{
				OrgID:     "test-org",
				Name:      "Test Org",
				Domain:    "test.org",
				PublicKey: "invalid-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("OrgCard.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOrgClient_RegisterOrg(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/orgs/register" {
			t.Errorf("expected path /orgs/register, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"challenge": "test-challenge"}`))
	}))
	defer ts.Close()

	client := NewOrgClient(ts.URL)
	card := &OrgCard{
		OrgID:     "test-org",
		Name:      "Test Org",
		Domain:    "test.org",
		PublicKey: generateTestPublicKey(t),
	}

	challenge, err := client.RegisterOrg(card)
	if err != nil {
		t.Errorf("RegisterOrg() error = %v", err)
	}
	if challenge != "test-challenge" {
		t.Errorf("RegisterOrg() challenge = %v, want %v", challenge, "test-challenge")
	}
}

func TestOrgClient_RequestToken(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/orgs/token" {
			t.Errorf("expected path /orgs/token, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"token": "test-token"}`))
	}))
	defer ts.Close()

	client := NewOrgClient(ts.URL)
	token, err := client.RequestToken("test-org", "test-challenge", "test-signature")
	if err != nil {
		t.Errorf("RequestToken() error = %v", err)
	}
	if token != "test-token" {
		t.Errorf("RequestToken() token = %v, want %v", token, "test-token")
	}
}

// Helper function to generate a test public key
func generateTestPublicKey(t *testing.T) string {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM)
}
