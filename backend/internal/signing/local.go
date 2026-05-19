package signing

import (
	"context"
	"crypto/ed25519"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// LocalSigner is a Signer backed by an Ed25519 private key loaded from
// a PEM file on local disk.
//
// SECURITY NOTE — for development, CI, and staging only.
// Production signing keys must never sit on a developer's laptop or in
// a service container's filesystem. Use a KMS-backed Signer
// implementation (kms_gcp.go, kms_vault.go) for prod.
//
// The kid embedded here is the key id NodePulse will see in JWS
// headers; it must match the public key registered in NodePulse's
// trust registry (server/license/keys_*.go).
type LocalSigner struct {
	kid  string
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

// LoadLocalSigner reads a PEM-encoded Ed25519 private key from path
// and returns a Signer ready to use. The path must contain the raw
// key bytes (32-byte seed expanded to 64 bytes via Go's
// ed25519.NewKeyFromSeed semantics — same format mint-license
// generates).
//
// Caller is responsible for ensuring the file has restrictive
// permissions (0400) before this is called.
func LoadLocalSigner(kid, path string) (*LocalSigner, error) {
	if kid == "" {
		return nil, errors.New("signing: empty kid")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("signing: read key file: %w", err)
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("signing: no PEM block in key file")
	}
	if len(block.Bytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("signing: unexpected private-key size %d (want %d)", len(block.Bytes), ed25519.PrivateKeySize)
	}
	priv := ed25519.PrivateKey(block.Bytes)
	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return nil, errors.New("signing: failed to derive public key")
	}
	return &LocalSigner{kid: kid, priv: priv, pub: pub}, nil
}

// NewLocalSignerFromKey constructs a LocalSigner from an in-memory key.
// Useful for tests that generate ephemeral keys via ed25519.GenerateKey.
func NewLocalSignerFromKey(kid string, priv ed25519.PrivateKey) (*LocalSigner, error) {
	if kid == "" {
		return nil, errors.New("signing: empty kid")
	}
	if len(priv) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("signing: invalid private-key size %d", len(priv))
	}
	pub, _ := priv.Public().(ed25519.PublicKey)
	return &LocalSigner{kid: kid, priv: priv, pub: pub}, nil
}

func (s *LocalSigner) KID() string { return s.kid }

func (s *LocalSigner) Sign(_ context.Context, signingInput []byte) ([]byte, error) {
	return ed25519.Sign(s.priv, signingInput), nil
}

func (s *LocalSigner) PublicKey(_ context.Context) ([]byte, error) {
	out := make([]byte, len(s.pub))
	copy(out, s.pub)
	return out, nil
}

// Close zeroes the private-key bytes in memory. Best-effort — Go's
// garbage collector may have already moved them, so this is hygiene
// rather than guarantee.
func (s *LocalSigner) Close() error {
	for i := range s.priv {
		s.priv[i] = 0
	}
	return nil
}
