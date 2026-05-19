package issuer

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
)

// EnrollmentTokenPrefix is the visible marker on enrollment tokens.
// Customers see "NP-ENROLL-XXXXXX..." and recognize it as ours.
const EnrollmentTokenPrefix = "NP-ENROLL-"

// generateEnrollmentToken returns a fresh plaintext token and its
// SHA-256 hash (hex). The plaintext is shown to the customer once at
// creation; only the hash is stored.
//
// Format: NP-ENROLL- + 32 base32 chars (160 bits of entropy)
func generateEnrollmentToken() (plaintext, hash string, err error) {
	raw := make([]byte, 20) // 160 bits
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("issuer: rand: %w", err)
	}
	body := strings.TrimRight(base32.StdEncoding.EncodeToString(raw), "=")
	plaintext = EnrollmentTokenPrefix + body
	sum := sha256.Sum256([]byte(plaintext))
	hash = hex.EncodeToString(sum[:])
	return plaintext, hash, nil
}

// hashEnrollmentToken computes SHA-256(plaintext) for lookup. Used
// on the verify side — compare against the stored TokenHash.
func hashEnrollmentToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// newOpaqueID returns a random hex ID with the given prefix.
// Used for license_id, jti, customer_id when not externally provided.
func newOpaqueID(prefix string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}
