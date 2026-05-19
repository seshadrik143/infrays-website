// JWS Compact Serialization construction for NodePulse licenses.
//
// This is the issuer-side mirror of the verifier in NodePulse's
// server/license/jws.go. The payload schema MUST stay in sync — if you
// add a field here, also add it to the verifier's jwsPayload struct,
// otherwise older NodePulse binaries will silently drop the new field
// (json.Unmarshal ignores unknown keys).
//
// The deliberate duplication of the payload schema (rather than
// importing the NodePulse package) keeps the issuer service free of
// any NodePulse dependency — the two codebases ship and deploy
// independently.
package signing

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Payload is the v1 license claim set. JSON tags match NodePulse's
// jwsPayload exactly — schema is the wire contract.
type Payload struct {
	V int `json:"v"`

	// Identity & lifecycle.
	Iss              string `json:"iss"`
	JTI              string `json:"jti"`
	LicenseID        string `json:"license_id"`
	CustomerID       string `json:"customer_id"`
	CustomerName     string `json:"customer_name,omitempty"`
	SubscriptionID   string `json:"subscription_id"`
	DeploymentID     string `json:"deployment_id"`
	Tier             string `json:"tier"`
	EntitlementSetID string `json:"entitlement_set_id,omitempty"`
	TenantID         string `json:"tenant_id,omitempty"`

	// Time bounds (unix seconds — JWS convention).
	Iat        int64 `json:"iat"`
	Nbf        int64 `json:"nbf"`
	Exp        int64 `json:"exp"`
	GraceUntil int64 `json:"grace_until,omitempty"`
	RefreshAt  int64 `json:"refresh_at,omitempty"`

	// Limits — convenience copies of entitlement-manifest values.
	MaxAgents        int64 `json:"max_agents,omitempty"`
	MaxMetricsPerSec int64 `json:"max_metrics_per_sec,omitempty"`
	MaxLogGBPerDay   int64 `json:"max_log_gb_per_day,omitempty"`
	MaxAlertRules    int64 `json:"max_alert_rules,omitempty"`
	RetentionDays    int   `json:"retention_days,omitempty"`

	// Posture flags.
	TelemetryMinimal bool `json:"telemetry_minimal,omitempty"`
	OfflineMode      bool `json:"offline_mode,omitempty"`
}

type header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// Sign produces a JWS Compact token signing payload with the given
// signer. Returns "header_b64.payload_b64.signature_b64".
//
// Defaults applied:
//   - V is set to 1 if zero
//   - Iss is set to "license.infrays.org" if empty
//   - Iat is set to time.Now().Unix() if zero
//
// Caller is responsible for setting LicenseID, CustomerID, Tier, and
// the time bounds (Nbf, Exp). Validation is minimal here — the issuer
// service does the business-layer checks.
func Sign(ctx context.Context, signer Signer, p Payload) (string, error) {
	if signer == nil {
		return "", errors.New("signing: nil signer")
	}
	if p.LicenseID == "" {
		return "", errors.New("signing: empty license_id")
	}
	if p.Tier == "" {
		return "", errors.New("signing: empty tier")
	}
	if p.Exp == 0 {
		return "", errors.New("signing: missing exp")
	}
	if p.V == 0 {
		p.V = 1
	}
	if p.Iss == "" {
		p.Iss = "license.infrays.org"
	}
	if p.Iat == 0 {
		p.Iat = time.Now().Unix()
	}

	hdr := header{Alg: "EdDSA", Kid: signer.KID(), Typ: "license+jws"}
	hdrJSON, err := json.Marshal(hdr)
	if err != nil {
		return "", fmt.Errorf("signing: marshal header: %w", err)
	}
	payloadJSON, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("signing: marshal payload: %w", err)
	}
	hdrB64 := base64.RawURLEncoding.EncodeToString(hdrJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := hdrB64 + "." + payloadB64

	sig, err := signer.Sign(ctx, []byte(signingInput))
	if err != nil {
		return "", fmt.Errorf("signing: signer.Sign: %w", err)
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + sigB64, nil
}
