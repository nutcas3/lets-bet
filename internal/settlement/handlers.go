package settlement

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	// Add dependencies here (settlement service, repositories, etc.)
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", h.healthCheckHandler).Methods("GET")

	// Internal API (called by other services)
	api := r.PathPrefix("/internal/v1").Subrouter()
	api.HandleFunc("/settlements/pending", h.getPendingSettlementsHandler).Methods("GET")
	api.HandleFunc("/settlements/process", h.processSettlementsHandler).Methods("POST")
	api.HandleFunc("/settlements/{id}", h.getSettlementHandler).Methods("GET")
}

func (h *Handler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"settlement"}`))
}

func (h *Handler) getPendingSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"pending_settlements":[{"id":"1","bet_id":"123","status":"pending"}]}`))
}

func (h *Handler) processSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"Settlement processing initiated"}`))
}

func (h *Handler) getSettlementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	settlementID := vars["id"]
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"id":"` + settlementID + `","status":"completed","amount":1000.00}`))
}
