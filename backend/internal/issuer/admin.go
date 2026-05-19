package issuer

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// Admin endpoints are gated behind a shared-secret header for
// Phase 49. Phase 52.5 replaces this with real session+RBAC.
//
// Env: NP_ISSUER_ADMIN_SECRET. Set on the issuer process; sent by
// the admin client as X-Admin-Secret.
const adminSecretEnv = "NP_ISSUER_ADMIN_SECRET"

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	want := os.Getenv(adminSecretEnv)
	got := r.Header.Get("X-Admin-Secret")
	if want == "" || got == "" || got != want {
		writeErr(w, http.StatusForbidden, "admin_only", "admin authentication required")
		return false
	}
	return true
}

// ── /internal/admin/customers ──────────────────────────────────────

type adminCreateCustomerReq struct {
	Email   string `json:"email"`
	Name    string `json:"name,omitempty"`
	Company string `json:"company,omitempty"`
}

func (s *Server) handleAdminCreateCustomer(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
	var req adminCreateCustomerReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "email is required")
		return
	}
	now := s.cfg.Now()
	c := &store.Customer{
		ID:        newOpaqueID("cust"),
		Email:     req.Email,
		Name:      req.Name,
		Company:   req.Company,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.cfg.Store.CreateCustomer(r.Context(), c); err != nil {
		writeErr(w, http.StatusConflict, "already_exists", err.Error())
		return
	}
	_, _ = s.cfg.Audit.Append(r.Context(), audit.Entry{
		EventType:  "admin.customer.created",
		CustomerID: c.ID,
		Actor:      "admin",
		Payload:    map[string]any{"email": c.Email},
	})
	writeJSON(w, http.StatusCreated, c)
}

// ── /internal/admin/subscriptions ──────────────────────────────────

type adminCreateSubscriptionReq struct {
	CustomerID            string `json:"customer_id"`
	Tier                  string `json:"tier"`
	StripeSubscriptionID  string `json:"stripe_subscription_id,omitempty"`
	StripePriceID         string `json:"stripe_price_id,omitempty"`
	Status                string `json:"status,omitempty"` // default "active"
	CurrentPeriodEndDays  int    `json:"current_period_end_days,omitempty"` // default 365
	TrialEndDays          int    `json:"trial_end_days,omitempty"`
	ManualOffline         bool   `json:"manual_offline,omitempty"`
}

func (s *Server) handleAdminCreateSubscription(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
	var req adminCreateSubscriptionReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.CustomerID == "" || req.Tier == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "customer_id and tier are required")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.CurrentPeriodEndDays == 0 {
		req.CurrentPeriodEndDays = 365
	}
	now := s.cfg.Now()
	sub := &store.Subscription{
		ID:                   newOpaqueID("sub"),
		CustomerID:           req.CustomerID,
		Tier:                 req.Tier,
		StripeSubscriptionID: req.StripeSubscriptionID,
		StripePriceID:        req.StripePriceID,
		Status:               req.Status,
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     now.Add(time.Duration(req.CurrentPeriodEndDays) * 24 * time.Hour),
		ManualOffline:        req.ManualOffline,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if req.TrialEndDays > 0 {
		sub.TrialEnd = now.Add(time.Duration(req.TrialEndDays) * 24 * time.Hour)
		sub.Status = "trialing"
	}
	if err := s.cfg.Store.CreateSubscription(r.Context(), sub); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	_, _ = s.cfg.Audit.Append(r.Context(), audit.Entry{
		EventType:      "admin.subscription.created",
		CustomerID:     sub.CustomerID,
		SubscriptionID: sub.ID,
		Actor:          "admin",
		Payload: map[string]any{
			"tier":             sub.Tier,
			"manual_offline":   sub.ManualOffline,
			"current_period_end": sub.CurrentPeriodEnd.Format(time.RFC3339),
		},
	})
	writeJSON(w, http.StatusCreated, sub)
}

// ── /internal/admin/enrollment-tokens ──────────────────────────────

type adminCreateEnrollmentTokenReq struct {
	CustomerID     string `json:"customer_id"`
	SubscriptionID string `json:"subscription_id"`
	Label          string `json:"label,omitempty"`
	TTLHours       int    `json:"ttl_hours,omitempty"` // default 24
}

