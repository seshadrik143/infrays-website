package adminportal

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// ─── Customers ──────────────────────────────────────────────────────

type customerDTO struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	Company         string    `json:"company"`
	Status          string    `json:"status"`
	StripeCustomerID string   `json:"stripe_customer_id,omitempty"`
	EmailVerified   bool      `json:"email_verified"`
	CreatedAt       time.Time `json:"created_at"`
}

func customerToDTO(c *store.Customer) customerDTO {
	return customerDTO{
		ID: c.ID, Email: c.Email, Name: c.Name, Company: c.Company,
		Status: c.Status, StripeCustomerID: c.StripeCustomerID,
		EmailVerified: !c.EmailVerifiedAt.IsZero(),
		CreatedAt:     c.CreatedAt,
	}
}

func (s *Server) handleListCustomers(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("q")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	customers, err := s.cfg.Store.ListCustomers(r.Context(), filter, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	out := make([]customerDTO, 0, len(customers))
	for _, c := range customers {
		out = append(out, customerToDTO(c))
	}
	writeJSON(w, http.StatusOK, map[string]any{"customers": out})
}

type customerDetailResponse struct {
	Customer         customerDTO              `json:"customer"`
	Subscriptions    []subscriptionDTO        `json:"subscriptions"`
	Deployments      []deploymentDTO          `json:"deployments"`
	EnrollmentTokens []enrollmentTokenDTO     `json:"enrollment_tokens"`
	Licenses         []licenseDTO             `json:"licenses"`
}

func (s *Server) handleGetCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cust, err := s.cfg.Store.GetCustomer(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	subs, _ := s.cfg.Store.ListSubscriptionsByCustomer(r.Context(), id)
	deps, _ := s.cfg.Store.ListDeploymentsByCustomer(r.Context(), id)
	tokens, _ := s.cfg.Store.ListEnrollmentTokensByCustomer(r.Context(), id)
	licenses, _ := s.cfg.Store.ListLicensesByCustomer(r.Context(), id)

	resp := customerDetailResponse{Customer: customerToDTO(cust)}
	for _, x := range subs {
		resp.Subscriptions = append(resp.Subscriptions, subToDTO(x))
	}
	for _, x := range deps {
		resp.Deployments = append(resp.Deployments, depToDTO(x))
	}
	for _, x := range tokens {
		resp.EnrollmentTokens = append(resp.EnrollmentTokens, etokToDTO(x))
	}
	// Dedup licenses by LicenseID, newest first.
	seen := map[string]bool{}
	for _, x := range licenses {
		if seen[x.LicenseID] {
			continue
		}
		seen[x.LicenseID] = true
		resp.Licenses = append(resp.Licenses, licToDTO(x))
	}
	writeJSON(w, http.StatusOK, resp)
}

type updateCustomerRequest struct {
	Status *string `json:"status,omitempty"`  // active | suspended | deleted
	Name   *string `json:"name,omitempty"`
	Company *string `json:"company,omitempty"`
}

func (s *Server) handleUpdateCustomer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cust, err := s.cfg.Store.GetCustomer(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	var req updateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Status != nil {
		switch *req.Status {
		case "active", "suspended", "deleted":
		default:
			writeError(w, http.StatusBadRequest, "invalid status")
			return
		}
		cust.Status = *req.Status
		// Suspending: also yank portal sessions so they can't keep
		// logging in until reactivated.
		if *req.Status != "active" {
			_ = s.cfg.Store.DeletePortalSessionsForCustomer(r.Context(), cust.ID)
		}
	}
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
	s.appendAudit("admin.customer_updated", adminFromContext(r.Context()), map[string]any{
		"customer_id": cust.ID, "status": cust.Status,
	})
	writeJSON(w, http.StatusOK, customerToDTO(cust))
}

// ─── Offline subscription creation ──────────────────────────────────

type createSubReq struct {
	Tier             string    `json:"tier"`              // free | professional | enterprise
	EntitlementSetID string    `json:"entitlement_set_id"`
	CurrentPeriodEnd time.Time `json:"current_period_end"`
	TrialEnd         time.Time `json:"trial_end,omitempty"`
}

func (s *Server) handleCreateOfflineSubscription(w http.ResponseWriter, r *http.Request) {
	custID := r.PathValue("id")
	if _, err := s.cfg.Store.GetCustomer(r.Context(), custID); err != nil {
		writeError(w, http.StatusNotFound, "customer not found")
		return
	}
	var req createSubReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Tier == "" {
		writeError(w, http.StatusBadRequest, "tier required")
		return
	}
	now := s.cfg.Now()
	sub := &store.Subscription{
		ID:                 "sub_" + opaqueID(),
		CustomerID:         custID,
		Tier:               req.Tier,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   req.CurrentPeriodEnd,
		TrialEnd:           req.TrialEnd,
		ManualOffline:      true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if !req.TrialEnd.IsZero() && req.TrialEnd.After(now) {
		sub.Status = "trialing"
	}
	if err := s.cfg.Store.CreateSubscription(r.Context(), sub); err != nil {
		writeError(w, http.StatusInternalServerError, "create")
		return
	}
	s.appendAudit("admin.offline_subscription_created", adminFromContext(r.Context()), map[string]any{
		"customer_id": custID, "subscription_id": sub.ID, "tier": sub.Tier,
	})
	writeJSON(w, http.StatusCreated, subToDTO(sub))
}

// ─── Admin-issued enrollment token ──────────────────────────────────

type createEnrollReq struct {
	SubscriptionID string `json:"subscription_id"`
	Label          string `json:"label"`
	TTLHours       int    `json:"ttl_hours"`
}

func (s *Server) handleAdminCreateEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	custID := r.PathValue("id")
	var req createEnrollReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.TTLHours <= 0 {
		req.TTLHours = 24
	}
	sub, err := s.cfg.Store.GetSubscription(r.Context(), req.SubscriptionID)
	if err != nil || sub.CustomerID != custID {
		writeError(w, http.StatusNotFound, "subscription not found")
		return
	}
	plain, hash, err := newEnrollmentToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token")
		return
	}
	now := s.cfg.Now()
	tok := &store.EnrollmentToken{
		ID:             "etok_" + opaqueID(),
		CustomerID:     custID,
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
	s.appendAudit("admin.enrollment_token_issued", adminFromContext(r.Context()), map[string]any{
		"customer_id": custID, "token_id": tok.ID,
	})
	writeJSON(w, http.StatusCreated, map[string]any{
		"token":     etokToDTO(tok),
		"plaintext": plain,
	})
}

// ─── Deployment flag toggle ─────────────────────────────────────────

type flagReq struct {
	Flagged bool   `json:"flagged"`
	Reason  string `json:"reason"`
}

func (s *Server) handleFlagDeployment(w http.ResponseWriter, r *http.Request) {
	depID := r.PathValue("id")
	dep, err := s.cfg.Store.GetDeployment(r.Context(), depID)
	if err != nil {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	var req flagReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	dep.FlaggedForReview = req.Flagged
	dep.FlagReason = req.Reason
	if err := s.cfg.Store.UpdateDeployment(r.Context(), dep); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	s.appendAudit("admin.deployment_flag_set", adminFromContext(r.Context()), map[string]any{
		"deployment_id": depID, "flagged": req.Flagged, "reason": req.Reason,
	})
	writeJSON(w, http.StatusOK, depToDTO(dep))
}

// ─── License revocation ─────────────────────────────────────────────

type revokeReq struct {
	Reason string `json:"reason"`
}

func (s *Server) handleRevokeLicense(w http.ResponseWriter, r *http.Request) {
	jti := r.PathValue("jti")
	var req revokeReq
	_ = json.NewDecoder(r.Body).Decode(&req)
	if err := s.cfg.Store.RevokeLicenseByJTI(r.Context(), jti, req.Reason); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "license not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "revoke")
		return
	}
	s.appendAudit("admin.license_revoked", adminFromContext(r.Context()), map[string]any{
		"jti": jti, "reason": req.Reason,
	})
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Audit log ──────────────────────────────────────────────────────

func (s *Server) handleListAudit(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	entries, err := s.cfg.Audit.Tail(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "audit")
		return
	}
	type entryDTO struct {
		Seq        uint64    `json:"seq"`
		Hash       string    `json:"hash"`
		EventType  string    `json:"event_type"`
		CustomerID string    `json:"customer_id,omitempty"`
		Actor      string    `json:"actor"`
		Payload    map[string]any `json:"payload,omitempty"`
		CreatedAt  time.Time `json:"created_at"`
	}
	out := make([]entryDTO, 0, len(entries))
	for _, e := range entries {
		out = append(out, entryDTO{
			Seq: e.Seq, Hash: e.Hash, EventType: e.EventType,
			CustomerID: e.CustomerID, Actor: e.Actor,
			Payload: e.Payload, CreatedAt: e.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": out})
}

// ─── DTOs reused across endpoints ───────────────────────────────────

type subscriptionDTO struct {
	ID                 string    `json:"id"`
	Tier               string    `json:"tier"`
	Status             string    `json:"status"`
	StripeSubscriptionID string  `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	TrialEnd           time.Time `json:"trial_end,omitempty"`
	ManualOffline      bool      `json:"manual_offline"`
	CreatedAt          time.Time `json:"created_at"`
}

func subToDTO(s *store.Subscription) subscriptionDTO {
	return subscriptionDTO{
		ID: s.ID, Tier: s.Tier, Status: s.Status,
		StripeSubscriptionID: s.StripeSubscriptionID,
		CurrentPeriodEnd:     s.CurrentPeriodEnd,
		TrialEnd:             s.TrialEnd, ManualOffline: s.ManualOffline,
		CreatedAt: s.CreatedAt,
	}
}

type deploymentDTO struct {
	ID               string    `json:"id"`
	DeploymentID     string    `json:"deployment_id"`
	DeploymentName   string    `json:"deployment_name"`
	LastSeenAt       time.Time `json:"last_seen_at"`
	LastVersion      string    `json:"last_version"`
	FlaggedForReview bool      `json:"flagged_for_review"`
	FlagReason       string    `json:"flag_reason,omitempty"`
}

func depToDTO(d *store.Deployment) deploymentDTO {
	return deploymentDTO{
		ID: d.ID, DeploymentID: d.DeploymentID,
		DeploymentName: d.DeploymentName, LastSeenAt: d.LastSeenAt,
		LastVersion: d.LastVersion, FlaggedForReview: d.FlaggedForReview,
		FlagReason: d.FlagReason,
	}
}

type enrollmentTokenDTO struct {
	ID                   string    `json:"id"`
	SubscriptionID       string    `json:"subscription_id"`
	Label                string    `json:"label,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	ExpiresAt            time.Time `json:"expires_at"`
	ConsumedAt           time.Time `json:"consumed_at,omitempty"`
	ConsumedByDeployment string    `json:"consumed_by_deployment,omitempty"`
}

func etokToDTO(t *store.EnrollmentToken) enrollmentTokenDTO {
	return enrollmentTokenDTO{
		ID: t.ID, SubscriptionID: t.SubscriptionID, Label: t.Label,
		CreatedAt: t.CreatedAt, ExpiresAt: t.ExpiresAt,
		ConsumedAt: t.ConsumedAt, ConsumedByDeployment: t.ConsumedByDeployment,
	}
}

type licenseDTO struct {
	LicenseID    string    `json:"license_id"`
	JTI          string    `json:"jti"`
	DeploymentID string    `json:"deployment_id"`
	Tier         string    `json:"tier"`
	ExpiresAt    time.Time `json:"expires_at"`
	Revoked      bool      `json:"revoked"`
	Kid          string    `json:"kid"`
}

func licToDTO(l *store.License) licenseDTO {
	return licenseDTO{
		LicenseID: l.LicenseID, JTI: l.JTI, DeploymentID: l.DeploymentID,
		Tier: l.Tier, ExpiresAt: l.ExpiresAt, Revoked: l.Revoked, Kid: l.Kid,
	}
}

// ─── Helpers ────────────────────────────────────────────────────────

func opaqueID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func newEnrollmentToken() (plain, hash string, err error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	body := strings.TrimRight(base32.StdEncoding.EncodeToString(raw), "=")
	plain = "NP-ENROLL-" + body
	sum := sha256.Sum256([]byte(plain))
	hash = hex.EncodeToString(sum[:])
	return plain, hash, nil
}

// keep audit import used; harmless on no-op.
var _ = audit.Entry{}
