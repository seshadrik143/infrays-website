package issuer

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// enrollRequest is the POST /v1/enroll body shape.
type enrollRequest struct {
	EnrollmentToken string `json:"enrollment_token"`
	DeploymentID    string `json:"deployment_id"`
	DeploymentName  string `json:"deployment_name,omitempty"`
	Version         string `json:"version,omitempty"`
}

// enrollResponse is the OK shape on POST /v1/enroll.
type enrollResponse struct {
	LicenseJWS         string `json:"license_jws"`
	EntitlementSetID   string `json:"entitlement_set_id,omitempty"`
	RefreshAt          int64  `json:"refresh_at"`
	RefreshURL         string `json:"refresh_url"`
	NextRefreshSeconds int    `json:"next_refresh_seconds"`
}

func (s *Server) handleEnroll(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var req enrollRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.EnrollmentToken = strings.TrimSpace(req.EnrollmentToken)
	req.DeploymentID = strings.TrimSpace(req.DeploymentID)
	if req.EnrollmentToken == "" || req.DeploymentID == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "enrollment_token and deployment_id are required")
		return
	}

	ctx := r.Context()
	tokenHash := hashEnrollmentToken(req.EnrollmentToken)

	tok, err := s.cfg.Store.GetEnrollmentTokenByHash(ctx, tokenHash)
	if err != nil {
		// Same error shape for unknown token and expired token —
		// don't help an attacker probe which tokens exist.
		writeErr(w, http.StatusForbidden, "invalid_token", "enrollment token not recognized")
		return
	}
	now := s.cfg.Now()
	if !tok.ExpiresAt.IsZero() && now.After(tok.ExpiresAt) {
		writeErr(w, http.StatusForbidden, "expired", "enrollment token expired; regenerate from the customer portal")
		return
	}

	// Idempotent re-redeem: if the same deployment is re-presenting
	// this token, return the cached response.
	if !tok.ConsumedAt.IsZero() && tok.ConsumedByDeployment == req.DeploymentID && tok.ConsumedResponseJWS != "" {
		writeJSON(w, http.StatusOK, decodeCachedEnroll(tok.ConsumedResponseJWS, s.cfg))
		return
	}
	if !tok.ConsumedAt.IsZero() && tok.ConsumedByDeployment != req.DeploymentID {
		writeErr(w, http.StatusForbidden, "consumed_by_other", "this token was redeemed by a different deployment")
		return
	}

	// Build the license payload from the subscription state.
	sub, err := s.cfg.Store.GetSubscription(ctx, tok.SubscriptionID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "subscription missing")
		return
	}
	cust, err := s.cfg.Store.GetCustomer(ctx, tok.CustomerID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "customer missing")
		return
	}

	licenseID := newOpaqueID("lic")
	jti := newOpaqueID("jti")
	entitlementSetID := tierToEntitlementSet(sub.Tier)
	expiresAt := sub.CurrentPeriodEnd
	if expiresAt.IsZero() {
		// Trial or offline — default to 365 days.
		expiresAt = now.Add(365 * 24 * time.Hour)
	}
	graceUntil := expiresAt.Add(time.Duration(s.cfg.DefaultGraceDays) * 24 * time.Hour)

	payload := signing.Payload{
		V:                1,
		Iss:              s.cfg.IssuerURL,
		JTI:              jti,
		LicenseID:        licenseID,
		CustomerID:       cust.ID,
		CustomerName:     cust.Name,
		SubscriptionID:   sub.ID,
		DeploymentID:     req.DeploymentID,
		Tier:             sub.Tier,
		EntitlementSetID: entitlementSetID,
		Iat:              now.Unix(),
		Nbf:              now.Unix(),
		Exp:              expiresAt.Unix(),
		GraceUntil:       graceUntil.Unix(),
		RefreshAt:        now.Add(time.Duration(s.cfg.RefreshIntervalHours) * time.Hour).Unix(),
	}
	// Copy entitlement limits onto the payload for fast local
	// enforcement. The manifest stays authoritative.
	if es, err := s.cfg.Store.GetEntitlementSet(ctx, entitlementSetID); err == nil {
		payload.MaxAgents = es.Limits.MaxAgents
		payload.MaxMetricsPerSec = es.Limits.MaxMetricsPerSec
		payload.MaxLogGBPerDay = es.Limits.MaxLogGBPerDay
		payload.MaxAlertRules = es.Limits.MaxAlertRules
		payload.RetentionDays = es.Limits.RetentionDays
	}

	jws, err := signing.Sign(ctx, s.cfg.Signer, payload)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sign_failed", "issuer could not sign license")
		return
	}

	// Persist deployment + license + audit BEFORE marking token consumed,
	// so a crash mid-handler can replay without violating idempotency.
	_ = s.cfg.Store.UpsertDeployment(ctx, &store.Deployment{
		ID:             newOpaqueID("dep"),
		CustomerID:     cust.ID,
		DeploymentID:   req.DeploymentID,
		DeploymentName: req.DeploymentName,
		FirstSeenAt:    now,
		LastSeenAt:     now,
		LastVersion:    req.Version,
		CreatedAt:      now,
	})
	_ = s.cfg.Store.CreateLicense(ctx, &store.License{
		ID:               newOpaqueID("licrow"),
		JTI:              jti,
		LicenseID:        licenseID,
		CustomerID:       cust.ID,
		SubscriptionID:   sub.ID,
		DeploymentID:     req.DeploymentID,
		EntitlementSetID: entitlementSetID,
		Tier:             sub.Tier,
		Iat:              now,
		NotBefore:        now,
		ExpiresAt:        expiresAt,
		GraceUntil:       graceUntil,
		Kid:              s.cfg.Signer.KID(),
		PayloadJWS:       jws,
		CreatedAt:        now,
	})
	_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
		EventType:      "license.issued",
		CustomerID:     cust.ID,
		SubscriptionID: sub.ID,
		LicenseID:      licenseID,
		DeploymentID:   req.DeploymentID,
		Actor:          "system",
		Payload: map[string]any{
			"jti":             jti,
			"tier":            sub.Tier,
			"entitlement_set": entitlementSetID,
			"expires_at":      expiresAt.Format(time.RFC3339),
			"via":             "enrollment",
		},
	})

	// Cache the response in the token for idempotent replay.
	resp := enrollResponse{
		LicenseJWS:         jws,
		EntitlementSetID:   entitlementSetID,
		RefreshAt:          payload.RefreshAt,
		RefreshURL:         "https://" + s.cfg.IssuerURL,
		NextRefreshSeconds: s.cfg.RefreshIntervalHours * 3600,
	}
	respJSON, _ := json.Marshal(resp)
	if err := s.cfg.Store.ConsumeEnrollmentToken(ctx, tokenHash, req.DeploymentID, string(respJSON)); err != nil {
		if errors.Is(err, store.ErrConsumed) {
			writeErr(w, http.StatusForbidden, "consumed_by_other", "this token was redeemed by a different deployment")
			return
		}
		// Token consume failed but license is already issued. Log
		// loudly — this is an inconsistency the operator must
		// investigate, but the customer's license is valid.
		_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
			EventType: "enrollment.consume_failed",
			LicenseID: licenseID,
			Actor:     "system",
			Payload:   map[string]any{"error": err.Error()},
		})
	}
	writeJSON(w, http.StatusOK, resp)
}

// tierToEntitlementSet picks the default entitlement set id for a
// given tier. Real production might use Stripe metadata to pick a
// specific set ID; for Phase 49 we use a simple mapping.
func tierToEntitlementSet(tier string) string {
	switch tier {
	case "free", "community":
		return "free-v1"
	case "professional", "pro":
		return "professional-v1"
	case "enterprise":
		return "enterprise-v1"
	}
	return "free-v1"
}

func decodeCachedEnroll(cached string, _ Config) enrollResponse {
	var r enrollResponse
	_ = json.Unmarshal([]byte(cached), &r)
	return r
}

// ── small response helpers (shared with the rest of issuer/) ─────

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeErr(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": msg,
	})
}
