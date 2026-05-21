package adminportal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/seshadrik143/infrays-website/backend/internal/audit"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

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
	admin, err := s.cfg.Store.GetAdminUserByEmail(r.Context(), req.Email)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$dummy.hash.to.equalize.timing.aaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), []byte(req.Password))
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	now := s.cfg.Now()
	sid, err := newSessionID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session error")
		return
	}
	sess := &store.AdminSession{
		ID:          sid,
		AdminUserID: admin.ID,
		IP:          clientIP(r),
		UserAgent:   r.UserAgent(),
		MFAVerified: false,
		CreatedAt:   now,
		LastSeen:    now,
		ExpiresAt:   now.Add(SessionIdleTTL),
	}
	if err := s.cfg.Store.CreateAdminSession(r.Context(), sess); err != nil {
		writeError(w, http.StatusInternalServerError, "create session")
		return
	}
	s.setSessionCookie(w, sid, sess.ExpiresAt)
	s.appendAudit("admin.login_stage1", admin, nil)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           admin.ID,
		"email":        admin.Email,
		"role":         admin.Role,
		"mfa_enrolled": admin.MFAEnrolled,
		"mfa_required": true,
	})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	sess := sessionFromContext(r.Context())
	_ = s.cfg.Store.DeleteAdminSession(r.Context(), sess.ID)
	s.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// AdminProfile is the safe shape returned to the SPA.
type AdminProfile struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	MFAEnrolled bool      `json:"mfa_enrolled"`
	MFAVerified bool      `json:"mfa_verified"`
	LastLogin   time.Time `json:"last_login,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	admin := adminFromContext(r.Context())
	sess := sessionFromContext(r.Context())
	writeJSON(w, http.StatusOK, AdminProfile{
		ID: admin.ID, Email: admin.Email, Role: admin.Role,
		MFAEnrolled: admin.MFAEnrolled, MFAVerified: sess.MFAVerified,
		LastLogin: admin.LastLogin, CreatedAt: admin.CreatedAt,
	})
}

// ─── MFA setup ──────────────────────────────────────────────────────

type mfaSetupResponse struct {
	Secret      string `json:"secret"`       // base32 — manual entry fallback
	OtpauthURL  string `json:"otpauth_url"`  // for QR encoding on the SPA side
}

func (s *Server) handleMFASetup(w http.ResponseWriter, r *http.Request) {
	admin := adminFromContext(r.Context())
	if admin.MFAEnrolled {
		writeError(w, http.StatusConflict, "already enrolled")
		return
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.TOTPIssuer,
		AccountName: admin.Email,
		SecretSize:  20,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "totp generate")
		return
	}
	// Stash the secret on the admin row — finalized only after
	// successful verify. If the admin abandons setup we leave it
	// dangling; the next setup call simply overwrites.
	admin.MFASecret = key.Secret()
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), admin); err != nil {
		writeError(w, http.StatusInternalServerError, "store secret")
		return
	}
	writeJSON(w, http.StatusOK, mfaSetupResponse{
		Secret:     key.Secret(),
		OtpauthURL: key.URL(),
	})
}

type codeRequest struct {
	Code string `json:"code"`
}

func (s *Server) handleMFASetupVerify(w http.ResponseWriter, r *http.Request) {
	admin := adminFromContext(r.Context())
	if admin.MFAEnrolled {
		writeError(w, http.StatusConflict, "already enrolled")
		return
	}
	if admin.MFASecret == "" {
		writeError(w, http.StatusBadRequest, "no setup in progress")
		return
	}
	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if !s.validateTOTP(req.Code, admin.MFASecret) {
		writeError(w, http.StatusUnauthorized, "invalid code")
		return
	}
	admin.MFAEnrolled = true
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), admin); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	// Flip the current session to verified — first-login flow shouldn't
	// require a separate challenge step right after enrollment.
	sess := sessionFromContext(r.Context())
	_ = s.cfg.Store.MarkAdminSessionMFAVerified(r.Context(), sess.ID)
	s.appendAudit("admin.mfa_enrolled", admin, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── MFA challenge (regular login flow) ─────────────────────────────

func (s *Server) handleMFAChallenge(w http.ResponseWriter, r *http.Request) {
	admin := adminFromContext(r.Context())
	sess := sessionFromContext(r.Context())
	if sess.MFAVerified {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}
	if !admin.MFAEnrolled {
		writeError(w, http.StatusBadRequest, "mfa not enrolled")
		return
	}
	var req codeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if !s.validateTOTP(req.Code, admin.MFASecret) {
		writeError(w, http.StatusUnauthorized, "invalid code")
		return
	}
	_ = s.cfg.Store.MarkAdminSessionMFAVerified(r.Context(), sess.ID)
	admin.LastLogin = s.cfg.Now()
	_ = s.cfg.Store.UpdateAdminUser(r.Context(), admin)
	s.appendAudit("admin.login_mfa_verified", admin, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
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
	if len(req.NewPassword) < 10 {
		writeError(w, http.StatusBadRequest, "new password must be at least 10 characters")
		return
	}
	admin := adminFromContext(r.Context())
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		writeError(w, http.StatusUnauthorized, "old password incorrect")
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash")
		return
	}
	admin.PasswordHash = string(hash)
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), admin); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	// Kill every other session.
	thisSession := sessionFromContext(r.Context())
	_ = s.cfg.Store.DeleteAdminSessionsForUser(r.Context(), admin.ID)
	_ = s.cfg.Store.CreateAdminSession(r.Context(), thisSession)
	s.appendAudit("admin.password_changed", admin, nil)
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Helpers ────────────────────────────────────────────────────────

func (s *Server) appendAudit(eventType string, admin *store.AdminUser, payload map[string]any) {
	if s.cfg.Audit == nil {
		return
	}
	_, _ = s.cfg.Audit.Append(context.Background(), audit.Entry{
		EventType: eventType,
		Actor:     admin.Email,
		Payload:   payload,
		CreatedAt: s.cfg.Now(),
	})
}

// validateTOTP wraps totp.ValidateCustom so the injected clock flows
// through during tests. Default opts (1 period, 6 digits, SHA1).
func (s *Server) validateTOTP(code, secret string) bool {
	ok, _ := totp.ValidateCustom(code, secret, s.cfg.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    6,
		Algorithm: 0, // SHA1 default
	})
	return ok
}

// HashPassword bcrypts the password; exposed so the issuer's bootstrap
// flow (env-based) can hash without importing bcrypt directly.
func HashPassword(plaintext string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// NewAdminID returns "admin_<16 hex>".
func NewAdminID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "admin_" + hex.EncodeToString(b)
}
