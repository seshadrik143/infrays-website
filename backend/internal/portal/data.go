package portal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// ─── Subscriptions ──────────────────────────────────────────────────

type subscriptionDTO struct {
	ID                   string    `json:"id"`
	Tier                 string    `json:"tier"`
	Status               string    `json:"status"`
	CurrentPeriodStart   time.Time `json:"current_period_start"`
	CurrentPeriodEnd     time.Time `json:"current_period_end"`
	CancelAt             time.Time `json:"cancel_at,omitempty"`
	CanceledAt           time.Time `json:"canceled_at,omitempty"`
	TrialEnd             time.Time `json:"trial_end,omitempty"`
	StripeSubscriptionID string    `json:"stripe_subscription_id,omitempty"`
	ManualOffline        bool      `json:"manual_offline"`
	CreatedAt            time.Time `json:"created_at"`
}

func subscriptionToDTO(s *store.Subscription) subscriptionDTO {
	return subscriptionDTO{
		ID:                   s.ID,
		Tier:                 s.Tier,
		Status:               s.Status,
		CurrentPeriodStart:   s.CurrentPeriodStart,
		CurrentPeriodEnd:     s.CurrentPeriodEnd,
		CancelAt:             s.CancelAt,
		CanceledAt:           s.CanceledAt,
		TrialEnd:             s.TrialEnd,
		StripeSubscriptionID: s.StripeSubscriptionID,
		ManualOffline:        s.ManualOffline,
		CreatedAt:            s.CreatedAt,
	}
}

func (s *Server) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	subs, err := s.cfg.Store.ListSubscriptionsByCustomer(r.Context(), cust.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	out := make([]subscriptionDTO, 0, len(subs))
	for _, sub := range subs {
		out = append(out, subscriptionToDTO(sub))
	}
	writeJSON(w, http.StatusOK, map[string]any{"subscriptions": out})
}

// ─── Deployments ────────────────────────────────────────────────────

