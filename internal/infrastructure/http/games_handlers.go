package http

import (
	"net/http"

	"github.com/betting-platform/internal/core/usecase/games"
	"github.com/gorilla/mux"
)

// GamesHandler handles games service HTTP requests
type GamesHandler struct {
	gamesEngine *games.CrashGameEngine
}

func NewGamesHandler(gamesEngine *games.CrashGameEngine) *GamesHandler {
	return &GamesHandler{
		gamesEngine: gamesEngine,
	}
}

func (h *GamesHandler) RegisterRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", h.healthCheckHandler).Methods("GET")

	// Game API
	api := r.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/games/crash/current", h.getCurrentGameHandler).Methods("GET")
	api.HandleFunc("/games/crash/history", h.getGameHistoryHandler).Methods("GET")
	api.HandleFunc("/games/crash/bet", h.placeGameBetHandler).Methods("POST")

	// WebSocket endpoint for live games
	r.HandleFunc("/ws/games/{gameId}", h.handleWebSocket)
}

func (h *GamesHandler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"games"}`))
}

func (h *GamesHandler) getCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"game_id":"123","round_number":42,"status":"RUNNING","current_multiplier":2.45}`))
}

func (h *GamesHandler) getGameHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"games":[{"round":41,"crash_point":3.52},{"round":40,"crash_point":1.23}]}`))
}

func (h *GamesHandler) placeGameBetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Game bet placed - connect to WebSocket for live updates"}`))
}

func (h *GamesHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// This would be implemented with the WebSocket hub
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"WebSocket endpoint - implement with hub"}`))
}
