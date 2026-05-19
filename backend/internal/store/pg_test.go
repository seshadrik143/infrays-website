package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

// To run these tests:
//
//   export TEST_PG_DSN="postgres://postgres:postgres@localhost:5432/issuer_test?sslmode=disable"
//   go test ./internal/store/...
//
// Without TEST_PG_DSN set, all tests in this file skip — matches the
// NodePulse pgstore convention so the test suite runs on laptops
// without a Postgres dependency.

func openPG(t *testing.T) *PG {
	t.Helper()
	dsn := os.Getenv("TEST_PG_DSN")
	if dsn == "" {
		t.Skip("TEST_PG_DSN not set")
	}
	pg, err := NewPG(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewPG: %v", err)
	}
	// Each test starts on a clean slate.
	_, _ = pg.pool.Exec(context.Background(), `
		TRUNCATE TABLE
			licenses, enrollment_tokens, deployments,
			subscriptions, customers, entitlement_sets, admin_users
		RESTART IDENTITY CASCADE
	`)
	t.Cleanup(func() { _ = pg.Close() })
	return pg
}

func TestPG_CustomerRoundTrip(t *testing.T) {
	pg := openPG(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	c := &Customer{
		ID:        "cust_test_001",
		Email:     "test@example.com",
		Name:      "Test User",
		Company:   "ACME Corp",
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := pg.CreateCustomer(context.Background(), c); err != nil {
		t.Fatal(err)
	}
	got, err := pg.GetCustomer(context.Background(), "cust_test_001")
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != c.Email || got.Name != c.Name {
		t.Errorf("got %+v", got)
	}
	dup := *c
	dup.ID = "cust_test_002"
	if err := pg.CreateCustomer(context.Background(), &dup); !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestPG_EnrollmentTokenIdempotentConsume(t *testing.T) {
	pg := openPG(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	_ = pg.CreateCustomer(ctx, &Customer{ID: "c1", Email: "a@b", Status: "active", CreatedAt: now, UpdatedAt: now})
	_ = pg.CreateSubscription(ctx, &Subscription{
		ID: "s1", CustomerID: "c1", Tier: "professional", Status: "active",
		CurrentPeriodEnd: now.Add(365 * 24 * time.Hour), CreatedAt: now, UpdatedAt: now,
	})
	tok := &EnrollmentToken{
		ID: "t1", CustomerID: "c1", SubscriptionID: "s1",
		TokenHash: "hash_abc", CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour),
	}
	if err := pg.CreateEnrollmentToken(ctx, tok); err != nil {
		t.Fatal(err)
	}
	if err := pg.ConsumeEnrollmentToken(ctx, "hash_abc", "dep_1", `{"jws":"first"}`); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if err := pg.ConsumeEnrollmentToken(ctx, "hash_abc", "dep_1", `{"jws":"second"}`); err != nil {
		t.Errorf("idempotent re-consume should succeed, got %v", err)
	}
	if err := pg.ConsumeEnrollmentToken(ctx, "hash_abc", "dep_2", `{}`); !errors.Is(err, ErrConsumed) {
		t.Errorf("different-deployment consume should return ErrConsumed, got %v", err)
	}
	stored, err := pg.GetEnrollmentTokenByHash(ctx, "hash_abc")
	if err != nil {
		t.Fatal(err)
	}
	if stored.ConsumedByDeployment != "dep_1" {
		t.Errorf("ConsumedByDeployment: got %q", stored.ConsumedByDeployment)
	}
	if stored.ConsumedResponseJWS != `{"jws":"first"}` {
		t.Errorf("ConsumedResponseJWS: got %q", stored.ConsumedResponseJWS)
	}
}

func TestPG_LicenseRevokeAndQuery(t *testing.T) {
	pg := openPG(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	_ = pg.CreateCustomer(ctx, &Customer{ID: "c1", Email: "a@b", Status: "active", CreatedAt: now, UpdatedAt: now})
	_ = pg.CreateSubscription(ctx, &Subscription{ID: "s1", CustomerID: "c1", Tier: "professional", Status: "active", CurrentPeriodEnd: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	for i := 0; i < 3; i++ {
		l := &License{
			ID:               fmt.Sprintf("lr_%d", i),
			JTI:              fmt.Sprintf("jti_%d", i),
			LicenseID:        "lic_stable",
			CustomerID:       "c1", SubscriptionID: "s1", DeploymentID: "dep_1",
			EntitlementSetID: "professional-v1", Tier: "professional",
			Iat: now.Add(time.Duration(i) * time.Second), NotBefore: now, ExpiresAt: now.Add(time.Hour),
			GraceUntil: now.Add(2 * time.Hour), Kid: "test-kid", PayloadJWS: "...",
			CreatedAt: now.Add(time.Duration(i) * time.Second),
		}
		if err := pg.CreateLicense(ctx, l); err != nil {
			t.Fatalf("create license %d: %v", i, err)
		}
	}
	latest, err := pg.GetLatestLicenseForLicenseID(ctx, "lic_stable")
	if err != nil {
		t.Fatal(err)
	}
	if latest.JTI != "jti_2" {
		t.Errorf("latest jti: got %q want jti_2", latest.JTI)
	}
	if err := pg.RevokeLicenseByJTI(ctx, "jti_2", "test revoke"); err != nil {
		t.Fatal(err)
	}
	got, _ := pg.GetLicenseByJTI(ctx, "jti_2")
	if !got.Revoked {
		t.Error("expected revoked=true")
	}
	if got.RevokedReason != "test revoke" {
		t.Errorf("reason: got %q", got.RevokedReason)
	}
	if revoked, _ := pg.IsJTIRevoked(ctx, "jti_2"); !revoked {
		t.Error("IsJTIRevoked should be true")
	}
	if revoked, _ := pg.IsJTIRevoked(ctx, "jti_unknown"); revoked {
		t.Error("unknown JTI should not be revoked")
	}
}

func TestPG_DeploymentUpsert(t *testing.T) {
	pg := openPG(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	_ = pg.CreateCustomer(ctx, &Customer{ID: "c1", Email: "a@b", Status: "active", CreatedAt: now, UpdatedAt: now})
	d1 := &Deployment{
		ID: "dr_1", CustomerID: "c1", DeploymentID: "dep_uuid_1",
		DeploymentName: "prod-east", LastVersion: "v0.50.0",
		FirstSeenAt: now, LastSeenAt: now, CreatedAt: now,
	}
	if err := pg.UpsertDeployment(ctx, d1); err != nil {
		t.Fatal(err)
	}
	later := now.Add(time.Hour)
	d2 := &Deployment{
		ID: "dr_2", CustomerID: "c1", DeploymentID: "dep_uuid_1",
		LastVersion: "v0.51.0",
		FirstSeenAt: now, LastSeenAt: later, CreatedAt: now,
	}
	if err := pg.UpsertDeployment(ctx, d2); err != nil {
		t.Fatal(err)
	}
	got, err := pg.GetDeployment(ctx, "dep_uuid_1")
	if err != nil {
		t.Fatal(err)
	}
	if !got.LastSeenAt.Equal(later) {
		t.Errorf("last_seen_at: got %v want %v", got.LastSeenAt, later)
	}
	if got.LastVersion != "v0.51.0" {
		t.Errorf("version: got %q", got.LastVersion)
	}
	if got.DeploymentName != "prod-east" {
		t.Errorf("name preservation: got %q want prod-east", got.DeploymentName)
	}
}

func TestPG_NotFoundShapes(t *testing.T) {
	pg := openPG(t)
	ctx := context.Background()
	if _, err := pg.GetCustomer(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("customer: want ErrNotFound, got %v", err)
	}
	if _, err := pg.GetSubscription(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("subscription: want ErrNotFound, got %v", err)
	}
	if _, err := pg.GetEnrollmentTokenByHash(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("token: want ErrNotFound, got %v", err)
	}
	if _, err := pg.GetLicenseByJTI(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Errorf("license: want ErrNotFound, got %v", err)
	}
}
