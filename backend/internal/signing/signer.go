// Package signing provides the abstraction over license-signing key
// custody. The issuer service and the kms-mint staff tool both call
// through a Signer interface, so the underlying key material can live
// in:
//
//   - a local Ed25519 PEM file (dev / CI / staging — LocalSigner)
//   - a hosted KMS (production — see kms_gcp.go, kms_vault.go, etc.)
//
// IMPORTANT — AWS KMS Ed25519 caveat:
// AWS KMS does NOT support Ed25519 keys natively. If you want Ed25519
// signatures (which NodePulse verifies as alg=EdDSA), the production
// options are:
//
//   1. GCP KMS — supports algorithm EC_SIGN_ED25519 directly.
//   2. HashiCorp Vault Transit — supports ed25519 keys.
//   3. AWS CloudHSM — Ed25519 via PKCS#11, more operational overhead.
//   4. AWS KMS + ECDSA_NIST_P256 — would require switching NodePulse
//      verification from EdDSA to ES256. Avoid: forces a forced binary
//      upgrade across the customer base.
//
// See docs/LICENSE_KEY_CUSTODY.md for the operational details.
//
// This file defines the interface only. Implementations live in
// local.go, kms_gcp.go (TODO), kms_vault.go (TODO).
package signing

import "context"

// Signer abstracts the cryptographic signing operation. Implementations
// must produce an Ed25519 signature (64 bytes) over the input bytes.
//
// The kid is exposed so the JWS header can declare which key signed
// the token. NodePulse verifies the kid against its embedded trust
// roots; an unknown kid is treated as "binary needs upgrading."
type Signer interface {
	// KID returns the key id that NodePulse will use to look up the
	// matching public key in its trust registry.
	KID() string

	// Sign produces an Ed25519 signature over signingInput. The
	// signingInput is the JWS Compact form's "header.payload" base64url
	// bytes — caller is responsible for assembling it before calling
	// Sign.
	Sign(ctx context.Context, signingInput []byte) ([]byte, error)

	// PublicKey returns the Ed25519 public key matching the private
	// key this signer wraps. Useful for sanity checks and for embedding
	// in the NodePulse binary at build time.
	PublicKey(ctx context.Context) ([]byte, error)

	// Close releases any underlying resources (e.g. KMS clients,
	// HSM sessions). Safe to call multiple times. Returns nil for
	// signers that hold nothing.
	Close() error
}
