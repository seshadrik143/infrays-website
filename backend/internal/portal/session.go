package portal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// ctxKey is the type for context keys this package writes. Unexported
// so other packages can't collide.
type ctxKey int

const (
	ctxKeyCustomer ctxKey = iota
	ctxKeySession
)

// requireSession reads the session cookie, looks up the session, then
// looks up the customer. On any failure, returns 401. On success,
// stores the customer + session in request context and calls next.
//
// Activity rolls the session TTL forward — every authenticated request
// extends the cookie's expiry by SessionTTL.
func (s *Server) requireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		sess, err := s.cfg.Store.GetPortalSession(r.Context(), cookie.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		now := s.cfg.Now()
		if !sess.ExpiresAt.IsZero() && now.After(sess.ExpiresAt) {
			_ = s.cfg.Store.DeletePortalSession(r.Context(), sess.ID)
			s.clearSessionCookie(w)
			writeError(w, http.StatusUnauthorized, "session expired")
			return
		}
		customer, err := s.cfg.Store.GetCustomer(r.Context(), sess.CustomerID)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated")
			return
		}
		// Roll TTL forward.
		newExpiry := now.Add(SessionTTL)
		_ = s.cfg.Store.TouchPortalSession(r.Context(), sess.ID, now, newExpiry)
		sess.LastSeen = now
		sess.ExpiresAt = newExpiry
		s.setSessionCookie(w, sess.ID, newExpiry)

		ctx := context.WithValue(r.Context(), ctxKeyCustomer, customer)
		ctx = context.WithValue(ctx, ctxKeySession, sess)
		next(w, r.WithContext(ctx))
	}
}

// customerFromContext retrieves the authenticated customer placed by
// requireSession. Panics if used outside an authenticated handler —
// that would be a programming error.
func customerFromContext(ctx context.Context) *store.Customer {
	c, ok := ctx.Value(ctxKeyCustomer).(*store.Customer)
	if !ok {
		panic("portal: no customer in context — requireSession not applied?")
	}
	return c
}

// sessionFromContext returns the active session row.
func sessionFromContext(ctx context.Context) *store.PortalSession {
	s, ok := ctx.Value(ctxKeySession).(*store.PortalSession)
	if !ok {
		panic("portal: no session in context — requireSession not applied?")
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

// newSessionID returns a 32-byte hex string suitable for a session
// cookie value. 256 bits of entropy.
func newSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// errStoreNotFound is a sentinel used by handlers to distinguish a
// not-found row from other store errors.
var errStoreNotFound = errors.New("portal: not found")
