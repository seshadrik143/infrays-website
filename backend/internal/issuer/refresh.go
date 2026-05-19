package issuer

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/signing"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

type refreshRequest struct {
	LicenseID    string `json:"license_id"`
	DeploymentID string `json:"deployment_id"`
	Version      string `json:"version,omitempty"`
}

type refreshResponse struct {
	LicenseJWS         string `json:"license_jws"`
	EntitlementSetID   string `json:"entitlement_set_id,omitempty"`
	RefreshAt          int64  `json:"refresh_at"`
	NextRefreshSeconds int    `json:"next_refresh_seconds"`
}

// handleRefresh re-signs the license that maps to (license_id,
// deployment_id) using current subscription state. Returns 403 if
// the license is revoked, 410 if the subscription has been canceled
// past its current_period_end.
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	var req refreshRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.LicenseID == "" || req.DeploymentID == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "license_id and deployment_id are required")
		return
	}

	ctx := r.Context()
	prev, err := s.cfg.Store.GetLatestLicenseForLicenseID(ctx, req.LicenseID)
	if err != nil {
		writeErr(w, http.StatusForbidden, "unknown_license", "license not recognized")
		return
	}
	if prev.DeploymentID != req.DeploymentID {
		// Mismatch — either a fingerprint rotation (Phase 53 flag for
		// review) or an attempt to use one customer's license from
		// another deployment. Don't 403 immediately — issue an audit
		// event and treat as an anomaly.
		_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
			EventType:    "deployment.fingerprint_mismatch",
			LicenseID:    req.LicenseID,
			DeploymentID: req.DeploymentID,
			Actor:        "system",
			Payload: map[string]any{
				"previous_deployment": prev.DeploymentID,
				"current_deployment":  req.DeploymentID,
			},
		})
		// For Phase 49: accept the new deployment_id (soft fingerprint).
		// Anomaly review happens out-of-band via the admin UI / cron.
	}
	if prev.Revoked {
		writeErr(w, http.StatusForbidden, "revoked", "license revoked: "+prev.RevokedReason)
		return
	}

	sub, err := s.cfg.Store.GetSubscription(ctx, prev.SubscriptionID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "subscription missing")
		return
	}

	now := s.cfg.Now()

	// If subscription is canceled and past current_period_end, return
	// 410. The NodePulse refresh client interprets 410 as "enter grace
	// period." Until then, we keep refreshing — Stripe handles the
	// dunning, and the customer's local grace state machine handles
	// the read-only / disabled transitions.
	if sub.Status == "canceled" && !sub.CurrentPeriodEnd.IsZero() && now.After(sub.CurrentPeriodEnd) {
		graceUntil := sub.CurrentPeriodEnd.Add(time.Duration(s.cfg.DefaultGraceDays) * 24 * time.Hour)
		writeJSON(w, http.StatusGone, map[string]any{
			"error":       "subscription_canceled",
			"message":     "subscription canceled past current_period_end",
			"grace_until": graceUntil.Unix(),
		})
		_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
			EventType:      "license.refresh_denied",
			CustomerID:     prev.CustomerID,
			SubscriptionID: prev.SubscriptionID,
			LicenseID:      prev.LicenseID,
			Actor:          "system",
			Payload:        map[string]any{"reason": "subscription_canceled"},
		})
		return
	}

	cust, err := s.cfg.Store.GetCustomer(ctx, prev.CustomerID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", "customer missing")
		return
	}

	expiresAt := sub.CurrentPeriodEnd
	if expiresAt.IsZero() {
		expiresAt = now.Add(365 * 24 * time.Hour)
	}
	graceUntil := expiresAt.Add(time.Duration(s.cfg.DefaultGraceDays) * 24 * time.Hour)
	newJTI := newOpaqueID("jti")
	entitlementSetID := tierToEntitlementSet(sub.Tier)

	payload := signing.Payload{
		V:                1,
		Iss:              s.cfg.IssuerURL,
		JTI:              newJTI,
		LicenseID:        prev.LicenseID,
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

	_ = s.cfg.Store.CreateLicense(ctx, &store.License{
		ID:               newOpaqueID("licrow"),
		JTI:              newJTI,
		LicenseID:        prev.LicenseID,
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
	_ = s.cfg.Store.UpsertDeployment(ctx, &store.Deployment{
		ID:           newOpaqueID("dep"),
		CustomerID:   cust.ID,
		DeploymentID: req.DeploymentID,
		FirstSeenAt:  now,
		LastSeenAt:   now,
		LastVersion:  req.Version,
		CreatedAt:    now,
	})
	_, _ = s.cfg.Audit.Append(ctx, audit.Entry{
		EventType:      "license.refreshed",
		CustomerID:     cust.ID,
		SubscriptionID: sub.ID,
		LicenseID:      prev.LicenseID,
		DeploymentID:   req.DeploymentID,
		Actor:          "system",
		Payload: map[string]any{
			"jti":             newJTI,
			"tier":            sub.Tier,
			"entitlement_set": entitlementSetID,
			"expires_at":      expiresAt.Format(time.RFC3339),
		},
	})

	writeJSON(w, http.StatusOK, refreshResponse{
		LicenseJWS:         jws,
		EntitlementSetID:   entitlementSetID,
		RefreshAt:          payload.RefreshAt,
		NextRefreshSeconds: s.cfg.RefreshIntervalHours * 3600,
	})
}
