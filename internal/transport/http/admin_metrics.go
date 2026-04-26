package http

import (
	"net/http"
	"time"

	"github.com/betting-platform/internal/admin"
)

// GetBettingMetrics returns betting metrics and analytics
func (h *AdminHandler) GetBettingMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	timeRange := r.URL.Query().Get("time_range")
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

	request := &admin.BettingMetricsRequest{
		TimeRange: timeRange,
		FromDate:  fromDate,
		ToDate:    toDate,
	}

	metrics, err := h.adminService.GetBettingMetrics(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get betting metrics", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    metrics,
	}, http.StatusOK)
}

// GetFinancialReports returns financial reports and analytics
func (h *AdminHandler) GetFinancialReports(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	reportType := r.URL.Query().Get("report_type")
	timeRange := r.URL.Query().Get("time_range")
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

	request := &admin.FinancialReportRequest{
		ReportType: reportType,
		TimeRange:  timeRange,
		FromDate:   fromDate,
		ToDate:     toDate,
	}

	report, err := h.adminService.GetFinancialReport(ctx, request)
	if err != nil {
		WriteError(w, err, "Failed to get financial report", http.StatusInternalServerError)
		return
	}

	WriteJSON(w, map[string]any{
		"success": true,
		"data":    report,
	}, http.StatusOK)
}

// RegisterMetricsRoutes registers metrics-related routes
func (h *AdminHandler) RegisterMetricsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/admin/betting/metrics", h.GetBettingMetrics)
	mux.HandleFunc("/api/admin/financial/reports", h.GetFinancialReports)
}
