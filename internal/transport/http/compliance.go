package http

import (
	"encoding/json"
	"net/http"
	"time"
)

// ValidateBetPlacement validates a bet placement request
func (h *ComplianceHandler) ValidateBetPlacement(w http.ResponseWriter, r *http.Request) {
	var req BetValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	check, err := h.bclbService.ValidateBetPlacement(r.Context(), req.UserID, req.BetAmount, req.BetType, req.Selections)
	if err != nil {
		WriteError(w, err, "Failed to validate bet placement", http.StatusInternalServerError)
		return
	}

	response := &BetValidationResponse{
		Success: true,
		Data:    check,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetUserComplianceStatus returns user compliance status
func (h *ComplianceHandler) GetUserComplianceStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/status/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	status, err := h.bclbService.GetUserComplianceStatus(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get user compliance status", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    status,
	}, http.StatusOK)
}

// SetUserLimits sets user-specific limits
func (h *ComplianceHandler) SetUserLimits(w http.ResponseWriter, r *http.Request) {
	var req UserLimits
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/limits/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.SetUserLimits(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to set user limits", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "User limits set successfully",
	}, http.StatusOK)
}

// AddUserRestriction adds a restriction to a user
func (h *ComplianceHandler) AddUserRestriction(w http.ResponseWriter, r *http.Request) {
	var req UserRestriction
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract user ID from URL path
	path := r.URL.Path[len("/api/compliance/user/restrictions/"):]
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.AddUserRestriction(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to add user restriction", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "User restriction added successfully",
	}, http.StatusOK)
}

// GetComplianceMetrics returns compliance metrics
func (h *ComplianceHandler) GetComplianceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	var fromDate, toDate *time.Time
	if fromDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, fromDateStr); err == nil {
			fromDate = &parsed
		}
	}
	if toDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, toDateStr); err == nil {
			toDate = &parsed
		}
	}

	period := &TimePeriod{}
	if fromDate != nil {
		period.From = *fromDate
	}
	if toDate != nil {
		period.To = *toDate
	}

	metrics, err := h.bclbService.GetComplianceMetrics(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    metrics,
	}, http.StatusOK)
}

// GetComplianceSettings returns compliance settings
func (h *ComplianceHandler) GetComplianceSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	settings, err := h.bclbService.GetComplianceSettings(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance settings", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    settings,
	}, http.StatusOK)
}

// UpdateComplianceSettings updates compliance settings
func (h *ComplianceHandler) UpdateComplianceSettings(w http.ResponseWriter, r *http.Request) {
	var req ComplianceSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.UpdateComplianceSettings(r.Context(), &req)
	if err != nil {
		WriteError(w, err, "Failed to update compliance settings", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "Compliance settings updated successfully",
	}, http.StatusOK)
}

// RegisterRoutes registers compliance routes
func (h *ComplianceHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/compliance/bet/validate", h.ValidateBetPlacement)
	mux.HandleFunc("/api/compliance/user/status/", h.GetUserComplianceStatus)
	mux.HandleFunc("/api/compliance/user/limits/", h.SetUserLimits)
	mux.HandleFunc("/api/compliance/user/restrictions/", h.AddUserRestriction)
	mux.HandleFunc("/api/compliance/metrics", h.GetComplianceMetrics)
	mux.HandleFunc("/api/compliance/settings", h.GetComplianceSettings)
	mux.HandleFunc("/api/compliance/settings/update", h.UpdateComplianceSettings)
}
