package adminportal

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/seshadrik143/infrays-website/backend/internal/obs"
	"github.com/seshadrik143/infrays-website/backend/internal/store"
)

// Bulk endpoints — sized at "ops day" workflows: suspending a batch of
// past-due customers, flagging a wave of suspicious deployments, etc.
//
// Semantics
//   - Operations run sequentially per-id (not transactional). One
//     failed id doesn't roll the rest back; the response reports
//     per-id outcomes so the operator can see what landed.
//   - Hard cap on batch size (BulkMaxItems) to keep request bodies
//     manageable and prevent a runaway action from locking the admin
//     mux for too long.
//   - Audit log gets one entry per successful change, NOT one entry
//     for the whole batch — that way the hash chain captures each
//     individual state transition.

const BulkMaxItems = 200

type bulkResult struct {
	ID    string `json:"id"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// ─── Bulk customer status ───────────────────────────────────────────

type bulkCustomerStatusReq struct {
	IDs    []string `json:"ids"`
	Status string   `json:"status"` // active | suspended | deleted
}

func (s *Server) handleBulkCustomerStatus(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	var req bulkCustomerStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids required")
		return
	}
	if len(req.IDs) > BulkMaxItems {
		writeError(w, http.StatusBadRequest, "too many ids (max 200 per batch)")
		return
	}
	switch req.Status {
	case "active", "suspended", "deleted":
	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}

	caller := adminFromContext(r.Context())
	results := make([]bulkResult, 0, len(req.IDs))
	successCount := 0

	for _, id := range req.IDs {
		cust, err := s.cfg.Store.GetCustomer(r.Context(), id)
		if err != nil {
			results = append(results, bulkResult{ID: id, OK: false, Error: bulkErrMsg(err)})
			continue
		}
		cust.Status = req.Status
		cust.UpdatedAt = s.cfg.Now()
		if err := s.cfg.Store.UpdateCustomer(r.Context(), cust); err != nil {
			results = append(results, bulkResult{ID: id, OK: false, Error: err.Error()})
			continue
		}
		// Suspended / deleted → drop active portal sessions.
		if req.Status != "active" {
			_ = s.cfg.Store.DeletePortalSessionsForCustomer(r.Context(), cust.ID)
		}
		s.appendAudit("admin.bulk_customer_status", caller, map[string]any{
			"customer_id": cust.ID, "status": req.Status,
		})
		results = append(results, bulkResult{ID: id, OK: true})
		successCount++
	}
	obs.AdminActionsTotal.WithLabelValues("bulk_customer_status").Inc()
	writeJSON(w, http.StatusOK, map[string]any{
		"results":  results,
		"success":  successCount,
		"failed":   len(req.IDs) - successCount,
		"status":   req.Status,
	})
}

// ─── Bulk deployment flag ───────────────────────────────────────────

type bulkDeploymentFlagReq struct {
	DeploymentIDs []string `json:"deployment_ids"`
	Flagged       bool     `json:"flagged"`
	Reason        string   `json:"reason"`
}

func (s *Server) handleBulkDeploymentFlag(w http.ResponseWriter, r *http.Request) {
	if !s.requireRoleAdmin(w, r) {
		return
	}
	var req bulkDeploymentFlagReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if len(req.DeploymentIDs) == 0 {
		writeError(w, http.StatusBadRequest, "deployment_ids required")
		return
	}
	if len(req.DeploymentIDs) > BulkMaxItems {
		writeError(w, http.StatusBadRequest, "too many ids (max 200 per batch)")
		return
	}
	if req.Flagged && req.Reason == "" {
		writeError(w, http.StatusBadRequest, "reason required when flagging")
		return
	}

	caller := adminFromContext(r.Context())
	results := make([]bulkResult, 0, len(req.DeploymentIDs))
	successCount := 0

	for _, depID := range req.DeploymentIDs {
		dep, err := s.cfg.Store.GetDeployment(r.Context(), depID)
		if err != nil {
			results = append(results, bulkResult{ID: depID, OK: false, Error: bulkErrMsg(err)})
			continue
		}
		dep.FlaggedForReview = req.Flagged
		dep.FlagReason = req.Reason
		if err := s.cfg.Store.UpdateDeployment(r.Context(), dep); err != nil {
			results = append(results, bulkResult{ID: depID, OK: false, Error: err.Error()})
			continue
		}
		s.appendAudit("admin.bulk_deployment_flag", caller, map[string]any{
			"deployment_id": depID, "flagged": req.Flagged, "reason": req.Reason,
		})
		results = append(results, bulkResult{ID: depID, OK: true})
		successCount++
	}
	obs.AdminActionsTotal.WithLabelValues("bulk_deployment_flag").Inc()
	writeJSON(w, http.StatusOK, map[string]any{
		"results": results,
		"success": successCount,
		"failed":  len(req.DeploymentIDs) - successCount,
	})
}

func bulkErrMsg(err error) string {
	if errors.Is(err, store.ErrNotFound) {
		return "not found"
	}
	return err.Error()
}
