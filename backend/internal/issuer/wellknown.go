package issuer

import (
	"encoding/base64"
	"encoding/pem"
	"net/http"
)

// keyInfo is one entry in the /v1/well-known/keys response.
type keyInfo struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Pub string `json:"pub"` // PEM-encoded public key
	Use string `json:"use"` // "signing" | "kill-switch"
}

// GET /v1/well-known/keys
// Returns the public keys this issuer signs with, suitable for
// NodePulse to validate (and for operators to verify our trust roots
// at audit time). Returns a kill-switch announcement when present.
//
// Cache 5 minutes — short enough that a kill-switch announcement
// propagates quickly, long enough to absorb refresh storms.
func (s *Server) handleWellKnownKeys(w http.ResponseWriter, r *http.Request) {
	pub, err := s.cfg.Signer.PublicKey(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "issuer cannot read its own pubkey")
		return
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub})

	w.Header().Set("Cache-Control", "public, max-age=300")
	writeJSON(w, http.StatusOK, map[string]any{
		"keys": []keyInfo{
			{
				Kid: s.cfg.Signer.KID(),
				Alg: "EdDSA",
				Pub: string(pemBytes),
				Use: "signing",
			},
		},
		// Kill-switch announcements (Phase 53 follow-up) live here:
		//   "kill_switch_announcement": {
		//     "signed_payload": "<base64>",
		//     "issued_at": <unix>,
		//     "compromised_kid": "np-prod-2026-01"
		//   }
	})
	_ = base64.StdEncoding
}
