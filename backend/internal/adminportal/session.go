package adminportal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

type ctxKey int

const (
	ctxKeyAdmin ctxKey = iota
	ctxKeySession
)

// requireStage1 enforces: valid session cookie + valid admin user.
// Used for routes that the user can hit before MFA verification
// (mfa-challenge, mfa-setup, me, logout).
func (s *Server) requireStage1(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admin, sess, ok := s.lookupSession(w, r)
		if !ok {
			return
		}
		ctx := context.WithValue(r.Context(), ctxKeyAdmin, admin)
		ctx = context.WithValue(ctx, ctxKeySession, sess)
		next(w, r.WithContext(ctx))
	}
}

// requireStage2 enforces stage-1 + MFAVerified=true.
func (s *Server) requireStage2(next http.HandlerFunc) http.HandlerFunc {
	return s.requireStage1(func(w http.ResponseWriter, r *http.Request) {
		sess := sessionFromContext(r.Context())
		if !sess.MFAVerified {
			writeError(w, http.StatusUnauthorized, "mfa required")
			return
		}
		next(w, r)
	})
}

// lookupSession validates the cookie + reads the admin row. Returns
// (admin, session, true) on success. On failure it writes a 401 and
// returns ok=false.
func (s *Server) lookupSession(w http.ResponseWriter, r *http.Request) (*store.AdminUser, *store.AdminSession, bool) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil || cookie.Value == "" {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return nil, nil, false
	}
	sess, err := s.cfg.Store.GetAdminSession(r.Context(), cookie.Value)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return nil, nil, false
	}
	now := s.cfg.Now()
	// Absolute lifetime cap.
	if now.Sub(sess.CreatedAt) > SessionMaxLife {
		_ = s.cfg.Store.DeleteAdminSession(r.Context(), sess.ID)
		s.clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "session expired (max lifetime)")
		return nil, nil, false
	}
	// Idle expiry.
	if !sess.ExpiresAt.IsZero() && now.After(sess.ExpiresAt) {
		_ = s.cfg.Store.DeleteAdminSession(r.Context(), sess.ID)
		s.clearSessionCookie(w)
		writeError(w, http.StatusUnauthorized, "session expired (idle)")
		return nil, nil, false
	}
	admin, err := s.cfg.Store.GetAdminUser(r.Context(), sess.AdminUserID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return nil, nil, false
	}
	// Roll idle TTL forward, never past absolute cap.
	newIdle := now.Add(SessionIdleTTL)
	cap := sess.CreatedAt.Add(SessionMaxLife)
	if newIdle.After(cap) {
		newIdle = cap
	}
	_ = s.cfg.Store.TouchAdminSession(r.Context(), sess.ID, now, newIdle)
	sess.LastSeen = now
	sess.ExpiresAt = newIdle
	s.setSessionCookie(w, sess.ID, newIdle)
	return admin, sess, true
}

func adminFromContext(ctx context.Context) *store.AdminUser {
	a, ok := ctx.Value(ctxKeyAdmin).(*store.AdminUser)
	if !ok {
		panic("adminportal: no admin in context — middleware not applied?")
	}
	return a
}

func sessionFromContext(ctx context.Context) *store.AdminSession {
	s, ok := ctx.Value(ctxKeySession).(*store.AdminSession)
	if !ok {
		panic("adminportal: no session in context — middleware not applied?")
	}
	return s
}

func (s *Server) setSessionCookie(w http.ResponseWriter, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *Server) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.cfg.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
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
