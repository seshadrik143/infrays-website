package adminportal_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

func TestBulkCustomerStatusForbiddenForNonAdmin(t *testing.T) {
	h := newHarness(t)
	secret := seedAdminWithRole(t, h, "support@x.com", "supportpw1234", "support")
	c := h.loginAndVerifyMFA("support@x.com", "supportpw1234", secret)
	resp, _ := h.do(c, "POST", "/api/admin/customers/bulk-status", map[string]any{
		"ids": []string{"x"}, "status": "suspended",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestBulkCustomerStatusSuspendsMultiple(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpw1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpw1234", secret)

	now := h.now()
	for _, e := range []string{"a@x.com", "b@x.com", "c@x.com"} {
		_ = h.store.CreateCustomer(context.Background(), &store.Customer{
			ID: "cust_" + e, Email: e, Status: "active", CreatedAt: now, UpdatedAt: now,
		})
		// Pretend each has an active portal session.
		_ = h.store.CreatePortalSession(context.Background(), &store.PortalSession{
			ID: "psess_" + e, CustomerID: "cust_" + e,
			CreatedAt: now, LastSeen: now, ExpiresAt: now.Add(time.Hour),
		})
	}

	resp, body := h.do(c, "POST", "/api/admin/customers/bulk-status", map[string]any{
		"ids":    []string{"cust_a@x.com", "cust_b@x.com", "cust_c@x.com"},
		"status": "suspended",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", resp.StatusCode, body)
	}
	var got struct {
		Success int `json:"success"`
		Failed  int `json:"failed"`
	}
	_ = json.Unmarshal(body, &got)
	if got.Success != 3 || got.Failed != 0 {
		t.Fatalf("expected 3/0, got %d/%d", got.Success, got.Failed)
	}

	// Confirm sessions dropped + statuses changed.
	for _, e := range []string{"a@x.com", "b@x.com", "c@x.com"} {
		cust, _ := h.store.GetCustomer(context.Background(), "cust_"+e)
		if cust.Status != "suspended" {
			t.Fatalf("%s not suspended: %+v", e, cust)
		}
		if _, err := h.store.GetPortalSession(context.Background(), "psess_"+e); err == nil {
			t.Fatalf("portal session for %s should be dropped", e)
		}
	}
}

func TestBulkCustomerStatusPartialFailure(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpw1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpw1234", secret)

	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "real_cust", Email: "real@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})

	resp, body := h.do(c, "POST", "/api/admin/customers/bulk-status", map[string]any{
		"ids":    []string{"real_cust", "missing_cust"},
		"status": "suspended",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var got struct {
		Success int `json:"success"`
		Failed  int `json:"failed"`
		Results []struct {
			ID    string `json:"id"`
			OK    bool   `json:"ok"`
			Error string `json:"error"`
		} `json:"results"`
	}
	_ = json.Unmarshal(body, &got)
	if got.Success != 1 || got.Failed != 1 {
		t.Fatalf("expected 1/1, got %d/%d", got.Success, got.Failed)
	}
	// missing_cust should report error.
	for _, r := range got.Results {
		if r.ID == "missing_cust" && (r.OK || r.Error == "") {
			t.Fatalf("expected missing_cust to have error, got %+v", r)
		}
	}
}

func TestBulkCustomerStatusRejectsHugeBatch(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpw1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpw1234", secret)
	ids := make([]string, 201)
	for i := range ids {
		ids[i] = "x"
	}
	resp, _ := h.do(c, "POST", "/api/admin/customers/bulk-status", map[string]any{
		"ids": ids, "status": "suspended",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for over-size batch, got %d", resp.StatusCode)
	}
}

func TestBulkDeploymentFlagFlagsMultiple(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpw1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpw1234", secret)
	now := h.now()
	_ = h.store.CreateCustomer(context.Background(), &store.Customer{
		ID: "cust1", Email: "c@x.com", Status: "active", CreatedAt: now, UpdatedAt: now,
	})
	for _, did := range []string{"d1", "d2"} {
		_ = h.store.UpsertDeployment(context.Background(), &store.Deployment{
			ID: "row_" + did, CustomerID: "cust1", DeploymentID: did,
			FirstSeenAt: now, LastSeenAt: now, CreatedAt: now,
		})
	}

	resp, body := h.do(c, "POST", "/api/admin/deployments/bulk-flag", map[string]any{
		"deployment_ids": []string{"d1", "d2"},
		"flagged":        true,
		"reason":         "ip drift",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", resp.StatusCode, body)
	}
	for _, did := range []string{"d1", "d2"} {
		dep, _ := h.store.GetDeployment(context.Background(), did)
		if !dep.FlaggedForReview || dep.FlagReason != "ip drift" {
			t.Fatalf("%s not flagged correctly: %+v", did, dep)
		}
	}
}

func TestBulkDeploymentFlagRequiresReasonWhenFlagging(t *testing.T) {
	h := newHarness(t)
	_, secret := h.seedAdmin("admin@x.com", "rightpw1234", true)
	c := h.loginAndVerifyMFA("admin@x.com", "rightpw1234", secret)
	resp, _ := h.do(c, "POST", "/api/admin/deployments/bulk-flag", map[string]any{
		"deployment_ids": []string{"d1"},
		"flagged":        true,
		"reason":         "",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
