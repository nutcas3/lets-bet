package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/sports/live"
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

// Helper functions
func generateID() string {
	return uuid.New().String()
}

func getUserID(_ context.Context) string {
	// In a real implementation, this would extract user ID from JWT token
	// For now, return a dummy ID
	return "user-" + uuid.New().String()
}

func WriteError(w http.ResponseWriter, err error, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"error":   message,
		"details": err.Error(),
	})
}

func WriteJSON(w http.ResponseWriter, data any, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// RegisterRoutes registers live betting routes
func (h *LiveHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/live/matches", h.GetLiveMatches)
	mux.HandleFunc("/api/live/matches/", h.GetLiveMatch)
}

// GetLiveMatches returns all live matches
func (h *LiveHandler) GetLiveMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	matches, err := h.liveService.GetLiveMatches(ctx)
	if err != nil {
		WriteError(w, err, "Failed to get live matches", http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"matches": matches,
		"count":   len(matches),
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetLiveMatch returns a specific live match
func (h *LiveHandler) GetLiveMatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
			break
		}
	}

	if matchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, matchID)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	WriteJSON(w, match, http.StatusOK)
}

// PlaceLiveBet places a live bet
func (h *LiveHandler) PlaceLiveBet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
			break
		}
	}

	if matchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	var req PlaceLiveBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validatePlaceLiveBetRequest(&req); err != nil {
		WriteError(w, err, "Validation failed", http.StatusBadRequest)
		return
	}

	// Create bet
	bet := &domain.SportBet{
		ID:        generateID(),
		UserID:    getUserID(ctx),
		EventID:   matchID,
		MarketID:  req.MarketID,
		OutcomeID: req.OutcomeID,
		Amount:    req.Amount,
		Odds:      req.Odds,
		Status:    domain.BetStatusPending,
		PlacedAt:  time.Now(),
	}

	err := h.liveService.PlaceLiveBet(ctx, bet)
	if err != nil {
		WriteError(w, err, "Failed to place live bet", http.StatusBadRequest)
		return
	}

	response := map[string]any{
		"bet_id":  bet.ID,
		"status":  "success",
		"message": "Live bet placed successfully",
	}

	WriteJSON(w, response, http.StatusCreated)
}

// GetLiveMarkets returns live markets for a match
func (h *LiveHandler) GetLiveMarkets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
			break
		}
	}

	if matchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, matchID)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	response := map[string]any{
		"markets": match.LiveMarkets,
		"count":   len(match.LiveMarkets),
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetLiveOutcomes returns live outcomes for a market
func (h *LiveHandler) GetLiveOutcomes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID and market ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID, marketID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
		}
		if part == "markets" && i+1 < len(parts) {
			marketID = parts[i+1]
		}
	}

	if matchID == "" || marketID == "" {
		WriteError(w, nil, "Match ID and Market ID are required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, matchID)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	// Find the market
	for _, market := range match.LiveMarkets {
		if market.Market.ID == marketID {
			response := map[string]any{
				"market_id": market.Market.ID,
				"outcomes":  market.LiveOdds,
				"count":     len(market.LiveOdds),
			}

			WriteJSON(w, response, http.StatusOK)
			return
		}
	}

	WriteError(w, nil, "Market not found", http.StatusNotFound)
}

// GetLiveOdds returns live odds for a match
func (h *LiveHandler) GetLiveOdds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
			break
		}
	}

	if matchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, matchID)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	odds := make(map[string]any)
	for _, market := range match.LiveMarkets {
		marketOdds := make(map[string]any)
		for _, outcome := range market.LiveOdds {
			marketOdds[outcome.Outcome.ID] = map[string]any{
				"odds":              outcome.Outcome.Odds,
				"price":             outcome.Outcome.Price,
				"status":            outcome.Outcome.Status,
				"current_odds":      outcome.CurrentOdds,
				"previous_odds":     outcome.PreviousOdds,
				"odds_change_time":  outcome.OddsChangeTime,
				"odds_change_count": outcome.OddsChangeCount,
				"total_volume":      outcome.TotalVolume,
				"live_volume":       outcome.LiveVolume,
			}
		}
		odds[market.Market.ID] = map[string]any{
			"id":                market.Market.ID,
			"type":              market.Market.Type,
			"name":              market.Market.Name,
			"status":            market.Market.Status,
			"is_live":           market.IsLive,
			"odds":              marketOdds,
			"is_suspended":      market.IsSuspended,
			"suspension_reason": market.SuspensionReason,
			"last_update":       market.LastOddsUpdate,
			"update_count":      market.OddsUpdateCount,
			"outcomes":          marketOdds,
		}
	}

	response := map[string]any{
		"match_id":   matchID,
		"odds":       odds,
		"updated_at": match.LastUpdated,
	}

	WriteJSON(w, response, http.StatusOK)
}