type deploymentDTO struct {
	ID               string    `json:"id"`
	DeploymentID     string    `json:"deployment_id"`
	DeploymentName   string    `json:"deployment_name"`
	FirstSeenAt      time.Time `json:"first_seen_at"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	LastVersion      string    `json:"last_version"`
	FlaggedForReview bool      `json:"flagged_for_review"`
	FlagReason       string    `json:"flag_reason,omitempty"`
}

func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	deps, err := s.cfg.Store.ListDeploymentsByCustomer(r.Context(), cust.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	out := make([]deploymentDTO, 0, len(deps))
	for _, d := range deps {
		out = append(out, deploymentDTO{
			ID:               d.ID,
			DeploymentID:     d.DeploymentID,
			DeploymentName:   d.DeploymentName,
			FirstSeenAt:      d.FirstSeenAt,
			LastSeenAt:       d.LastSeenAt,
			LastVersion:      d.LastVersion,
			FlaggedForReview: d.FlaggedForReview,
			FlagReason:       d.FlagReason,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"deployments": out})
}

// ─── Enrollment tokens ──────────────────────────────────────────────

type enrollmentTokenDTO struct {
	ID                   string    `json:"id"`
	SubscriptionID       string    `json:"subscription_id"`
	Label                string    `json:"label,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at"`
	ConsumedAt           time.Time `json:"consumed_at,omitempty"`
	ConsumedByDeployment string    `json:"consumed_by_deployment,omitempty"`
}

func enrollmentTokenToDTO(t *store.EnrollmentToken) enrollmentTokenDTO {
	return enrollmentTokenDTO{
		ID:                   t.ID,
		SubscriptionID:       t.SubscriptionID,
		Label:                t.Label,
		CreatedAt:            t.CreatedAt,
		ExpiresAt:            t.ExpiresAt,
		ConsumedAt:           t.ConsumedAt,
		ConsumedByDeployment: t.ConsumedByDeployment,
	}
}

func (s *Server) handleListEnrollmentTokens(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	tokens, err := s.cfg.Store.ListEnrollmentTokensByCustomer(r.Context(), cust.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	out := make([]enrollmentTokenDTO, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, enrollmentTokenToDTO(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tokens": out})
}

type createTokenRequest struct {
	SubscriptionID string `json:"subscription_id"`
	Label          string `json:"label"`
	TTLHours       int    `json:"ttl_hours"`
}

type createTokenResponse struct {
	enrollmentTokenDTO
	Plaintext string `json:"plaintext"` // shown ONCE
}

func (s *Server) handleCreateEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	var req createTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.SubscriptionID == "" {
		writeError(w, http.StatusBadRequest, "subscription_id required")
		return
	}
	if req.TTLHours <= 0 {
		req.TTLHours = 24
	}
	if req.TTLHours > 24*30 {
		req.TTLHours = 24 * 30
	}

	cust := customerFromContext(r.Context())
	// Verify the subscription belongs to this customer.
	sub, err := s.cfg.Store.GetSubscription(r.Context(), req.SubscriptionID)
	if err != nil || sub.CustomerID != cust.ID {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	if cust.EmailVerifiedAt.IsZero() {
		writeError(w, http.StatusForbidden, "verify your email before creating enrollment tokens")
		return
	}

	plain, hash, err := newEnrollmentTokenPlaintext()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	now := s.cfg.Now()
	tok := &store.EnrollmentToken{
		ID:             "etok_" + plainID(),
		CustomerID:     cust.ID,
		SubscriptionID: req.SubscriptionID,
		TokenHash:      hash,
		Label:          req.Label,
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Duration(req.TTLHours) * time.Hour),
	}
	if err := s.cfg.Store.CreateEnrollmentToken(r.Context(), tok); err != nil {
		writeError(w, http.StatusInternalServerError, "create")
		return
	}
	s.appendAudit("portal.enrollment_token.created", cust, map[string]any{
		"token_id":        tok.ID,
		"subscription_id": tok.SubscriptionID,
		"label":           tok.Label,
	})
	writeJSON(w, http.StatusCreated, createTokenResponse{
		enrollmentTokenDTO: enrollmentTokenToDTO(tok),
		Plaintext:          plain,
	})
}

type revokeTokenRequest struct {
	TokenID string `json:"token_id"`
}

func (s *Server) handleRevokeEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	var req revokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	cust := customerFromContext(r.Context())
	// Confirm ownership: list tokens for customer + match id.
	tokens, err := s.cfg.Store.ListEnrollmentTokensByCustomer(r.Context(), cust.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	owned := false
	for _, t := range tokens {
		if t.ID == req.TokenID {
			owned = true
			break
		}
	}
	if !owned {
		writeError(w, http.StatusNotFound, "token not found")
		return
	}
	if err := s.cfg.Store.RevokeEnrollmentToken(r.Context(), req.TokenID); err != nil {
		writeError(w, http.StatusInternalServerError, "revoke")
		return
	}
	s.appendAudit("portal.enrollment_token.revoked", cust, map[string]any{"token_id": req.TokenID})
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Offline license download ───────────────────────────────────────
// Returns the most recent JWS the issuer minted for the given
// deployment. Lets a customer re-fetch a license they've already
// enrolled with — handy if their NodePulse install lost local state.

func (s *Server) handleOfflineLicense(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	licenseID := r.URL.Query().Get("license_id")
	if licenseID == "" {
		writeError(w, http.StatusBadRequest, "license_id required")
		return
	}
	lic, err := s.cfg.Store.GetLatestLicenseForLicenseID(r.Context(), licenseID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "license not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	if lic.CustomerID != cust.ID {
		writeError(w, http.StatusNotFound, "license not found")
		return
	}
	w.Header().Set("Content-Type", "application/jwt")
	w.Header().Set("Content-Disposition", `attachment; filename="`+licenseID+`.license.jws"`)
	_, _ = w.Write([]byte(lic.PayloadJWS))
}

func (s *Server) handleListLicenses(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	licenses, err := s.cfg.Store.ListLicensesByCustomer(r.Context(), cust.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	type licDTO struct {
		LicenseID    string    `json:"license_id"`
		JTI          string    `json:"jti"`
		DeploymentID string    `json:"deployment_id"`
		Tier         string    `json:"tier"`
		Iat          time.Time `json:"issued_at"`
		ExpiresAt    time.Time `json:"expires_at"`
		GraceUntil   time.Time `json:"grace_until"`
		Revoked      bool      `json:"revoked"`
		Kid          string    `json:"kid"`
	}
	// Deduplicate by license_id — only show the most recent issuance.
	seen := map[string]bool{}
	out := make([]licDTO, 0, len(licenses))
	for _, l := range licenses {
		if seen[l.LicenseID] {
			continue
		}
		seen[l.LicenseID] = true
		out = append(out, licDTO{
			LicenseID: l.LicenseID, JTI: l.JTI, DeploymentID: l.DeploymentID,
			Tier: l.Tier, Iat: l.Iat, ExpiresAt: l.ExpiresAt, GraceUntil: l.GraceUntil,
			Revoked: l.Revoked, Kid: l.Kid,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"licenses": out})
}

// ─── Account update ─────────────────────────────────────────────────

type updateAccountRequest struct {
	Name    *string `json:"name,omitempty"`
	Company *string `json:"company,omitempty"`
}

func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	var req updateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	cust := customerFromContext(r.Context())
	if req.Name != nil {
		cust.Name = strings.TrimSpace(*req.Name)
	}
	if req.Company != nil {
		cust.Company = strings.TrimSpace(*req.Company)
	}
	cust.UpdatedAt = s.cfg.Now()
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	s.appendAudit("portal.account_updated", cust, nil)
	writeJSON(w, http.StatusOK, customerProfile(cust))
}

// ─── Helpers ────────────────────────────────────────────────────────

// newEnrollmentTokenPlaintext returns ("NP-ENROLL-..." plaintext, sha256(plain)).
// Mirrors the issuer/tokens.go format so admin- and portal-issued
// tokens are indistinguishable to enrolling clients.
func newEnrollmentTokenPlaintext() (string, string, error) {
	// Borrow the same approach as issuer/tokens.go: 20 random bytes
	// → base32 → prefixed "NP-ENROLL-". SHA-256 hex over plaintext.
	plain, hashHex, err := generateEnrollmentTokenInternal()
	return plain, hashHex, err
}
