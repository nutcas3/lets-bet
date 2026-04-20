package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/sports/live"
	"github.com/shopspring/decimal"
)

// LiveHandler handles live betting HTTP requests
type LiveHandler struct {
	liveService *live.LiveBettingService
}

// NewLiveHandler creates a new live handler
func NewLiveHandler(liveService *live.LiveBettingService) *LiveHandler {
	return &LiveHandler{
		liveService: liveService,
	}
}

var httpLiveGenerator *id.SnowflakeGenerator

func init() {
	var err error
	httpLiveGenerator, err = id.ServiceTypeGenerator("http")
	if err != nil {
		panic(fmt.Sprintf("Failed to create HTTP live ID generator: %v", err))
	}
}

// Helper functions
func generateID() string {
	return httpLiveGenerator.GenerateID()
}

func getUserID(_ context.Context) string {
	// In a real implementation, this would extract user ID from JWT token
	// For now, return a dummy ID
	return "user-" + generateID()
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, err error, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]any{
		"error":   message,
		"details": err.Error(),
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(response)
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// Request/Response types for live betting API

// LiveBetRequest represents a live betting request
type LiveBetRequest struct {
	MatchID   string          `json:"match_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
}

// LiveBetResponse represents a live betting response
type LiveBetResponse struct {
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
	BetID        string          `json:"bet_id,omitempty"`
	Amount       decimal.Decimal `json:"amount"`
	Odds         decimal.Decimal `json:"odds"`
	PotentialWin decimal.Decimal `json:"potential_win"`
}

// GetLiveMatchResponse represents a live match response
type GetLiveMatchResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    *LiveMatch `json:"data,omitempty"`
}

// LiveMatch represents a live match for API response
type LiveMatch struct {
	ID               string       `json:"id"`
	Sport            string       `json:"sport"`
	HomeTeam         string       `json:"home_team"`
	AwayTeam         string       `json:"away_team"`
	Status           string       `json:"status"`
	StartTime        time.Time    `json:"start_time"`
	CurrentMinute    int          `json:"current_minute"`
	HomeScore        int          `json:"home_score"`
	AwayScore        int          `json:"away_score"`
	HomePossession   float64      `json:"home_possession"`
	AwayPossession   float64      `json:"away_possession"`
	HomeCorners      int          `json:"home_corners"`
	AwayCorners      int          `json:"away_corners"`
	HomeYellowCards  int          `json:"home_yellow_cards"`
	AwayYellowCards  int          `json:"away_yellow_cards"`
	HomeRedCards     int          `json:"home_red_cards"`
	AwayRedCards     int          `json:"away_red_cards"`
	LiveMarkets      []LiveMarket `json:"live_markets"`
	LastUpdated      time.Time    `json:"last_updated"`
	IsSuspended      bool         `json:"is_suspended"`
	SuspensionReason string       `json:"suspension_reason,omitempty"`
}

// LiveMarket represents a live market for API response
type LiveMarket struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	IsLive           bool          `json:"is_live"`
	LiveOdds         []LiveOutcome `json:"live_odds"`
	LastOddsUpdate   time.Time     `json:"last_odds_update"`
	OddsUpdateCount  int           `json:"odds_update_count"`
	IsSuspended      bool          `json:"is_suspended"`
	SuspensionReason string        `json:"suspension_reason,omitempty"`
}

// LiveOutcome represents a live outcome for API response
type LiveOutcome struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CurrentOdds     decimal.Decimal `json:"current_odds"`
	PreviousOdds    decimal.Decimal `json:"previous_odds"`
	OddsChangeTime  time.Time       `json:"odds_change_time"`
	OddsChangeCount int             `json:"odds_change_count"`
	TotalVolume     decimal.Decimal `json:"total_volume"`
	LiveVolume      decimal.Decimal `json:"live_volume"`
}

// GetLiveMatchesResponse represents a response for getting live matches
type GetLiveMatchesResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    []LiveMatch `json:"data,omitempty"`
}

// OddsUpdateRequest represents an odds update request
type OddsUpdateRequest struct {
	MatchID   string          `json:"match_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	NewOdds   decimal.Decimal `json:"new_odds"`
}

// SuspendMatchRequest represents a suspend match request
type SuspendMatchRequest struct {
	MatchID string `json:"match_id"`
	Reason  string `json:"reason"`
}

// LiveMetricsResponse represents live betting metrics response
type LiveMetricsResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Data    *LiveBettingMetrics `json:"data,omitempty"`
}

// LiveBettingMetrics represents live betting metrics
type LiveBettingMetrics struct {
	TotalMatches       int64         `json:"total_matches"`
	ActiveMatches      int64         `json:"active_matches"`
	SuspendedMatches   int64         `json:"suspended_matches"`
	OddsUpdateInterval time.Duration `json:"odds_update_interval"`
	LastActivity       time.Time     `json:"last_activity"`
}
