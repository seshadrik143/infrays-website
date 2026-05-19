package signing

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func newTestSigner(t *testing.T) (Signer, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	s, err := NewLocalSignerFromKey("np-test-kms-mint", priv)
	if err != nil {
		t.Fatalf("new signer: %v", err)
	}
	return s, pub
}

func TestSign_BasicRoundTrip(t *testing.T) {
	signer, pub := newTestSigner(t)
	defer signer.Close()

	now := time.Now().UTC()
	p := Payload{
		LicenseID:    "lic_test_001",
		CustomerID:   "cust_test",
		CustomerName: "Test Corp",
		Tier:         "professional",
		Iat:          now.Unix(),
		Nbf:          now.Unix(),
		Exp:          now.Add(365 * 24 * time.Hour).Unix(),
	}

	tok, err := Sign(context.Background(), signer, p)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWS segments, got %d", len(parts))
	}

	// Verify by reconstructing the signing input and checking with the
	// caller's pubkey — the same code path NodePulse uses.
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	if !ed25519.Verify(pub, []byte(signingInput), sig) {
		t.Fatal("signature did not verify against signer's public key")
	}
}

func TestSign_HeaderShape(t *testing.T) {
	signer, _ := newTestSigner(t)
	defer signer.Close()
	now := time.Now()
	tok, _ := Sign(context.Background(), signer, Payload{
		LicenseID: "x", CustomerID: "c", Tier: "free",
		Iat: now.Unix(), Exp: now.Add(time.Hour).Unix(),
	})
	parts := strings.Split(tok, ".")
	hdrJSON, _ := base64.RawURLEncoding.DecodeString(parts[0])
	var hdr struct{ Alg, Kid, Typ string }
	_ = json.Unmarshal(hdrJSON, &hdr)
	if hdr.Alg != "EdDSA" {
		t.Errorf("alg: got %q want EdDSA", hdr.Alg)
	}
	if hdr.Kid != "np-test-kms-mint" {
		t.Errorf("kid: got %q", hdr.Kid)
	}
	if hdr.Typ != "license+jws" {
		t.Errorf("typ: got %q", hdr.Typ)
	}
}

func TestSign_DefaultsApplied(t *testing.T) {
	signer, _ := newTestSigner(t)
	defer signer.Close()
	now := time.Now()
	p := Payload{
		LicenseID:  "x",
		CustomerID: "c",
		Tier:       "free",
		Exp:        now.Add(time.Hour).Unix(),
		// V, Iss, Iat intentionally zero
	}
	tok, err := Sign(context.Background(), signer, p)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parts := strings.Split(tok, ".")
	payloadJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var out Payload
	_ = json.Unmarshal(payloadJSON, &out)
	if out.V != 1 {
		t.Errorf("v default: got %d want 1", out.V)
	}
	if out.Iss != "license.infrays.org" {
		t.Errorf("iss default: got %q", out.Iss)
	}
	if out.Iat == 0 {
		t.Error("iat default not applied")
	}
}

func TestSign_RejectsMissingFields(t *testing.T) {
	signer, _ := newTestSigner(t)
	defer signer.Close()
	cases := []struct {
		name string
		p    Payload
	}{
		{"missing license_id", Payload{CustomerID: "c", Tier: "free", Exp: 1}},
		{"missing tier", Payload{LicenseID: "x", CustomerID: "c", Exp: 1}},
		{"missing exp", Payload{LicenseID: "x", CustomerID: "c", Tier: "free"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Sign(context.Background(), signer, c.p); err == nil {
				t.Errorf("expected error for %s", c.name)
			}
		})
	}
}

func TestLocalSigner_LoadFromMissingFile(t *testing.T) {
	if _, err := LoadLocalSigner("k", "/nonexistent/path"); err == nil {
		t.Fatal("expected error")
	}
}

func TestLocalSigner_PublicKeyMatchesPrivate(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	s, err := NewLocalSignerFromKey("k", priv)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	got, err := s.PublicKey(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(pub) {
		t.Errorf("pubkey mismatch")
	}
}
