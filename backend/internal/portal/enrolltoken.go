package portal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"strings"
)

// enrollmentTokenPrefix MUST match issuer.EnrollmentTokenPrefix so
// portal-issued and admin-issued tokens are indistinguishable on the
// enroll side.
const enrollmentTokenPrefix = "NP-ENROLL-"

// generateEnrollmentTokenInternal mirrors issuer.generateEnrollmentToken
// — 20 random bytes → base32 → "NP-ENROLL-" prefix. SHA-256 hex over
// the plaintext is what the store + enroll route compare against.
//
// Kept package-private to avoid cross-importing the issuer package
// (which would create a cycle once portal is wired into issuer.main).
func generateEnrollmentTokenInternal() (plaintext, hashHex string, err error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("portal: rand: %w", err)
	}
	body := strings.TrimRight(base32.StdEncoding.EncodeToString(raw), "=")
	plaintext = enrollmentTokenPrefix + body
	sum := sha256.Sum256([]byte(plaintext))
	hashHex = hex.EncodeToString(sum[:])
	return plaintext, hashHex, nil
}

// plainID returns a 16-byte hex string for opaque row IDs.
func plainID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
