package store

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Memory is an in-memory implementation of Store. Goroutine-safe.
// Used for tests, local dev, and pre-Postgres deployments. State is
// lost on process restart — do not use in production.
type Memory struct {
	mu                sync.RWMutex
	customers         map[string]*Customer        // by ID
	customersByEmail  map[string]string           // email → ID
	subscriptions     map[string]*Subscription    // by ID
	subsByStripeID    map[string]string           // Stripe ID → sub ID
	deployments       map[string]*Deployment      // by DeploymentID (the UUID, not row ID)
	enrollmentTokens  map[string]*EnrollmentToken // by token hash
	licenses          map[string]*License         // by JTI
	licensesByLicID   map[string][]string         // license_id → []jti (newest last)
	entitlementSets   map[string]*EntitlementSet  // by ID
	adminUsers        map[string]*AdminUser       // by email
	webhookEvents     map[string]*WebhookEvent    // by provider event ID
}

// NewMemory returns a fresh in-memory Store.
func NewMemory() *Memory {
	return &Memory{
		customers:        map[string]*Customer{},
		customersByEmail: map[string]string{},
		subscriptions:    map[string]*Subscription{},
		subsByStripeID:   map[string]string{},
		deployments:      map[string]*Deployment{},
		enrollmentTokens: map[string]*EnrollmentToken{},
		licenses:         map[string]*License{},
		licensesByLicID:  map[string][]string{},
		entitlementSets:  map[string]*EntitlementSet{},
		adminUsers:       map[string]*AdminUser{},
		webhookEvents:    map[string]*WebhookEvent{},
	}
}

// Close clears all state. Safe to call multiple times. Must not
// replace the struct itself — the mutex is held during this call,
// so we clear maps in place.
func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customers = map[string]*Customer{}
	m.customersByEmail = map[string]string{}
	m.subscriptions = map[string]*Subscription{}
	m.subsByStripeID = map[string]string{}
	m.deployments = map[string]*Deployment{}
	m.enrollmentTokens = map[string]*EnrollmentToken{}
	m.licenses = map[string]*License{}
	m.licensesByLicID = map[string][]string{}
	m.entitlementSets = map[string]*EntitlementSet{}
	m.adminUsers = map[string]*AdminUser{}
	m.webhookEvents = map[string]*WebhookEvent{}
	return nil
}

// ── Webhook events ─────────────────────────────────────────────

func (m *Memory) RecordWebhookEvent(_ context.Context, e *WebhookEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.webhookEvents[e.ID]; exists {
		return ErrAlreadyExists
	}
	cp := *e
	m.webhookEvents[e.ID] = &cp
	return nil
}