// GetLiveStats returns live statistics for a match
func (h *LiveHandler) GetLiveStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var matchID string
	for i, part := range parts {
		if part == "matches" && i+1 < len(parts) {
			matchID = parts[i+1]
			break
		}
	}

	if matchID == "" {
		WriteError(w, nil, "Match ID is required", http.StatusBadRequest)
		return
	}

	match, err := h.liveService.GetLiveMatch(ctx, matchID)
	if err != nil {
		WriteError(w, err, "Failed to get live match", http.StatusNotFound)
		return
	}

	stats := map[string]any{
		"match_id":          matchID,
		"current_minute":    match.CurrentMinute,
		"home_score":        match.HomeScore,
		"away_score":        match.AwayScore,
		"home_possession":   match.HomePossession,
		"away_possession":   match.AwayPossession,
		"home_corners":      match.HomeCorners,
		"away_corners":      match.AwayCorners,
		"home_yellow_cards": match.HomeYellowCards,
		"away_yellow_cards": match.AwayYellowCards,
		"home_red_cards":    match.HomeRedCards,
		"away_red_cards":    match.AwayRedCards,
		"is_suspended":      match.IsSuspended,
		"suspension_reason": match.SuspensionReason,
		"last_updated":      match.LastUpdated,
	}

	// Add betting stats
	totalVolume := decimal.Zero
	totalBets := 0
	for _, market := range match.LiveMarkets {
		for _, outcome := range market.LiveOdds {
			totalVolume = totalVolume.Add(outcome.LiveVolume)
			totalBets++
		}
	}

	stats["betting"] = map[string]any{
		"total_volume":   totalVolume,
		"total_bets":     totalBets,
		"active_markets": len(match.LiveMarkets),
	}

	WriteJSON(w, stats, http.StatusOK)
}

// PlaceLiveBetRequest represents a place live bet request
type PlaceLiveBetRequest struct {
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
}

// validatePlaceLiveBetRequest validates a place live bet request
func (h *LiveHandler) validatePlaceLiveBetRequest(req *PlaceLiveBetRequest) error {
	if req.MarketID == "" {
		return fmt.Errorf("market_id is required")
	}

	if req.OutcomeID == "" {
		return fmt.Errorf("outcome_id is required")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be greater than 0")
	}

	if req.Amount.GreaterThan(decimal.NewFromInt(100000)) {
		return fmt.Errorf("amount exceeds maximum limit")
	}

	if req.Odds.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("odds must be greater than 0")
	}

	return nil
}

// LiveMatchResponse represents a live match response
type LiveMatchResponse struct {
	ID               string                `json:"id"`
	Sport            string                `json:"sport"`
	Tournament       string                `json:"tournament"`
	HomeTeam         string                `json:"home_team"`
	AwayTeam         string                `json:"away_team"`
	Status           domain.MatchStatus    `json:"status"`
	StartTime        time.Time             `json:"start_time"`
	Score            *domain.MatchScore    `json:"score,omitempty"`
	CurrentMinute    int                   `json:"current_minute"`
	HomeScore        int                   `json:"home_score"`
	AwayScore        int                   `json:"away_score"`
	HomePossession   float64               `json:"home_possession"`
	AwayPossession   float64               `json:"away_possession"`
	HomeCorners      int                   `json:"home_corners"`
	AwayCorners      int                   `json:"away_corners"`
	HomeYellowCards  int                   `json:"home_yellow_cards"`
	AwayYellowCards  int                   `json:"away_yellow_cards"`
	HomeRedCards     int                   `json:"home_red_cards"`
	AwayRedCards     int                   `json:"away_red_cards"`
	LiveMarkets      []*LiveMarketResponse `json:"live_markets"`
	IsSuspended      bool                  `json:"is_suspended"`
	SuspensionReason string                `json:"suspension_reason,omitempty"`
	LastUpdated      time.Time             `json:"last_updated"`
}

