package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ProcessGDPRRequest processes a GDPR request
func (h *SecurityHandler) ProcessGDPRRequest(w http.ResponseWriter, r *http.Request) {
	var req GDPRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.Type == "" {
		WriteError(w, nil, "User ID and request type are required", http.StatusBadRequest)
		return
	}

	err := h.gdprService.ProcessRequest(r.Context(), req)
	if err != nil {
		WriteError(w, err, "Failed to process GDPR request", http.StatusInternalServerError)
		return
	}

	gdprResponse := &GDPRResponse{
		Success: true,
		Data:    req,
	}

	WriteJSON(w, gdprResponse, http.StatusOK)
}

// GetGDPRRequestStatus returns GDPR request status
func (h *SecurityHandler) GetGDPRRequestStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract request ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/security/gdpr/status/")
	if path == "" {
		WriteError(w, nil, "Request ID is required", http.StatusBadRequest)
		return
	}

	status, err := h.gdprService.GetRequestStatus(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get GDPR request status", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    status,
	}, http.StatusOK)
}

// RegisterGDPRRoutes registers GDPR-related routes
func (h *SecurityHandler) RegisterGDPRRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/security/gdpr", h.ProcessGDPRRequest)
	mux.HandleFunc("/api/security/gdpr/status/", h.GetGDPRRequestStatus)
}
