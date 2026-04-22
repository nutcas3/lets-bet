package http

import (
	"net/http"

	"github.com/betting-platform/internal/core/usecase/odds"
	"github.com/gorilla/mux"
)

// EngineHandler handles odds engine HTTP requests
type EngineHandler struct {
	oddsEngine *odds.OddsEngine
}

func NewEngineHandler(oddsEngine *odds.OddsEngine) *EngineHandler {
	return &EngineHandler{
		oddsEngine: oddsEngine,
	}
}

func (h *EngineHandler) RegisterRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", h.healthCheckHandler).Methods("GET")

	// Internal API (called by other services)
	api := r.PathPrefix("/internal/v1").Subrouter()
	api.HandleFunc("/odds/sync", h.syncOddsHandler).Methods("POST")
	api.HandleFunc("/odds/calculate", h.calculateOddsHandler).Methods("POST")
	api.HandleFunc("/odds/live", h.getLiveOddsHandler).Methods("GET")
}

func (h *EngineHandler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy","service":"engine"}`))
}

func (h *EngineHandler) syncOddsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"message":"Odds sync initiated"}`))
}

func (h *EngineHandler) calculateOddsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"total_odds":8.50,"potential_win":850.00}`))
}

func (h *EngineHandler) getLiveOddsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"matches":[{"id":"1","home":"Arsenal","away":"Chelsea","odds":{"home":2.50,"draw":3.20,"away":2.80}}]}`))
}