// LiveMarketResponse represents a live market response
type LiveMarketResponse struct {
	ID               string                 `json:"id"`
	MatchID          string                 `json:"match_id"`
	Type             domain.MarketType      `json:"type"`
	Name             string                 `json:"name"`
	Status           domain.MarketStatus    `json:"status"`
	IsLive           bool                   `json:"is_live"`
	LiveOdds         []*LiveOutcomeResponse `json:"live_odds"`
	LastOddsUpdate   time.Time              `json:"last_odds_update"`
	OddsUpdateCount  int                    `json:"odds_update_count"`
	IsSuspended      bool                   `json:"is_suspended"`
	SuspensionReason string                 `json:"suspension_reason,omitempty"`
}

// LiveOutcomeResponse represents a live outcome response
type LiveOutcomeResponse struct {
	ID              string               `json:"id"`
	MarketID        string               `json:"market_id"`
	Name            string               `json:"name"`
	Odds            decimal.Decimal      `json:"odds"`
	Price           decimal.Decimal      `json:"price"`
	Status          domain.OutcomeStatus `json:"status"`
	CurrentOdds     decimal.Decimal      `json:"current_odds"`
	PreviousOdds    decimal.Decimal      `json:"previous_odds"`
	OddsChangeTime  time.Time            `json:"odds_change_time"`
	OddsChangeCount int                  `json:"odds_change_count"`
	TotalVolume     decimal.Decimal      `json:"total_volume"`
	LiveVolume      decimal.Decimal      `json:"live_volume"`
}

// LiveMatchesResponse represents live matches response
type LiveMatchesResponse struct {
	Matches []*LiveMatchResponse `json:"matches"`
	Count   int                  `json:"count"`
}

// LiveOddsResponse represents live odds response
type LiveOddsResponse struct {
	MatchID   string         `json:"match_id"`
	Odds      map[string]any `json:"odds"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// LiveStatsResponse represents live stats response
type LiveStatsResponse struct {
	MatchID          string         `json:"match_id"`
	CurrentMinute    int            `json:"current_minute"`
	HomeScore        int            `json:"home_score"`
	AwayScore        int            `json:"away_score"`
	HomePossession   float64        `json:"home_possession"`
	AwayPossession   float64        `json:"away_possession"`
	HomeCorners      int            `json:"home_corners"`
	AwayCorners      int            `json:"away_corners"`
	HomeYellowCards  int            `json:"home_yellow_cards"`
	AwayYellowCards  int            `json:"away_yellow_cards"`
	HomeRedCards     int            `json:"home_red_cards"`
	AwayRedCards     int            `json:"away_red_cards"`
	IsSuspended      bool           `json:"is_suspended"`
	SuspensionReason string         `json:"suspension_reason,omitempty"`
	LastUpdated      time.Time      `json:"last_updated"`
	Betting          map[string]any `json:"betting"`
}

// PlaceLiveBetResponse represents place live bet response
type PlaceLiveBetResponse struct {
	BetID   string `json:"bet_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// GetLiveMarketsResponse represents get live markets response
type GetLiveMarketsResponse struct {
	Markets []*LiveMarketResponse `json:"markets"`
	Count   int                   `json:"count"`
}

// GetLiveOutcomesResponse represents get live outcomes response
type GetLiveOutcomesResponse struct {
	Outcomes []*LiveOutcomeResponse `json:"outcomes"`
	Count    int                    `json:"count"`
}
