package portal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/email"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// ─── Signup ─────────────────────────────────────────────────────────
// A signup either:
//   - creates a brand-new Customer (no existing row by email), OR
//   - claims an existing Customer that has no PasswordHash yet
//     (created via admin/Stripe Checkout flows).
// In both cases an email-verification token is generated and emailed.

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Company  string `json:"company"`
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash error")
		return
	}

	now := s.cfg.Now()
	tokenPlain, tokenHash, err := newToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}

	existing, err := s.cfg.Store.GetCustomerByEmail(r.Context(), req.Email)
	switch {
	case err == nil && existing.PasswordHash != "":
		writeError(w, http.StatusConflict, "email already registered")
		return
	case err == nil:
		// Claim existing row.
		existing.PasswordHash = string(hash)
		existing.TokenHash = tokenHash
		existing.TokenExpires = now.Add(TokenTTL)
		existing.TokenPurpose = "verify_email"
		if req.Name != "" {
			existing.Name = req.Name
		}
		if req.Company != "" {
			existing.Company = req.Company
		}
		existing.UpdatedAt = now
		if err := s.cfg.Store.UpdateCustomer(r.Context(), existing); err != nil {
			writeError(w, http.StatusInternalServerError, "update customer")
			return
		}
		s.sendVerificationEmail(existing, tokenPlain)
		s.appendAudit("portal.signup_claimed", existing, nil)
		writeJSON(w, http.StatusOK, map[string]any{"status": "verification_sent"})
		return
	case errors.Is(err, store.ErrNotFound):
		// Fall through to create.
	default:
		writeError(w, http.StatusInternalServerError, "lookup customer")
		return
	}

	id, err := newOpaqueID("cust")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "id error")
		return
	}
	cust := &store.Customer{
		ID:           id,
		Email:        req.Email,
		Name:         req.Name,
		Company:      req.Company,
		Status:       "active",
		PasswordHash: string(hash),
		TokenHash:    tokenHash,
		TokenExpires: now.Add(TokenTTL),
		TokenPurpose: "verify_email",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.cfg.Store.CreateCustomer(r.Context(), cust); err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "create customer")
		return
	}
	s.sendVerificationEmail(cust, tokenPlain)
	s.appendAudit("portal.signup_created", cust, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "verification_sent"})
}

// ─── Login ──────────────────────────────────────────────────────────

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	cust, err := s.cfg.Store.GetCustomerByEmail(r.Context(), req.Email)
	if err != nil || cust.PasswordHash == "" {
		// Run a dummy bcrypt to keep timing consistent.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$dummy.hash.to.equalize.timing.aaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), []byte(req.Password))
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cust.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if cust.Status != "active" {
		writeError(w, http.StatusForbidden, "account "+cust.Status)
		return
	}

	now := s.cfg.Now()
	sid, err := newSessionID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}
	sess := &store.PortalSession{
		ID:         sid,
		CustomerID: cust.ID,
		IP:         clientIP(r),
		UserAgent:  r.UserAgent(),
		CreatedAt:  now,
		LastSeen:   now,
		ExpiresAt:  now.Add(SessionTTL),
	}
	if err := s.cfg.Store.CreatePortalSession(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "create session")
		return
	}
	s.setSessionCookie(w, sid, sess.ExpiresAt)
	s.appendAudit("portal.login", cust, nil)
	writeJSON(w, http.StatusOK, customerProfile(cust))
}

// ─── Logout ─────────────────────────────────────────────────────────

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromContext(r.Context())
	_ = s.cfg.Store.DeletePortalSession(r.Context(), sess.ID)
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Me ─────────────────────────────────────────────────────────────

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	writeJSON(w, http.StatusOK, customerProfile(cust))
}

