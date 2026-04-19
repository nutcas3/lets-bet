package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// SettlementHandler handles settlement service HTTP requests
type SettlementHandler struct{}

func NewSettlementHandler() *SettlementHandler {
	return &SettlementHandler{}
}

func (h *SettlementHandler) RegisterRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", h.healthCheckHandler).Methods("GET")

	// Internal API (called by other services)
	api := r.PathPrefix("/internal/v1").Subrouter()
	api.HandleFunc("/settlements/pending", h.getPendingSettlementsHandler).Methods("GET")
	api.HandleFunc("/settlements/process", h.processSettlementsHandler).Methods("POST")
	api.HandleFunc("/settlements/{id}", h.getSettlementHandler).Methods("GET")
}

func (h *SettlementHandler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"settlement"}`))
}

func (h *SettlementHandler) getPendingSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"pending_settlements":[{"id":"1","bet_id":"123","status":"pending"}]}`))
}

func (h *SettlementHandler) processSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"Settlement processing initiated"}`))
}

func (h *SettlementHandler) getSettlementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID := vars["id"]

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     settlementID,
		"status": "completed",
		"amount": 1000.00,
	}); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}
