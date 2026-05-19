package signing

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
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
// and returns a Signer ready to use.
//
// Supported PEM block contents (auto-detected):
//
//   - PKCS#8 DER (48 bytes for Ed25519) — what openssl genpkey produces.
//   - Raw 64-byte ed25519.PrivateKey form — what Go's ed25519 package
//     marshals to with custom code; also what older NodePulse
//     mint-license dev keys use.
//   - Raw 32-byte Ed25519 seed — expanded via ed25519.NewKeyFromSeed.
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

	var priv ed25519.PrivateKey
	switch len(block.Bytes) {
	case ed25519.PrivateKeySize: // 64
		priv = ed25519.PrivateKey(block.Bytes)
	case ed25519.SeedSize: // 32
		priv = ed25519.NewKeyFromSeed(block.Bytes)
	default:
		// Try PKCS#8 (what openssl genpkey -algorithm ED25519 produces).
		anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("signing: unrecognized key format (size=%d): %w", len(block.Bytes), err)
		}
		ed, ok := anyKey.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("signing: PKCS#8 key is not Ed25519 (got %T)", anyKey)
		}
		priv = ed
	}
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

// NewLocalSignerFromPEMBytes constructs a LocalSigner from PEM-encoded
// private-key bytes already in memory. Same auto-detection logic as
// LoadLocalSigner (PKCS#8, raw 64-byte, raw 32-byte seed) — useful when
// the key comes from a secrets manager / env var instead of a file
// path on disk.
//
// This is the path used by the Fly.io deploy: a Fly Secret named
// NP_LICENSE_SIGNER_KEY holds the PEM bytes; the issuer reads them at
// startup via --signer-key-pem-env=NP_LICENSE_SIGNER_KEY. The key
// never touches the filesystem inside the container.
func NewLocalSignerFromPEMBytes(kid string, raw []byte) (*LocalSigner, error) {
	if kid == "" {
		return nil, errors.New("signing: empty kid")
	}
	if len(raw) == 0 {
		return nil, errors.New("signing: empty PEM bytes")
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("signing: no PEM block in input bytes")
	}
	var priv ed25519.PrivateKey
	switch len(block.Bytes) {
	case ed25519.PrivateKeySize:
		priv = ed25519.PrivateKey(block.Bytes)
	case ed25519.SeedSize:
		priv = ed25519.NewKeyFromSeed(block.Bytes)
	default:
		anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("signing: unrecognized key format (size=%d): %w", len(block.Bytes), err)
		}
		ed, ok := anyKey.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("signing: PKCS#8 key is not Ed25519 (got %T)", anyKey)
		}
		priv = ed
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