func (m *Memory) GetWebhookEvent(_ context.Context, id string) (*WebhookEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.webhookEvents[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *e
	return &cp, nil
}

func (m *Memory) MarkWebhookProcessed(_ context.Context, id string, status, lastError string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.webhookEvents[id]
	if !ok {
		return ErrNotFound
	}
	e.Status = status
	e.LastError = lastError
	e.ProcessedAt = time.Now().UTC()
	return nil
}

// ── Customers ───────────────────────────────────────────────────

func (m *Memory) CreateCustomer(_ context.Context, c *Customer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.customersByEmail[c.Email]; exists {
		return ErrAlreadyExists
	}
	cp := *c
	m.customers[c.ID] = &cp
	m.customersByEmail[c.Email] = c.ID
	return nil
}

func (m *Memory) GetCustomer(_ context.Context, id string) (*Customer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.customers[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *Memory) GetCustomerByEmail(_ context.Context, email string) (*Customer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.customersByEmail[email]
	if !ok {
		return nil, ErrNotFound
	}
	c := m.customers[id]
	cp := *c
	return &cp, nil
}

func (m *Memory) UpdateCustomer(_ context.Context, c *Customer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.customers[c.ID]; !ok {
		return ErrNotFound
	}
	cp := *c
	m.customers[c.ID] = &cp
	m.customersByEmail[c.Email] = c.ID
	return nil
}

// ── Subscriptions ──────────────────────────────────────────────

func (m *Memory) CreateSubscription(_ context.Context, s *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.subscriptions[s.ID]; exists {
		return ErrAlreadyExists
	}
	cp := *s
	m.subscriptions[s.ID] = &cp
	if s.StripeSubscriptionID != "" {
		m.subsByStripeID[s.StripeSubscriptionID] = s.ID
	}
	return nil
}

func (m *Memory) GetSubscription(_ context.Context, id string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.subscriptions[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *Memory) GetSubscriptionByStripeID(_ context.Context, stripeID string) (*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id, ok := m.subsByStripeID[stripeID]
	if !ok {
		return nil, ErrNotFound
	}
	s := m.subscriptions[id]
	cp := *s
	return &cp, nil
}

func (m *Memory) ListSubscriptionsByCustomer(_ context.Context, customerID string) ([]*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Subscription
	for _, s := range m.subscriptions {
		if s.CustomerID == customerID {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ListSubscriptionsWithTrialEndIn returns subs with non-zero
// TrialEnd in the (start, end] window. Iterates everything — fine
// at the scale this implementation supports.
func (m *Memory) ListSubscriptionsWithTrialEndIn(_ context.Context, start, end time.Time) ([]*Subscription, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Subscription
	for _, s := range m.subscriptions {
		if s.TrialEnd.IsZero() {
			continue
		}
		if s.TrialEnd.After(start) && !s.TrialEnd.After(end) {
			cp := *s
			// Defensive copy of slice so caller mutations don't
			// affect stored state.
			if len(s.TrialRemindersSent) > 0 {
				cp.TrialRemindersSent = append([]int{}, s.TrialRemindersSent...)
			}
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *Memory) UpdateSubscription(_ context.Context, s *Subscription) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.subscriptions[s.ID]; !ok {
		return ErrNotFound
	}
	cp := *s
	m.subscriptions[s.ID] = &cp
	if s.StripeSubscriptionID != "" {
		m.subsByStripeID[s.StripeSubscriptionID] = s.ID
	}
	return nil
}

// ── Deployments ────────────────────────────────────────────────

func (m *Memory) UpsertDeployment(_ context.Context, d *Deployment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *d
	m.deployments[d.DeploymentID] = &cp
	return nil
}

func (m *Memory) GetDeployment(_ context.Context, deploymentID string) (*Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.deployments[deploymentID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *d
	return &cp, nil
}

func (m *Memory) ListDeploymentsByCustomer(_ context.Context, customerID string) ([]*Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Deployment
	for _, d := range m.deployments {
		if d.CustomerID == customerID {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *Memory) ListFlaggedDeployments(_ context.Context) ([]*Deployment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []*Deployment
	for _, d := range m.deployments {
		if d.FlaggedForReview {
			cp := *d
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── Enrollment tokens ──────────────────────────────────────────

func (m *Memory) CreateEnrollmentToken(_ context.Context, t *EnrollmentToken) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.enrollmentTokens[t.TokenHash]; exists {
		return ErrAlreadyExists
	}
	cp := *t
	m.enrollmentTokens[t.TokenHash] = &cp
	return nil
}

func (m *Memory) GetEnrollmentTokenByHash(_ context.Context, tokenHash string) (*EnrollmentToken, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.enrollmentTokens[tokenHash]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *t
	return &cp, nil
}

func (m *Memory) ConsumeEnrollmentToken(_ context.Context, tokenHash, deploymentID, responseJWS string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.enrollmentTokens[tokenHash]
	if !ok {
		return ErrNotFound
	}
	if !t.ConsumedAt.IsZero() {
		// Already consumed. If by the same deployment, idempotent:
		// caller will re-serve the cached response. If by a different
		// deployment, hard reject.
		if t.ConsumedByDeployment != deploymentID {
			return ErrConsumed
		}
		return nil
	}
	t.ConsumedAt = time.Now().UTC()
	t.ConsumedByDeployment = deploymentID
	t.ConsumedResponseJWS = responseJWS
	return nil
}

// ── Licenses ───────────────────────────────────────────────────

func (m *Memory) CreateLicense(_ context.Context, l *License) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.licenses[l.JTI]; exists {
		return ErrAlreadyExists
	}
	cp := *l
	m.licenses[l.JTI] = &cp
	m.licensesByLicID[l.LicenseID] = append(m.licensesByLicID[l.LicenseID], l.JTI)
	return nil
}

func (m *Memory) GetLicenseByJTI(_ context.Context, jti string) (*License, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.licenses[jti]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *l
	return &cp, nil
}

func (m *Memory) GetLatestLicenseForLicenseID(_ context.Context, licenseID string) (*License, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	jtis := m.licensesByLicID[licenseID]
	if len(jtis) == 0 {
		return nil, ErrNotFound
	}
	latest := m.licenses[jtis[len(jtis)-1]]
	cp := *latest
	return &cp, nil
}

func (m *Memory) RevokeLicenseByJTI(_ context.Context, jti, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	l, ok := m.licenses[jti]
	if !ok {
		return ErrNotFound
	}
	l.Revoked = true
	l.RevokedAt = time.Now().UTC()
	l.RevokedReason = reason
	return nil
}

func (m *Memory) IsJTIRevoked(_ context.Context, jti string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	l, ok := m.licenses[jti]
	if !ok {
		return false, nil // unknown JTI is not revoked (might be from a different issuer state)
	}
	return l.Revoked, nil
}

// ── Entitlement sets ───────────────────────────────────────────

func (m *Memory) CreateEntitlementSet(_ context.Context, e *EntitlementSet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.entitlementSets[e.ID]; exists {
		return ErrAlreadyExists
	}
	cp := *e
	m.entitlementSets[e.ID] = &cp
	return nil
}

func (m *Memory) GetEntitlementSet(_ context.Context, id string) (*EntitlementSet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	e, ok := m.entitlementSets[id]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *e
	return &cp, nil
}

// ── Admin users ────────────────────────────────────────────────

func (m *Memory) CreateAdminUser(_ context.Context, a *AdminUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.adminUsers[a.Email]; exists {
		return ErrAlreadyExists
	}
	cp := *a
	m.adminUsers[a.Email] = &cp
	return nil
}

func (m *Memory) GetAdminUserByEmail(_ context.Context, email string) (*AdminUser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	a, ok := m.adminUsers[email]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *a
	return &cp, nil
}

// Compile-time check that Memory implements Store.
var _ Store = (*Memory)(nil)

// errOrNil is a small helper used by tests.
func errOrNil(err error) error {
	if errors.Is(err, ErrNotFound) {
		return err
	}
	return err
}

var _ = errOrNil