type adminCreateEnrollmentTokenResp struct {
	TokenID    string    `json:"token_id"`
	Plaintext  string    `json:"plaintext"` // shown ONCE
	Label      string    `json:"label,omitempty"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func (s *Server) handleAdminCreateEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
	var req adminCreateEnrollmentTokenReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.CustomerID == "" || req.SubscriptionID == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "customer_id and subscription_id are required")
		return
	}
	if req.TTLHours == 0 {
		req.TTLHours = 24
	}
	plaintext, hash, err := generateEnrollmentToken()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	now := s.cfg.Now()
	tok := &store.EnrollmentToken{
		ID:             newOpaqueID("etok"),
		CustomerID:     req.CustomerID,
		SubscriptionID: req.SubscriptionID,
		TokenHash:      hash,
		Label:          req.Label,
		CreatedAt:      now,
		ExpiresAt:      now.Add(time.Duration(req.TTLHours) * time.Hour),
	}
	if err := s.cfg.Store.CreateEnrollmentToken(r.Context(), tok); err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	_, _ = s.cfg.Audit.Append(r.Context(), audit.Entry{
		EventType:      "admin.enrollment_token.created",
		CustomerID:     tok.CustomerID,
		SubscriptionID: tok.SubscriptionID,
		Actor:          "admin",
		Payload: map[string]any{
			"label":      tok.Label,
			"ttl_hours":  req.TTLHours,
			"expires_at": tok.ExpiresAt.Format(time.RFC3339),
		},
	})
	writeJSON(w, http.StatusCreated, adminCreateEnrollmentTokenResp{
		TokenID:   tok.ID,
		Plaintext: plaintext,
		Label:     tok.Label,
		ExpiresAt: tok.ExpiresAt,
	})
}

// ── /internal/admin/licenses/revoke ────────────────────────────────

type adminRevokeLicenseReq struct {
	JTI    string `json:"jti"`
	Reason string `json:"reason"`
}

func (s *Server) handleAdminRevokeLicense(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 4096))
	var req adminRevokeLicenseReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.JTI == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "jti is required")
		return
	}
	if err := s.cfg.Store.RevokeLicenseByJTI(r.Context(), req.JTI, req.Reason); err != nil {
		writeErr(w, http.StatusNotFound, "not_found", err.Error())
		return
	}
	_, _ = s.cfg.Audit.Append(r.Context(), audit.Entry{
		EventType: "license.revoked",
		Actor:     "admin",
		Payload:   map[string]any{"jti": req.JTI, "reason": req.Reason},
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "jti": req.JTI})
}

// ── /internal/admin/entitlement-sets ───────────────────────────────

type adminCreateEntitlementSetReq struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Version  int      `json:"version"`
	Features []string `json:"features"`
	Limits   struct {
		MaxAgents        int64 `json:"max_agents"`
		MaxMetricsPerSec int64 `json:"max_metrics_per_sec"`
		MaxLogGBPerDay   int64 `json:"max_log_gb_per_day"`
		MaxAlertRules    int64 `json:"max_alert_rules"`
		RetentionDays    int   `json:"retention_days"`
	} `json:"limits"`
}

func (s *Server) handleAdminCreateEntitlementSet(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	body, _ := io.ReadAll(io.LimitReader(r.Body, 8192))
	var req adminCreateEntitlementSetReq
	if err := json.Unmarshal(body, &req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if req.ID == "" || req.Name == "" {
		writeErr(w, http.StatusBadRequest, "missing_fields", "id and name are required")
		return
	}
	es := &store.EntitlementSet{
		ID:       req.ID,
		Name:     req.Name,
		Version:  req.Version,
		Features: req.Features,
		Limits: store.Limits{
			MaxAgents:        req.Limits.MaxAgents,
			MaxMetricsPerSec: req.Limits.MaxMetricsPerSec,
			MaxLogGBPerDay:   req.Limits.MaxLogGBPerDay,
			MaxAlertRules:    req.Limits.MaxAlertRules,
			RetentionDays:    req.Limits.RetentionDays,
		},
		CreatedAt: s.cfg.Now(),
	}
	if err := s.cfg.Store.CreateEntitlementSet(r.Context(), es); err != nil {
		writeErr(w, http.StatusConflict, "already_exists", err.Error())
		return
	}
	_, _ = s.cfg.Audit.Append(r.Context(), audit.Entry{
		EventType: "admin.entitlement_set.created",
		Actor:     "admin",
		Payload:   map[string]any{"id": es.ID, "name": es.Name, "version": es.Version},
	})
	writeJSON(w, http.StatusCreated, es)
}

// ── /internal/admin/audit ──────────────────────────────────────────

func (s *Server) handleAdminAuditTail(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	n := 100
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
			if n > 1000 {
				n = 1000
			}
		}
	}
	tail, err := s.cfg.Audit.Tail(r.Context(), n)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "internal", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"count":   s.cfg.Audit.Count(r.Context()),
		"entries": tail,
	})
}
