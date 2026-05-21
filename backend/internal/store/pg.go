// PostgreSQL Store implementation (Phase 49 follow-up).
//
// Wire up by setting the issuer's --pg-url flag (or PG_URL env). When
// absent, the in-memory implementation in memory.go is used instead.
//
// Schema lives in pg_schema.sql and is applied idempotently on Open()
// via the embedded SQL — so a fresh PG database becomes usable without
// a separate migration step.
//
// For tests, set TEST_PG_DSN. Without it, pg_test.go skips all tests
// (matches the NodePulse pgstore convention).
package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed pg_schema.sql
var pgSchemaSQL string

// PG implements Store backed by a Postgres connection pool.
type PG struct {
	pool *pgxpool.Pool
}

// NewPG opens a connection pool to the given DSN, applies the schema
// idempotently, and returns a ready Store.
//
//	dsn: postgres://user:pass@host:port/dbname?sslmode=disable
func NewPG(ctx context.Context, dsn string) (*PG, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("pg: dial: %w", err)
	}
	if _, err := pool.Exec(ctx, pgSchemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg: apply schema: %w", err)
	}
	return &PG{pool: pool}, nil
}

func (s *PG) Close() error {
	if s.pool != nil {
		s.pool.Close()
	}
	return nil
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

// scanStrOrEmpty extracts a *string into a plain string, treating NULL as "".
type optStr struct{ v *string }

func (o *optStr) String() string {
	if o.v == nil {
		return ""
	}
	return *o.v
}

type optTime struct{ v *time.Time }

func (o *optTime) Time() time.Time {
	if o.v == nil {
		return time.Time{}
	}
	return *o.v
}

// ── Customers ───────────────────────────────────────────────────

func (s *PG) CreateCustomer(ctx context.Context, c *Customer) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO customers (id, email, name, company, stripe_customer_id, status,
			password_hash, email_verified_at, token_hash, token_expires, token_purpose,
			created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, c.ID, c.Email, nullStr(c.Name), nullStr(c.Company), nullStr(c.StripeCustomerID), c.Status,
		c.PasswordHash, nullTime(c.EmailVerifiedAt), c.TokenHash, nullTime(c.TokenExpires), c.TokenPurpose,
		c.CreatedAt, c.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

const sqlSelectCustomer = `
	SELECT id, email, name, company, stripe_customer_id, status,
		password_hash, email_verified_at, token_hash, token_expires, token_purpose,
		created_at, updated_at
	FROM customers
`

func (s *PG) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	return s.scanCustomer(s.pool.QueryRow(ctx, sqlSelectCustomer+` WHERE id=$1`, id))
}

func (s *PG) GetCustomerByEmail(ctx context.Context, email string) (*Customer, error) {
	return s.scanCustomer(s.pool.QueryRow(ctx, sqlSelectCustomer+` WHERE email=$1`, email))
}

func (s *PG) GetCustomerByTokenHash(ctx context.Context, tokenHash string) (*Customer, error) {
	if tokenHash == "" {
		return nil, ErrNotFound
	}
	return s.scanCustomer(s.pool.QueryRow(ctx, sqlSelectCustomer+` WHERE token_hash=$1`, tokenHash))
}

func (s *PG) scanCustomer(row pgx.Row) (*Customer, error) {
	var c Customer
	var name, company, stripeID *string
	var emailVerifiedAt, tokenExpires *time.Time
	err := row.Scan(&c.ID, &c.Email, &name, &company, &stripeID, &c.Status,
		&c.PasswordHash, &emailVerifiedAt, &c.TokenHash, &tokenExpires, &c.TokenPurpose,
		&c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	c.Name = (&optStr{name}).String()
	c.Company = (&optStr{company}).String()
	c.StripeCustomerID = (&optStr{stripeID}).String()
	c.EmailVerifiedAt = (&optTime{emailVerifiedAt}).Time()
	c.TokenExpires = (&optTime{tokenExpires}).Time()
	return &c, nil
}

func (s *PG) UpdateCustomer(ctx context.Context, c *Customer) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE customers SET email=$2,name=$3,company=$4,stripe_customer_id=$5,status=$6,
			password_hash=$7, email_verified_at=$8, token_hash=$9, token_expires=$10, token_purpose=$11,
			updated_at=$12
		WHERE id=$1
	`, c.ID, c.Email, nullStr(c.Name), nullStr(c.Company), nullStr(c.StripeCustomerID), c.Status,
		c.PasswordHash, nullTime(c.EmailVerifiedAt), c.TokenHash, nullTime(c.TokenExpires), c.TokenPurpose,
		c.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Subscriptions ──────────────────────────────────────────────

func (s *PG) CreateSubscription(ctx context.Context, sub *Subscription) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO subscriptions (id,customer_id,tier,stripe_subscription_id,stripe_price_id,status,
			current_period_start,current_period_end,cancel_at,canceled_at,trial_end,manual_offline,trial_reminders_sent,created_at,updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, sub.ID, sub.CustomerID, sub.Tier, nullStr(sub.StripeSubscriptionID), nullStr(sub.StripePriceID), sub.Status,
		nullTime(sub.CurrentPeriodStart), nullTime(sub.CurrentPeriodEnd),
		nullTime(sub.CancelAt), nullTime(sub.CanceledAt), nullTime(sub.TrialEnd),
		sub.ManualOffline, intSliceToCSV(sub.TrialRemindersSent), sub.CreatedAt, sub.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

// ListSubscriptionsWithTrialEndIn returns subs with non-zero
// trial_end falling in (start, end]. Caller passes both bounds
// rather than using SQL NOW() so the scheduler's injected clock
// flows through to tests. Indexed via idx_subscriptions_trial_end.
func (s *PG) ListSubscriptionsWithTrialEndIn(ctx context.Context, start, end time.Time) ([]*Subscription, error) {
	rows, err := s.pool.Query(ctx, sqlSelectSubscription+`
		WHERE trial_end IS NOT NULL
		  AND trial_end > $1
		  AND trial_end <= $2
		ORDER BY trial_end ASC
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Subscription
	for rows.Next() {
		sub, err := s.scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *PG) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	return s.scanSubscription(s.pool.QueryRow(ctx, sqlSelectSubscription+` WHERE id=$1`, id))
}

func (s *PG) GetSubscriptionByStripeID(ctx context.Context, stripeID string) (*Subscription, error) {
	return s.scanSubscription(s.pool.QueryRow(ctx, sqlSelectSubscription+` WHERE stripe_subscription_id=$1`, stripeID))
}

func (s *PG) ListSubscriptionsByCustomer(ctx context.Context, customerID string) ([]*Subscription, error) {
	rows, err := s.pool.Query(ctx, sqlSelectSubscription+` WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Subscription
	for rows.Next() {
		sub, err := s.scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *PG) UpdateSubscription(ctx context.Context, sub *Subscription) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE subscriptions SET tier=$2,stripe_subscription_id=$3,stripe_price_id=$4,status=$5,
			current_period_start=$6,current_period_end=$7,cancel_at=$8,canceled_at=$9,trial_end=$10,
			manual_offline=$11,trial_reminders_sent=$12,updated_at=$13
		WHERE id=$1
	`, sub.ID, sub.Tier, nullStr(sub.StripeSubscriptionID), nullStr(sub.StripePriceID), sub.Status,
		nullTime(sub.CurrentPeriodStart), nullTime(sub.CurrentPeriodEnd),
		nullTime(sub.CancelAt), nullTime(sub.CanceledAt), nullTime(sub.TrialEnd),
		sub.ManualOffline, intSliceToCSV(sub.TrialRemindersSent), sub.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const sqlSelectSubscription = `
	SELECT id,customer_id,tier,stripe_subscription_id,stripe_price_id,status,
		current_period_start,current_period_end,cancel_at,canceled_at,trial_end,
		manual_offline,trial_reminders_sent,created_at,updated_at
	FROM subscriptions
`

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *PG) scanSubscription(row rowScanner) (*Subscription, error) {
	var sub Subscription
	var stripeID, stripePriceID *string
	var startTime, endTime, cancelAt, canceledAt, trialEnd *time.Time
	var remindersCSV string
	err := row.Scan(&sub.ID, &sub.CustomerID, &sub.Tier, &stripeID, &stripePriceID, &sub.Status,
		&startTime, &endTime, &cancelAt, &canceledAt, &trialEnd,
		&sub.ManualOffline, &remindersCSV, &sub.CreatedAt, &sub.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sub.StripeSubscriptionID = (&optStr{stripeID}).String()
	sub.StripePriceID = (&optStr{stripePriceID}).String()
	sub.CurrentPeriodStart = (&optTime{startTime}).Time()
	sub.CurrentPeriodEnd = (&optTime{endTime}).Time()
	sub.CancelAt = (&optTime{cancelAt}).Time()
	sub.CanceledAt = (&optTime{canceledAt}).Time()
	sub.TrialEnd = (&optTime{trialEnd}).Time()
	sub.TrialRemindersSent = csvToIntSlice(remindersCSV)
	return &sub, nil
}

// intSliceToCSV serializes []int{30, 7} to "30,7" for PG storage.
// Empty slice → "".
func intSliceToCSV(xs []int) string {
	if len(xs) == 0 {
		return ""
	}
	parts := make([]string, len(xs))
	for i, n := range xs {
		parts[i] = fmt.Sprintf("%d", n)
	}
	return strings.Join(parts, ",")
}

// csvToIntSlice parses "30,7" back to []int{30, 7}. Empty input → nil.
// Malformed entries are silently skipped (defensive — don't break
// scheduler on a corrupted row).
func csvToIntSlice(s string) []int {
	if s == "" {
		return nil
	}
	var out []int
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err == nil {
			out = append(out, n)
		}
	}
	return out
}

// ── Deployments ────────────────────────────────────────────────

func (s *PG) UpsertDeployment(ctx context.Context, d *Deployment) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO deployments (id,customer_id,deployment_id,deployment_name,first_seen_at,last_seen_at,last_version,flagged_for_review,flag_reason,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (deployment_id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			last_version = COALESCE(EXCLUDED.last_version, deployments.last_version),
			deployment_name = COALESCE(EXCLUDED.deployment_name, deployments.deployment_name)
	`, d.ID, d.CustomerID, d.DeploymentID, nullStr(d.DeploymentName), d.FirstSeenAt, d.LastSeenAt, nullStr(d.LastVersion), d.FlaggedForReview, nullStr(d.FlagReason), d.CreatedAt)
	return err
}

func (s *PG) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	return s.scanDeployment(s.pool.QueryRow(ctx, sqlSelectDeployment+` WHERE deployment_id=$1`, deploymentID))
}

func (s *PG) ListDeploymentsByCustomer(ctx context.Context, customerID string) ([]*Deployment, error) {
	rows, err := s.pool.Query(ctx, sqlSelectDeployment+` WHERE customer_id=$1 ORDER BY last_seen_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Deployment
	for rows.Next() {
		d, err := s.scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *PG) ListFlaggedDeployments(ctx context.Context) ([]*Deployment, error) {
	rows, err := s.pool.Query(ctx, sqlSelectDeployment+` WHERE flagged_for_review = TRUE ORDER BY last_seen_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*Deployment
	for rows.Next() {
		d, err := s.scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

const sqlSelectDeployment = `
	SELECT id,customer_id,deployment_id,deployment_name,first_seen_at,last_seen_at,last_version,flagged_for_review,flag_reason,created_at
	FROM deployments
`

func (s *PG) scanDeployment(row rowScanner) (*Deployment, error) {
	var d Deployment
	var name, version, reason *string
	err := row.Scan(&d.ID, &d.CustomerID, &d.DeploymentID, &name, &d.FirstSeenAt, &d.LastSeenAt, &version, &d.FlaggedForReview, &reason, &d.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	d.DeploymentName = (&optStr{name}).String()
	d.LastVersion = (&optStr{version}).String()
	d.FlagReason = (&optStr{reason}).String()
	return &d, nil
}

// ── Enrollment tokens ──────────────────────────────────────────

func (s *PG) CreateEnrollmentToken(ctx context.Context, t *EnrollmentToken) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO enrollment_tokens (id,customer_id,subscription_id,token_hash,label,created_at,expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, t.ID, t.CustomerID, t.SubscriptionID, t.TokenHash, nullStr(t.Label), t.CreatedAt, t.ExpiresAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PG) GetEnrollmentTokenByHash(ctx context.Context, tokenHash string) (*EnrollmentToken, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id,customer_id,subscription_id,token_hash,label,created_at,expires_at,consumed_at,consumed_by_deployment,consumed_response_jws
		FROM enrollment_tokens WHERE token_hash=$1
	`, tokenHash)
	var t EnrollmentToken
	var label, consumedBy, response *string
	var consumedAt *time.Time
	err := row.Scan(&t.ID, &t.CustomerID, &t.SubscriptionID, &t.TokenHash, &label, &t.CreatedAt, &t.ExpiresAt, &consumedAt, &consumedBy, &response)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.Label = (&optStr{label}).String()
	t.ConsumedByDeployment = (&optStr{consumedBy}).String()
	t.ConsumedResponseJWS = (&optStr{response}).String()
	t.ConsumedAt = (&optTime{consumedAt}).Time()
	return &t, nil
}

func (s *PG) ListEnrollmentTokensByCustomer(ctx context.Context, customerID string) ([]*EnrollmentToken, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id,customer_id,subscription_id,token_hash,label,created_at,expires_at,consumed_at,consumed_by_deployment,consumed_response_jws
		FROM enrollment_tokens WHERE customer_id=$1 ORDER BY created_at DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*EnrollmentToken
	for rows.Next() {
		var t EnrollmentToken
		var label, consumedBy, response *string
		var consumedAt *time.Time
		if err := rows.Scan(&t.ID, &t.CustomerID, &t.SubscriptionID, &t.TokenHash, &label, &t.CreatedAt, &t.ExpiresAt, &consumedAt, &consumedBy, &response); err != nil {
			return nil, err
		}
		t.Label = (&optStr{label}).String()
		t.ConsumedByDeployment = (&optStr{consumedBy}).String()
		t.ConsumedResponseJWS = (&optStr{response}).String()
		t.ConsumedAt = (&optTime{consumedAt}).Time()
		out = append(out, &t)
	}
	return out, rows.Err()
}

func (s *PG) RevokeEnrollmentToken(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `UPDATE enrollment_tokens SET expires_at=NOW() WHERE id=$1 AND expires_at > NOW()`, id)
	return err
}

// ConsumeEnrollmentToken serialises consume-or-cached-replay via a
// SELECT FOR UPDATE on the token row inside a transaction. Idempotent
// for the same deployment_id; returns ErrConsumed for a different one.
func (s *PG) ConsumeEnrollmentToken(ctx context.Context, tokenHash, deploymentID, responseJWS string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var consumedAt *time.Time
	var consumedBy *string
	row := tx.QueryRow(ctx, `SELECT consumed_at, consumed_by_deployment FROM enrollment_tokens WHERE token_hash=$1 FOR UPDATE`, tokenHash)
	if err := row.Scan(&consumedAt, &consumedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if consumedAt != nil {
		if consumedBy != nil && *consumedBy != deploymentID {
			return ErrConsumed
		}
		// Same-deployment retry — idempotent. Don't update.
		return tx.Commit(ctx)
	}
	_, err = tx.Exec(ctx, `
		UPDATE enrollment_tokens
		SET consumed_at=NOW(), consumed_by_deployment=$2, consumed_response_jws=$3
		WHERE token_hash=$1
	`, tokenHash, deploymentID, responseJWS)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ── Licenses ───────────────────────────────────────────────────

func (s *PG) CreateLicense(ctx context.Context, l *License) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO licenses (id,jti,license_id,customer_id,subscription_id,deployment_id,entitlement_set_id,tier,
			iat,not_before,expires_at,grace_until,revoked,revoked_at,revoked_reason,kid,payload_jws,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
	`, l.ID, l.JTI, l.LicenseID, l.CustomerID, l.SubscriptionID, l.DeploymentID, l.EntitlementSetID, l.Tier,
		l.Iat, l.NotBefore, l.ExpiresAt, l.GraceUntil,
		l.Revoked, nullTime(l.RevokedAt), nullStr(l.RevokedReason),
		l.Kid, l.PayloadJWS, l.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PG) GetLicenseByJTI(ctx context.Context, jti string) (*License, error) {
	return s.scanLicense(s.pool.QueryRow(ctx, sqlSelectLicense+` WHERE jti=$1`, jti))
}

func (s *PG) GetLatestLicenseForLicenseID(ctx context.Context, licenseID string) (*License, error) {
	return s.scanLicense(s.pool.QueryRow(ctx, sqlSelectLicense+` WHERE license_id=$1 ORDER BY created_at DESC LIMIT 1`, licenseID))
}

func (s *PG) ListLicensesByCustomer(ctx context.Context, customerID string) ([]*License, error) {
	rows, err := s.pool.Query(ctx, sqlSelectLicense+` WHERE customer_id=$1 ORDER BY created_at DESC`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*License
	for rows.Next() {
		l, err := s.scanLicense(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (s *PG) RevokeLicenseByJTI(ctx context.Context, jti, reason string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE licenses SET revoked=TRUE, revoked_at=NOW(), revoked_reason=$2 WHERE jti=$1
	`, jti, nullStr(reason))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PG) IsJTIRevoked(ctx context.Context, jti string) (bool, error) {
	var revoked bool
	err := s.pool.QueryRow(ctx, `SELECT revoked FROM licenses WHERE jti=$1`, jti).Scan(&revoked)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return revoked, nil
}

const sqlSelectLicense = `
	SELECT id,jti,license_id,customer_id,subscription_id,deployment_id,entitlement_set_id,tier,
		iat,not_before,expires_at,grace_until,revoked,revoked_at,revoked_reason,kid,payload_jws,created_at
	FROM licenses
`

func (s *PG) scanLicense(row rowScanner) (*License, error) {
	var l License
	var revokedAt *time.Time
	var revokedReason *string
	err := row.Scan(&l.ID, &l.JTI, &l.LicenseID, &l.CustomerID, &l.SubscriptionID, &l.DeploymentID, &l.EntitlementSetID, &l.Tier,
		&l.Iat, &l.NotBefore, &l.ExpiresAt, &l.GraceUntil,
		&l.Revoked, &revokedAt, &revokedReason,
		&l.Kid, &l.PayloadJWS, &l.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	l.RevokedAt = (&optTime{revokedAt}).Time()
	l.RevokedReason = (&optStr{revokedReason}).String()
	return &l, nil
}

// ── Entitlement sets ───────────────────────────────────────────

func (s *PG) CreateEntitlementSet(ctx context.Context, e *EntitlementSet) error {
	manifest, err := json.Marshal(map[string]any{
		"features": e.Features,
		"limits": map[string]any{
			"max_agents":          e.Limits.MaxAgents,
			"max_metrics_per_sec": e.Limits.MaxMetricsPerSec,
			"max_log_gb_per_day":  e.Limits.MaxLogGBPerDay,
			"max_alert_rules":     e.Limits.MaxAlertRules,
			"retention_days":      e.Limits.RetentionDays,
		},
	})
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO entitlement_sets (id,name,version,manifest,created_at,deprecated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, e.ID, e.Name, e.Version, manifest, e.CreatedAt, nullTime(e.DeprecatedAt))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PG) GetEntitlementSet(ctx context.Context, id string) (*EntitlementSet, error) {
	var e EntitlementSet
	var manifest []byte
	var deprecatedAt *time.Time
	err := s.pool.QueryRow(ctx, `SELECT id,name,version,manifest,created_at,deprecated_at FROM entitlement_sets WHERE id=$1`, id).
		Scan(&e.ID, &e.Name, &e.Version, &manifest, &e.CreatedAt, &deprecatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var m struct {
		Features []string `json:"features"`
		Limits   struct {
			MaxAgents        int64 `json:"max_agents"`
			MaxMetricsPerSec int64 `json:"max_metrics_per_sec"`
			MaxLogGBPerDay   int64 `json:"max_log_gb_per_day"`
			MaxAlertRules    int64 `json:"max_alert_rules"`
			RetentionDays    int   `json:"retention_days"`
		} `json:"limits"`
	}
	if err := json.Unmarshal(manifest, &m); err != nil {
		return nil, fmt.Errorf("pg: parse entitlement manifest: %w", err)
	}
	e.Features = m.Features
	e.Limits = Limits{
		MaxAgents:        m.Limits.MaxAgents,
		MaxMetricsPerSec: m.Limits.MaxMetricsPerSec,
		MaxLogGBPerDay:   m.Limits.MaxLogGBPerDay,
		MaxAlertRules:    m.Limits.MaxAlertRules,
		RetentionDays:    m.Limits.RetentionDays,
	}
	e.DeprecatedAt = (&optTime{deprecatedAt}).Time()
	return &e, nil
}

// ── Admin users ────────────────────────────────────────────────

func (s *PG) CreateAdminUser(ctx context.Context, a *AdminUser) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_users (id,email,password_hash,role,mfa_secret,last_login,created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, a.ID, a.Email, a.PasswordHash, a.Role, a.MFASecret, nullTime(a.LastLogin), a.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PG) GetAdminUserByEmail(ctx context.Context, email string) (*AdminUser, error) {
	var a AdminUser
	var lastLogin *time.Time
	err := s.pool.QueryRow(ctx, `SELECT id,email,password_hash,role,mfa_secret,last_login,created_at FROM admin_users WHERE email=$1`, email).
		Scan(&a.ID, &a.Email, &a.PasswordHash, &a.Role, &a.MFASecret, &lastLogin, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a.LastLogin = (&optTime{lastLogin}).Time()
	return &a, nil
}

// ── Webhook events ─────────────────────────────────────────────

func (s *PG) RecordWebhookEvent(ctx context.Context, e *WebhookEvent) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO webhook_events (id, type, payload, received_at, status, last_error)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, e.ID, e.Type, e.Payload, e.ReceivedAt, e.Status, nullStr(e.LastError))
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *PG) GetWebhookEvent(ctx context.Context, id string) (*WebhookEvent, error) {
	var e WebhookEvent
	var processedAt *time.Time
	var lastError *string
	err := s.pool.QueryRow(ctx, `
		SELECT id, type, payload, received_at, processed_at, status, last_error
		FROM webhook_events WHERE id=$1
	`, id).Scan(&e.ID, &e.Type, &e.Payload, &e.ReceivedAt, &processedAt, &e.Status, &lastError)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	e.ProcessedAt = (&optTime{processedAt}).Time()
	e.LastError = (&optStr{lastError}).String()
	return &e, nil
}

func (s *PG) MarkWebhookProcessed(ctx context.Context, id, status, lastError string) error {
	tag, err := s.pool.Exec(ctx, `
		UPDATE webhook_events SET processed_at=NOW(), status=$2, last_error=$3 WHERE id=$1
	`, id, status, nullStr(lastError))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ── Portal sessions ────────────────────────────────────────────

func (p *PG) CreatePortalSession(ctx context.Context, s *PortalSession) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO portal_sessions (id, customer_id, ip, user_agent, created_at, last_seen, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, s.ID, s.CustomerID, s.IP, s.UserAgent, s.CreatedAt, s.LastSeen, s.ExpiresAt)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (p *PG) GetPortalSession(ctx context.Context, id string) (*PortalSession, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, customer_id, ip, user_agent, created_at, last_seen, expires_at
		FROM portal_sessions WHERE id=$1
	`, id)
	var s PortalSession
	if err := row.Scan(&s.ID, &s.CustomerID, &s.IP, &s.UserAgent, &s.CreatedAt, &s.LastSeen, &s.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (p *PG) TouchPortalSession(ctx context.Context, id string, lastSeen, expiresAt time.Time) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE portal_sessions SET last_seen=$2, expires_at=$3 WHERE id=$1
	`, id, lastSeen, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *PG) DeletePortalSession(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM portal_sessions WHERE id=$1`, id)
	return err
}

func (p *PG) DeletePortalSessionsForCustomer(ctx context.Context, customerID string) error {
	_, err := p.pool.Exec(ctx, `DELETE FROM portal_sessions WHERE customer_id=$1`, customerID)
	return err
}

// isUniqueViolation matches pgx's exposure of Postgres SQLSTATE 23505
// without dropping a hard dependency on jackc/pgerrcode.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "duplicate key") || contains(msg, "SQLSTATE 23505") || contains(msg, "(23505)")
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Compile-time check that PG implements Store.
var _ Store = (*PG)(nil)
