package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/betting-platform/internal/admin"
)

// GetUserManagementData returns user management data with pagination
func (h *AdminHandler) GetUserManagementData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort_by")
	sortDir := r.URL.Query().Get("sort_dir")

	limit := 20 // default
	offset := 0 // default

	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	request := &admin.UserManagementRequest{
		Limit:   limit,
		Offset:  offset,
		Search:  search,
		Status:  status,
		SortBy:  sortBy,
		SortDir: sortDir,
	}

	userData, err := h.adminService.GetUserManagementData(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get user management data", http.StatusInternalServerError)
		return
	}

	response := &UserManagementResponse{
		Success: true,
		Data:    userData,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetUserDetails returns detailed information about a specific user
func (h *AdminHandler) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/admin/users/")
	if path == "" {
		WriteError(w, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	userDetails, err := h.adminService.GetUserDetails(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get user details", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    userDetails,
	}, http.StatusOK)
}

// PerformUserAction performs an action on a user (ban, unban, verify, etc.)
func (h *AdminHandler) PerformUserAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UserActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.UserID == "" || req.Action == "" {
		WriteError(w, nil, "User ID and action are required", http.StatusBadRequest)
		return
	}

	actionReq := &admin.UserActionRequest{
		UserID: req.UserID,
		Action: req.Action,
		Reason: req.Reason,
	}

	err := h.adminService.PerformUserAction(ctx, actionReq)
	if err != nil {
		WriteError(w, err, "Failed to perform user action", http.StatusInternalServerError)
		return
	}

	response := &UserActionResponse{
		Success: true,
		Message: "User action completed successfully",
		UserID:  req.UserID,
	}

	WriteJSON(w, response, http.StatusOK)
}

// RegisterUserRoutes registers user management routes
func (h *AdminHandler) RegisterUserRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/admin/users", h.GetUserManagementData)
	mux.HandleFunc("/api/admin/users/", h.GetUserDetails)
	mux.HandleFunc("/api/admin/users/action", h.PerformUserAction)
}
