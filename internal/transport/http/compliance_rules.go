package http

import (
	"encoding/json"
	"net/http"
)

// GetComplianceRules returns compliance rules
func (h *ComplianceHandler) GetComplianceRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rules, err := h.bclbService.GetComplianceRules(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get compliance rules", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    rules,
	}, http.StatusOK)
}

// CreateComplianceRule creates a new compliance rule
func (h *ComplianceHandler) CreateComplianceRule(w http.ResponseWriter, r *http.Request) {
	var req ComplianceRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.bclbService.CreateComplianceRule(r.Context(), &req)
	if err != nil {
		WriteError(w, err, "Failed to create compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    req,
		"message": "Compliance rule created successfully",
	}, http.StatusCreated)
}

// UpdateComplianceRule updates an existing compliance rule
func (h *ComplianceHandler) UpdateComplianceRule(w http.ResponseWriter, r *http.Request) {
	var req ComplianceRule
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract rule ID from URL path
	path := r.URL.Path[len("/api/compliance/rules/"):]
	if path == "" {
		WriteError(w, nil, "Rule ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.UpdateComplianceRule(r.Context(), path, &req)
	if err != nil {
		WriteError(w, err, "Failed to update compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    req,
		"message": "Compliance rule updated successfully",
	}, http.StatusOK)
}

// DeleteComplianceRule deletes a compliance rule
func (h *ComplianceHandler) DeleteComplianceRule(w http.ResponseWriter, r *http.Request) {
	// Extract rule ID from URL path
	path := r.URL.Path[len("/api/compliance/rules/"):]
	if path == "" {
		WriteError(w, nil, "Rule ID is required", http.StatusBadRequest)
		return
	}

	err := h.bclbService.DeleteComplianceRule(r.Context(), path)
	if err != nil {
		WriteError(w, err, "Failed to delete compliance rule", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "Compliance rule deleted successfully",
	}, http.StatusOK)
}

// RegisterRuleRoutes registers rule-related routes
func (h *ComplianceHandler) RegisterRuleRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/compliance/rules", h.GetComplianceRules)
	mux.HandleFunc("/api/compliance/rules/", h.UpdateComplianceRule)
	mux.HandleFunc("/api/compliance/rules/create", h.CreateComplianceRule)
	mux.HandleFunc("/api/compliance/rules/delete/", h.DeleteComplianceRule)
}
