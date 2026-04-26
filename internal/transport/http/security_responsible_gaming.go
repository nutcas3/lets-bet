package http

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ProcessResponsibleGamingRequest processes a responsible gaming request
func (h *SecurityHandler) ProcessResponsibleGamingRequest(w http.ResponseWriter, r *http.Request) {
	var req ResponsibleGamingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.Action == "" {
		WriteError(w, nil, "User ID and action are required", http.StatusBadRequest)
		return
	}

	err := h.rgService.ProcessRequest(r.Context(), req)
	if err != nil {
		WriteError(w, err, "Failed to process responsible gaming request", http.StatusInternalServerError)
		return
	}

	rgResponse := &ResponsibleGamingResponse{
		Success: true,
		Data:    req,
	}

	WriteJSON(w, rgResponse, http.StatusOK)
}

// GetUserGamingProfile returns user gaming profile
func (h *SecurityHandler) GetUserGamingProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/security/responsible-gaming/profile/")
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	profile, err := h.rgService.GetUserGamingProfile(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get user gaming profile", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    profile,
	}, http.StatusOK)
}

// SetGamingLimits sets gaming limits for a user
func (h *SecurityHandler) SetGamingLimits(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string         `json:"user_id"`
		Limits []*GamingLimit `json:"limits"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || len(req.Limits) == 0 {
		WriteError(w, nil, "User ID and limits are required", http.StatusBadRequest)
		return
	}

	err := h.rgService.SetGamingLimits(r.Context(), req.UserID, req.Limits)
	if err != nil {
		WriteError(w, err, "Failed to set gaming limits", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"message": "Gaming limits set successfully",
	}, http.StatusOK)
}

// RegisterResponsibleGamingRoutes registers responsible gaming routes
func (h *SecurityHandler) RegisterResponsibleGamingRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/security/responsible-gaming", h.ProcessResponsibleGamingRequest)
	mux.HandleFunc("/api/security/responsible-gaming/profile/", h.GetUserGamingProfile)
	mux.HandleFunc("/api/security/responsible-gaming/limits", h.SetGamingLimits)
}
