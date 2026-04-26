package http

import (
	"encoding/json"
	"net/http"
)

// GetComplianceAlerts returns compliance alerts
func (h *ComplianceHandler) GetComplianceAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	alerts, err := h.bclbService.GetComplianceAlerts(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance alerts", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    alerts,
	}, http.StatusOK)
}

// AcknowledgeComplianceAlert acknowledges a compliance alert
func (h *ComplianceHandler) AcknowledgeComplianceAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID string `json:"alert_id"`
		Notes   string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.AcknowledgeComplianceAlert(r.Context(), req.AlertID)
	if err != nil {
		WriteError(w, err, "Failed to acknowledge compliance alert", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "Compliance alert acknowledged successfully",
	}, http.StatusOK)
}

// ResolveComplianceAlert resolves a compliance alert
func (h *ComplianceHandler) ResolveComplianceAlert(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertID    string `json:"alert_id"`
		Resolution string `json:"resolution"`
		Notes      string `json:"notes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.ResolveComplianceAlert(r.Context(), req.AlertID)
	if err != nil {
		WriteError(w, err, "Failed to resolve compliance alert", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "Compliance alert resolved successfully",
	}, http.StatusOK)
}

// RegisterAlertRoutes registers alert-related routes
func (h *ComplianceHandler) RegisterAlertRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/compliance/alerts", h.GetComplianceAlerts)
	mux.HandleFunc("/api/compliance/alerts/acknowledge", h.AcknowledgeComplianceAlert)
	mux.HandleFunc("/api/compliance/alerts/resolve", h.ResolveComplianceAlert)
}
