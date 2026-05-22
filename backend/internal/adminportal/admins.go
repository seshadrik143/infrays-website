package adminportal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/obs"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// Admin management is restricted to actors with role="admin". Other
// roles (support / sales / engineer) get 403 from every endpoint here.
//
// New admins get a one-time temporary password shown ONCE in the
// invite response — the inviting admin is responsible for getting it
// to the new admin via a secure side channel (Slack DM, password
// manager share, etc.). On first login the new admin is required to
// change their password before they can do anything else.
//
// Enroll-MFA flow on first login is already implemented by Phase 52.5
// since MustChangePassword is enforced first and MFAEnrolled is false
// by default.

const (
	// Temp password is base32, 16 chars — readable, copy-pasteable,
	// 80 bits of entropy. Mirrors enrollment token style.
	tempPasswordBytes = 10
	// Roles supported by the admin portal. Order matters for the SPA
	// dropdown.
	roleAdmin    = "admin"
	roleSupport  = "support"
	roleSales    = "sales"
	roleEngineer = "engineer"
)

var validRoles = map[string]bool{
	roleAdmin:    true,
	roleSupport:  true,
	roleSales:    true,
	roleEngineer: true,
}

// requireRoleAdmin guards routes that only role=admin can hit.
func (s *Server) requireRoleAdmin(w http.ResponseWriter, r *http.Request) bool {
	admin := adminFromContext(r.Context())
	if admin.Role != roleAdmin {
		writeError(w, http.StatusForbidden, "admin role required")
		return false
	}
	return true
}

// ─── DTO ────────────────────────────────────────────────────────────

