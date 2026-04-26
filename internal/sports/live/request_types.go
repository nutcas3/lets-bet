package live

import (
	"time"

	"github.com/shopspring/decimal"
)

// LiveBetRequest represents a request to place a live bet
type LiveBetRequest struct {
	MatchID   string          `json:"match_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
	UserID    string          `json:"user_id"`
}

// LiveBetResponse represents the response to a live bet request
type LiveBetResponse struct {
	Success      bool            `json:"success"`
	BetID        string          `json:"bet_id,omitempty"`
	Message      string          `json:"message"`
	Balance      decimal.Decimal `json:"balance"`
	Odds         decimal.Decimal `json:"odds"`
	Amount       decimal.Decimal `json:"amount"`
	PotentialWin decimal.Decimal `json:"potential_win"`
}

// OddsUpdateRequest represents a request to update odds
type OddsUpdateRequest struct {
	MatchID   string             `json:"match_id"`
	MarketID  string             `json:"market_id"`
	OutcomeID string             `json:"outcome_id"`
	NewOdds   decimal.Decimal    `json:"new_odds"`
	Markets   []MarketOddsUpdate `json:"markets"`
	MatchOdds *MatchOddsUpdate   `json:"match_odds,omitempty"`
}

// MarketOddsUpdate represents odds update for a market
type MarketOddsUpdate struct {
	MarketID string              `json:"market_id"`
	Outcomes []OutcomeOddsUpdate `json:"outcomes"`
}

// OutcomeOddsUpdate represents odds update for an outcome
type OutcomeOddsUpdate struct {
	OutcomeID string          `json:"outcome_id"`
	Odds      decimal.Decimal `json:"odds"`
}

// MatchOddsUpdate represents match-level odds update
type MatchOddsUpdate struct {
	HomeWin decimal.Decimal `json:"home_win"`
	Draw    decimal.Decimal `json:"draw"`
	AwayWin decimal.Decimal `json:"away_win"`
}

// LiveBettingMetrics represents live betting metrics
type LiveBettingMetrics struct {
	ActiveMatches      int64           `json:"active_matches"`
	TotalBets          int64           `json:"total_bets"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	AverageBetSize     decimal.Decimal `json:"average_bet_size"`
	LastUpdated        time.Time       `json:"last_updated"`
	TotalMatches       int64           `json:"total_matches"`
	SuspendedMatches   int64           `json:"suspended_matches"`
	OddsUpdateInterval time.Duration   `json:"odds_update_interval"`
	LastActivity       time.Time       `json:"last_activity"`
}
