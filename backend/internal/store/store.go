package store

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a lookup misses. Callers should check
// for this with errors.Is — implementations may wrap it.
var ErrNotFound = errors.New("store: not found")

// ErrAlreadyExists is returned when uniqueness is violated (e.g.
// creating a customer with an existing email).
var ErrAlreadyExists = errors.New("store: already exists")

// ErrConsumed is returned by ConsumeEnrollmentToken when the token
// was already redeemed by a DIFFERENT deployment. Same-deployment
// retries inside the idempotency window return the cached response.
var ErrConsumed = errors.New("store: enrollment token already consumed by another deployment")

// Store is the persistence abstraction. The in-memory implementation
// in memory.go satisfies it; a PostgreSQL adapter in pg.go (TODO)
// satisfies it equivalently for production. Handlers depend on this
// interface, never on a concrete type.
//
// All methods take context.Context so a Postgres impl can honor
// request cancellation. The memory impl ignores ctx (besides
// checking for early cancel).
type Store interface {
	// ── Customers ───────────────────────────────────────────────
	CreateCustomer(ctx context.Context, c *Customer) error
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	GetCustomerByEmail(ctx context.Context, email string) (*Customer, error)
	// GetCustomerByTokenHash finds the customer whose TokenHash matches.
	// Used for verify-email + reset-password redemption (Phase 52). The
	// caller is responsible for checking TokenExpires + TokenPurpose.
	// Returns ErrNotFound if no row matches or the hash is empty.
	GetCustomerByTokenHash(ctx context.Context, tokenHash string) (*Customer, error)
	UpdateCustomer(ctx context.Context, c *Customer) error

	// ── Subscriptions ──────────────────────────────────────────
	CreateSubscription(ctx context.Context, s *Subscription) error
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*Subscription, error)
	ListSubscriptionsByCustomer(ctx context.Context, customerID string) ([]*Subscription, error)
	// ListSubscriptionsWithTrialEndIn returns subscriptions whose
	// TrialEnd is non-zero and falls in (start, end]. Caller passes
	// both bounds explicitly so the scheduler's injected clock can
	// flow through (the in-memory store can't see the scheduler's
	// Now() function otherwise). Drives the trial-expiring email
	// scheduler.
	ListSubscriptionsWithTrialEndIn(ctx context.Context, start, end time.Time) ([]*Subscription, error)
	UpdateSubscription(ctx context.Context, s *Subscription) error

	// ── Deployments ────────────────────────────────────────────
	UpsertDeployment(ctx context.Context, d *Deployment) error
	GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error)
	ListDeploymentsByCustomer(ctx context.Context, customerID string) ([]*Deployment, error)
	ListFlaggedDeployments(ctx context.Context) ([]*Deployment, error)

	// ── Enrollment tokens ──────────────────────────────────────
	CreateEnrollmentToken(ctx context.Context, t *EnrollmentToken) error
	GetEnrollmentTokenByHash(ctx context.Context, tokenHash string) (*EnrollmentToken, error)
	// ListEnrollmentTokensByCustomer returns all enrollment tokens for
	// a customer, newest first. Used by the portal (Phase 52).
	ListEnrollmentTokensByCustomer(ctx context.Context, customerID string) ([]*EnrollmentToken, error)
	// RevokeEnrollmentToken marks a token unusable by setting its
	// ExpiresAt to now. Idempotent — safe to call on already-consumed
	// or already-expired tokens.
	RevokeEnrollmentToken(ctx context.Context, id string) error
	// ConsumeEnrollmentToken records a successful redemption. If the
	// token was already consumed by the SAME deployment within the
	// idempotency window, returns the cached response JWS (no error).
	// If consumed by a DIFFERENT deployment, returns ErrConsumed.
	ConsumeEnrollmentToken(ctx context.Context, tokenHash, deploymentID, responseJWS string) error

	// ── Licenses ───────────────────────────────────────────────
	CreateLicense(ctx context.Context, l *License) error
	GetLicenseByJTI(ctx context.Context, jti string) (*License, error)
	GetLatestLicenseForLicenseID(ctx context.Context, licenseID string) (*License, error)
	// ListLicensesByCustomer returns all licenses for a customer,
	// newest first. Used by the portal to show issued/active licenses.
	ListLicensesByCustomer(ctx context.Context, customerID string) ([]*License, error)
	RevokeLicenseByJTI(ctx context.Context, jti, reason string) error
	IsJTIRevoked(ctx context.Context, jti string) (bool, error)

	// ── Entitlement sets ───────────────────────────────────────
	CreateEntitlementSet(ctx context.Context, e *EntitlementSet) error
	GetEntitlementSet(ctx context.Context, id string) (*EntitlementSet, error)

	// ── Admin users ────────────────────────────────────────────
	CreateAdminUser(ctx context.Context, a *AdminUser) error
	GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error)

	// ── Portal sessions ────────────────────────────────────────
	// Cookie-backed sessions for the customer portal at
	// app.infrays.org. Separate from AdminUser sessions (admin
	// portal is Phase 52.5).
	CreatePortalSession(ctx context.Context, s *PortalSession) error
	GetPortalSession(ctx context.Context, id string) (*PortalSession, error)
	TouchPortalSession(ctx context.Context, id string, lastSeen, expiresAt time.Time) error
	DeletePortalSession(ctx context.Context, id string) error
	DeletePortalSessionsForCustomer(ctx context.Context, customerID string) error

	// ── Webhook events (idempotency) ───────────────────────────
	// RecordWebhookEvent inserts an event by its provider ID. Returns
	// ErrAlreadyExists if the same event ID has been seen before — the
	// handler should respond 200 OK to Stripe without re-processing.
	RecordWebhookEvent(ctx context.Context, e *WebhookEvent) error
	GetWebhookEvent(ctx context.Context, id string) (*WebhookEvent, error)
	MarkWebhookProcessed(ctx context.Context, id string, status, lastError string) error

	// ── Lifecycle ──────────────────────────────────────────────
	Close() error
}