type adminUserDTO struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	Status             string    `json:"status"`
	MFAEnrolled        bool      `json:"mfa_enrolled"`
	MustChangePassword bool      `json:"must_change_password"`
	LastLogin          time.Time `json:"last_login,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

func adminToDTO(a *store.AdminUser) adminUserDTO {
	status := a.Status
	if status == "" {
		status = "active"
	}
	return adminUserDTO{
		ID: a.ID, Email: a.Email, Role: a.Role, Status: status,
		MFAEnrolled:        a.MFAEnrolled,
		MustChangePassword: a.MustChangePassword,
		LastLogin:          a.LastLogin,
		CreatedAt:          a.CreatedAt,
	}
}

// ─── List ───────────────────────────────────────────────────────────

func (s *Server) handleListAdmins(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	admins, err := s.cfg.Store.ListAdminUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list")
		return
	}
	out := make([]adminUserDTO, 0, len(admins))
	for _, a := range admins {
		out = append(out, adminToDTO(a))
	}
	writeJSON(w, http.StatusOK, map[string]any{"admins": out})
}

// ─── Create / invite ────────────────────────────────────────────────

type createAdminReq struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type createAdminResp struct {
	adminUserDTO
	TempPassword string `json:"temp_password"` // shown ONCE
}

func (s *Server) handleCreateAdmin(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	var req createAdminReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		writeError(w, http.StatusBadRequest, "invalid email")
		return
	}
	if !validRoles[req.Role] {
		writeError(w, http.StatusBadRequest, "invalid role")
		return
	}
	if _, err := s.cfg.Store.GetAdminUserByEmail(r.Context(), req.Email); err == nil {
		writeError(w, http.StatusConflict, "admin already exists")
		return
	} else if !errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusInternalServerError, "lookup")
		return
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "password gen")
		return
	}
	a := &store.AdminUser{
		ID:                 NewAdminID(),
		Email:              req.Email,
		PasswordHash:       hash,
		Role:               req.Role,
		Status:             "active",
		MustChangePassword: true,
		CreatedAt:          s.cfg.Now(),
	}
	if err := s.cfg.Store.CreateAdminUser(r.Context(), a); err != nil {
		writeError(w, http.StatusInternalServerError, "create")
		return
	}
	obs.AdminActionsTotal.WithLabelValues("create_admin").Inc()
	s.appendAudit("admin.user_created", adminFromContext(r.Context()), map[string]any{
		"new_admin_id": a.ID, "new_admin_email": a.Email, "role": a.Role,
	})
	writeJSON(w, http.StatusCreated, createAdminResp{
		adminUserDTO: adminToDTO(a),
		TempPassword: temp,
	})
}

// ─── Update role ────────────────────────────────────────────────────

type updateAdminReq struct {
	Role *string `json:"role,omitempty"`
}

func (s *Server) handleUpdateAdmin(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	target, err := s.cfg.Store.GetAdminUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "admin not found")
		return
	}
	var req updateAdminReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	caller := adminFromContext(r.Context())
	if req.Role != nil {
		if target.ID == caller.ID {
			writeError(w, http.StatusBadRequest, "cannot change your own role")
			return
		}
		if !validRoles[*req.Role] {
			writeError(w, http.StatusBadRequest, "invalid role")
			return
		}
		target.Role = *req.Role
	}
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), target); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	obs.AdminActionsTotal.WithLabelValues("update_admin").Inc()
	s.appendAudit("admin.user_updated", caller, map[string]any{
		"target_admin_id": target.ID, "role": target.Role,
	})
	writeJSON(w, http.StatusOK, adminToDTO(target))
}

// ─── Disable / Re-enable ────────────────────────────────────────────

type setStatusReq struct {
	Status string `json:"status"` // active | disabled
}

func (s *Server) handleSetAdminStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	caller := adminFromContext(r.Context())
	if id == caller.ID {
		writeError(w, http.StatusBadRequest, "cannot change your own status")
		return
	}
	target, err := s.cfg.Store.GetAdminUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "admin not found")
		return
	}
	var req setStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	switch req.Status {
	case "active", "disabled":
	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	target.Status = req.Status
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), target); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	// Disabling kills active sessions immediately.
	if req.Status == "disabled" {
		_ = s.cfg.Store.DeleteAdminSessionsForUser(r.Context(), target.ID)
	}
	obs.AdminActionsTotal.WithLabelValues("set_admin_status").Inc()
	s.appendAudit("admin.user_status_changed", caller, map[string]any{
		"target_admin_id": target.ID, "status": target.Status,
	})
	writeJSON(w, http.StatusOK, adminToDTO(target))
}

// ─── Reset password ─────────────────────────────────────────────────
// Returns a new temp password ONCE. Drops sessions for the target so
// they're forced to re-login with the new temp.

type resetPasswordResp struct {
	adminUserDTO
	TempPassword string `json:"temp_password"`
}

func (s *Server) handleResetAdminPassword(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	caller := adminFromContext(r.Context())
	if id == caller.ID {
		writeError(w, http.StatusBadRequest, "cannot reset your own password here — use change-password")
		return
	}
	target, err := s.cfg.Store.GetAdminUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "admin not found")
		return
	}
	temp, hash, err := newTempPassword()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "password gen")
		return
	}
	target.PasswordHash = hash
	target.MustChangePassword = true
	if err := s.cfg.Store.UpdateAdminUser(r.Context(), target); err != nil {
		writeError(w, http.StatusInternalServerError, "update")
		return
	}
	_ = s.cfg.Store.DeleteAdminSessionsForUser(r.Context(), target.ID)
	obs.AdminActionsTotal.WithLabelValues("reset_admin_password").Inc()
	s.appendAudit("admin.user_password_reset", caller, map[string]any{
		"target_admin_id": target.ID,
	})
	writeJSON(w, http.StatusOK, resetPasswordResp{
		adminUserDTO: adminToDTO(target),
		TempPassword: temp,
	})
}

// ─── Delete ─────────────────────────────────────────────────────────

func (s *Server) handleDeleteAdmin(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	id := r.PathValue("id")
	caller := adminFromContext(r.Context())
	if id == caller.ID {
		writeError(w, http.StatusBadRequest, "cannot delete yourself")
		return
	}
	target, err := s.cfg.Store.GetAdminUser(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "admin not found")
		return
	}
	if err := s.cfg.Store.DeleteAdminUser(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "delete")
		return
	}
	obs.AdminActionsTotal.WithLabelValues("delete_admin").Inc()
	s.appendAudit("admin.user_deleted", caller, map[string]any{
		"target_admin_id": target.ID, "target_admin_email": target.Email,
	})
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

// ─── Helpers ────────────────────────────────────────────────────────

// newTempPassword returns (plaintext, bcrypt hash). Plaintext is 16
// uppercase base32 chars — readable + copy-pasteable.
func newTempPassword() (string, string, error) {
	raw := make([]byte, tempPasswordBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	// hex is fine; user pastes once then changes it anyway.
	plain := strings.ToUpper(hex.EncodeToString(raw))
	hash, err := HashPassword(plain)
	if err != nil {
		return "", "", err
	}
	return plain, hash, nil
}
