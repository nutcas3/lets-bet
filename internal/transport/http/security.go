package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	audit "github.com/betting-platform/internal/security/audit"
	gdpr "github.com/betting-platform/internal/security/gdpr"
	pentest "github.com/betting-platform/internal/security/pentest"
	responsiblegaming "github.com/betting-platform/internal/security/responsiblegaming"
)

// SecurityHandler handles security-related HTTP requests
type SecurityHandler struct {
	auditService   *audit.SecurityAuditService
	pentestService *pentest.PenetrationTestService
	gdprService    *gdpr.GDPRService
	rgService      *responsiblegaming.ResponsibleGamingService
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(
	auditService *audit.SecurityAuditService,
	pentestService *pentest.PenetrationTestService,
	gdprService *gdpr.GDPRService,
	rgService *responsiblegaming.ResponsibleGamingService,
) *SecurityHandler {
	return &SecurityHandler{
		auditService:   auditService,
		pentestService: pentestService,
		gdprService:    gdprService,
		rgService:      rgService,
	}
}

// PerformSecurityAudit performs a comprehensive security audit
func (h *SecurityHandler) PerformSecurityAudit(w http.ResponseWriter, r *http.Request) {
	var req SecurityAuditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	audit, err := h.auditService.PerformSecurityAudit(r.Context())
	if err != nil {
		WriteError(w, err, "Failed to perform security audit", http.StatusInternalServerError)
		return
	}

	response := &SecurityAuditResponse{
		Success: true,
		Data:    audit,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetSecurityAuditHistory returns security audit history
func (h *SecurityHandler) GetSecurityAuditHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history, err := h.auditService.GetAuditHistory(ctx, limit)
	if err != nil {
		WriteError(w, err, "Failed to get audit history", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    history,
	}, http.StatusOK)
}

// PerformPenetrationTest performs a penetration test
func (h *SecurityHandler) PerformPenetrationTest(w http.ResponseWriter, r *http.Request) {
	var req PenetrationTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	test, err := h.pentestService.PerformPenetrationTest(r.Context(), pentest.TestType(req.TestType), req.Scope)
	if err != nil {
		WriteError(w, err, "Failed to perform penetration test", http.StatusInternalServerError)
		return
	}

	response := &PenetrationTestResponse{
		Success: true,
		Data:    test,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetPenetrationTestResults returns penetration test results
func (h *SecurityHandler) GetPenetrationTestResults(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract test ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/security/pentest/results/")
	if path == "" {
		WriteError(w, nil, "Test ID is required", http.StatusBadRequest)
		return
	}

	results, err := h.pentestService.GetTestResults(ctx, path)
	if err != nil {
		WriteError(w, err, "Failed to get penetration test results", http.StatusNotFound)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    results,
	}, http.StatusOK)
}

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

// GetSecurityMetrics returns security metrics
func (h *SecurityHandler) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.auditService.GetSecurityMetrics(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get security metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    metrics,
	}, http.StatusOK)
}

// GetVulnerabilityReport returns vulnerability report
func (h *SecurityHandler) GetVulnerabilityReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	report, err := h.auditService.GetVulnerabilityReport(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get vulnerability report", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    report,
	}, http.StatusOK)
}

// RegisterRoutes registers security routes
func (h *SecurityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/security/audit", h.PerformSecurityAudit)
	mux.HandleFunc("/api/security/audit/history", h.GetSecurityAuditHistory)
	mux.HandleFunc("/api/security/pentest", h.PerformPenetrationTest)
	mux.HandleFunc("/api/security/pentest/results/", h.GetPenetrationTestResults)
	mux.HandleFunc("/api/security/responsible-gaming", h.ProcessResponsibleGamingRequest)
	mux.HandleFunc("/api/security/responsible-gaming/profile/", h.GetUserGamingProfile)
	mux.HandleFunc("/api/security/responsible-gaming/limits", h.SetGamingLimits)
	mux.HandleFunc("/api/security/metrics", h.GetSecurityMetrics)
	mux.HandleFunc("/api/security/vulnerabilities", h.GetVulnerabilityReport)
}
