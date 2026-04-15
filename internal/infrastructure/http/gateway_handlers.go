package http

import (
	"net/http"

	"github.com/betting-platform/internal/core/usecase/games"
	"github.com/betting-platform/internal/core/usecase/odds"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// GatewayHandler handles all HTTP requests for the API gateway
type GatewayHandler struct {
	gamesEngine *games.CrashGameEngine
	oddsEngine  *odds.OddsEngine
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, validate origin
	},
}

func NewGatewayHandler(gamesEngine *games.CrashGameEngine, oddsEngine *odds.OddsEngine) *GatewayHandler {
	return &GatewayHandler{
		gamesEngine: gamesEngine,
		oddsEngine:  oddsEngine,
	}
}

func (h *GatewayHandler) RegisterRoutes(r *mux.Router) {
	// Health check
	r.HandleFunc("/health", h.healthCheckHandler).Methods("GET")

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Auth routes
	api.HandleFunc("/auth/register", h.registerHandler).Methods("POST")
	api.HandleFunc("/auth/login", h.loginHandler).Methods("POST")

	// User routes (protected)
	api.HandleFunc("/users/me", h.getUserHandler).Methods("GET")
	api.HandleFunc("/users/me/wallet", h.getWalletHandler).Methods("GET")

	// Betting routes
	api.HandleFunc("/bets", h.placeBetHandler).Methods("POST")
	api.HandleFunc("/bets/{id}", h.getBetHandler).Methods("GET")
	api.HandleFunc("/bets/history", h.getBetHistoryHandler).Methods("GET")

	// Payment routes
	api.HandleFunc("/payments/deposit", h.depositHandler).Methods("POST")
	api.HandleFunc("/payments/withdraw", h.withdrawHandler).Methods("POST")
	api.HandleFunc("/payments/mpesa/callback", h.mpesaCallbackHandler).Methods("POST")

	// Games routes
	api.HandleFunc("/games/crash/current", h.getCurrentGameHandler).Methods("GET")
	api.HandleFunc("/games/crash/history", h.getGameHistoryHandler).Methods("GET")
	api.HandleFunc("/games/crash/bet", h.placeGameBetHandler).Methods("POST")

	// Odds routes
	api.HandleFunc("/odds/live", h.getLiveOddsHandler).Methods("GET")
	api.HandleFunc("/odds/calculate", h.calculateOddsHandler).Methods("POST")

	// WebSocket endpoint for live games
	r.HandleFunc("/ws/games/{gameId}", h.handleWebSocket)

	// CORS middleware
	r.Use(h.CorsMiddleware)
}

func (h *GatewayHandler) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"gateway"}`))
}

func (h *GatewayHandler) registerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Registration endpoint - implement with KYC validation"}`))
}

func (h *GatewayHandler) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Login endpoint - implement with JWT token generation"}`))
}

func (h *GatewayHandler) getUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get user profile - implement with auth middleware"}`))
}

func (h *GatewayHandler) getWalletHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"balance":1000.00,"currency":"KES","bonus_balance":50.00}`))
}

func (h *GatewayHandler) placeBetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Place bet endpoint - implement with PlaceBetUseCase"}`))
}

func (h *GatewayHandler) getBetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get bet details"}`))
}

func (h *GatewayHandler) getBetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Get bet history"}`))
}

func (h *GatewayHandler) depositHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"Deposit initiated - check M-Pesa prompt on phone"}`))
}

func (h *GatewayHandler) withdrawHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"Withdrawal initiated - M-Pesa payout in progress"}`))
}

func (h *GatewayHandler) mpesaCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Handle M-Pesa callbacks (STK Push result, B2C result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"ResultCode":0,"ResultDesc":"Accepted"}`))
}

func (h *GatewayHandler) getCurrentGameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"game_id":"123","round_number":42,"status":"RUNNING","current_multiplier":2.45}`))
}

func (h *GatewayHandler) getGameHistoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"games":[{"round":41,"crash_point":3.52},{"round":40,"crash_point":1.23}]}`))
}

func (h *GatewayHandler) placeGameBetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Game bet placed - connect to WebSocket for live updates"}`))
}

func (h *GatewayHandler) getLiveOddsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"matches":[{"id":"1","home":"Arsenal","away":"Chelsea","odds":{"home":2.50,"draw":3.20,"away":2.80}}]}`))
}

func (h *GatewayHandler) calculateOddsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"total_odds":8.50,"potential_win":850.00}`))
}

func (h *GatewayHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusBadRequest)
		return
	}

	defer conn.Close()

	// Send welcome message
	conn.WriteJSON(map[string]any{
		"type":    "connected",
		"message": "Connected to crash game",
	})

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *GatewayHandler) CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