// ─── Change password ────────────────────────────────────────────────

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "new password must be at least 8 characters")
		return
	}
	cust := customerFromContext(r.Context())
	if err := bcrypt.CompareHashAndPassword([]byte(cust.PasswordHash), []byte(req.OldPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "old password incorrect")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash error")
		return
	}
	cust.PasswordHash = string(hash)
	cust.UpdatedAt = s.cfg.Now()
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	// Invalidate all other sessions; keep this one.
	thisSession := sessionFromContext(r.Context())
	_ = s.cfg.Store.DeletePortalSessionsForCustomer(r.Context(), cust.ID)
	// Re-create the current session so the user stays logged in.
	thisSession.CreatedAt = s.cfg.Now()
	thisSession.LastSeen = thisSession.CreatedAt
	thisSession.ExpiresAt = thisSession.CreatedAt.Add(SessionTTL)
	_ = s.cfg.Store.CreatePortalSession(r.Context(), thisSession)
	s.setSessionCookie(w, thisSession.ID, thisSession.ExpiresAt)
	s.appendAudit("portal.password_changed", cust, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Verify email ───────────────────────────────────────────────────

type verifyEmailRequest struct {
	Token string `json:"token"`
}

func (s *Server) handleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	cust, err := s.findByToken(r.Context(), req.Token, "verify_email")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	now := s.cfg.Now()
	cust.EmailVerifiedAt = now
	cust.TokenHash = ""
	cust.TokenExpires = time.Time{}
	cust.TokenPurpose = ""
	cust.UpdatedAt = now
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	s.appendAudit("portal.email_verified", cust, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleResendVerification(w http.ResponseWriter, r *http.Request) {
	cust := customerFromContext(r.Context())
	if !cust.EmailVerifiedAt.IsZero() {
		writeError(w, http.StatusBadRequest, "already verified")
		return
	}
	now := s.cfg.Now()
	tokenPlain, tokenHash, err := newToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token error")
		return
	}
	cust.TokenHash = tokenHash
	cust.TokenExpires = now.Add(TokenTTL)
	cust.TokenPurpose = "verify_email"
	cust.UpdatedAt = now
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	s.sendVerificationEmail(cust, tokenPlain)
	writeJSON(w, http.StatusOK, map[string]any{"status": "verification_sent"})
}

// ─── Password reset ─────────────────────────────────────────────────

type requestResetRequest struct {
	Email string `json:"email"`
}

func (s *Server) handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req requestResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	// Always respond 200 — don't leak which emails exist.
	defer writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})

	cust, err := s.cfg.Store.GetCustomerByEmail(r.Context(), req.Email)
	if err != nil || cust.PasswordHash == "" {
		return
	}
	now := s.cfg.Now()
	tokenPlain, tokenHash, err := newToken()
	if err != nil {
		return
	}
	cust.TokenHash = tokenHash
	cust.TokenExpires = now.Add(TokenTTL)
	cust.TokenPurpose = "reset_password"
	cust.UpdatedAt = now
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		return
	}
	s.sendPasswordResetEmail(cust, tokenPlain)
	s.appendAudit("portal.password_reset_requested", cust, nil)
}

type resetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (s *Server) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	cust, err := s.findByToken(r.Context(), req.Token, "reset_password")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid or expired token")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash error")
		return
	}
	now := s.cfg.Now()
	cust.PasswordHash = string(hash)
	cust.TokenHash = ""
	cust.TokenExpires = time.Time{}
	cust.TokenPurpose = ""
	cust.UpdatedAt = now
	if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	// Invalidate any active sessions — force re-login.
	_ = s.cfg.Store.DeletePortalSessionsForCustomer(r.Context(), cust.ID)
	s.appendAudit("portal.password_reset", cust, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Helpers ────────────────────────────────────────────────────────

// findByToken looks up a customer by the SHA-256 hash of a plaintext
// token, verifies purpose, and verifies the token hasn't expired.
func (s *Server) findByToken(ctx context.Context, plain, expectedPurpose string) (*store.Customer, error) {
	if plain == "" {
		return nil, errors.New("empty token")
	}
	cust, err := s.cfg.Store.GetCustomerByTokenHash(ctx, hashToken(plain))
	if err != nil {
		return nil, err
	}
	if cust.TokenPurpose != expectedPurpose {
		return nil, errors.New("wrong token purpose")
	}
	if cust.TokenExpires.IsZero() || s.cfg.Now().After(cust.TokenExpires) {
		return nil, errors.New("token expired")
	}
	return cust, nil
}

func (s *Server) sendVerificationEmail(cust *store.Customer, tokenPlain string) {
	if s.cfg.Email == nil {
		return
	}
	link := s.cfg.AppURL + "/verify-email?token=" + tokenPlain
	body := "Click to verify your infraYS account: " + link + "\n\nThis link expires in 24 hours."
	html := `<p>Click to verify your infraYS account: <a href="` + link + `">` + link + `</a></p><p>This link expires in 24 hours.</p>`
	_ = s.cfg.Email.Send(context.Background(), email.Message{
		To:          cust.Email,
		Subject:     "Verify your infraYS account",
		TextBody:    body,
		HTMLBody:    html,
		MessageType: "portal_verify_email",
	})
}

func (s *Server) sendPasswordResetEmail(cust *store.Customer, tokenPlain string) {
	if s.cfg.Email == nil {
		return
	}
	link := s.cfg.AppURL + "/reset-password?token=" + tokenPlain
	body := "Click to reset your infraYS password: " + link + "\n\nThis link expires in 24 hours. If you didn't request this, ignore this email."
	html := `<p>Click to reset your infraYS password: <a href="` + link + `">` + link + `</a></p><p>This link expires in 24 hours. If you didn't request this, ignore this email.</p>`
	_ = s.cfg.Email.Send(context.Background(), email.Message{
		To:          cust.Email,
		Subject:     "Reset your infraYS password",
		TextBody:    body,
		HTMLBody:    html,
		MessageType: "portal_password_reset",
	})
}

func (s *Server) appendAudit(eventType string, cust *store.Customer, payload map[string]any) {
	if s.cfg.Audit == nil {
		return
	}
	_, _ = s.cfg.Audit.Append(context.Background(), audit.Entry{
		EventType:  eventType,
		CustomerID: cust.ID,
		Actor:      cust.Email,
		Payload:    payload,
		CreatedAt:  s.cfg.Now(),
	})
}

// hashToken returns hex(sha256(plain)).
func hashToken(plain string) string {
	h := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(h[:])
}

// newToken returns (plaintext, sha256_hex). Plaintext is 32 random
// bytes hex-encoded — 256 bits of entropy, URL-safe.
func newToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	plain := hex.EncodeToString(buf)
	return plain, hashToken(plain), nil
}

// newOpaqueID returns "<prefix>_<32 hex>".
func newOpaqueID(prefix string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buf), nil
}

// customerProfile is the safe shape we return to the browser — no
// password hash, no token state.
type CustomerProfile struct {
	ID             string    `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Company        string    `json:"company"`
	Status         string    `json:"status"`
	EmailVerified  bool      `json:"email_verified"`
	CreatedAt      time.Time `json:"created_at"`
}

func customerProfile(c *store.Customer) CustomerProfile {
	return CustomerProfile{
		ID:            c.ID,
		Email:         c.Email,
		Name:          c.Name,
		Company:       c.Company,
		Status:        c.Status,
		EmailVerified: !c.EmailVerifiedAt.IsZero(),
		CreatedAt:     c.CreatedAt,
	}
}

func clientIP(r *http.Request) string {
	if v := r.Header.Get("X-Forwarded-For"); v != "" {
		if i := strings.IndexByte(v, ','); i >= 0 {
			return strings.TrimSpace(v[:i])
		}
		return strings.TrimSpace(v)
	}
	return r.RemoteAddr
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
