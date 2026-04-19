package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Match represents a sports event/match
type Match struct {
	ID          string      `json:"id" db:"id"`
	Sport       Sport       `json:"sport" db:"sport"`
	League      string      `json:"league" db:"league"`
	HomeTeam    string      `json:"home_team" db:"home_team"`
	AwayTeam    string      `json:"away_team" db:"away_team"`
	StartTime   time.Time   `json:"start_time" db:"start_time"`
	Status      MatchStatus `json:"status" db:"status"`
	Score       *MatchScore `json:"score,omitempty" db:"score"`
	Markets     []Market    `json:"markets" db:"-"`
	CountryCode string      `json:"country_code" db:"country_code"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type Sport string

const (
	SportFootball   Sport = "FOOTBALL"
	SportBasketball Sport = "BASKETBALL"
	SportTennis     Sport = "TENNIS"
	SportCricket    Sport = "CRICKET"
	SportRugby      Sport = "RUGBY"
)

type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "SCHEDULED"
	MatchStatusLive      MatchStatus = "LIVE"
	MatchStatusFinished  MatchStatus = "FINISHED"
	MatchStatusPostponed MatchStatus = "POSTPONED"
	MatchStatusCancelled MatchStatus = "CANCELLED"
)

type MatchScore struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`

	// Additional scores for different sports
	HomeScoreHT *int `json:"home_score_ht,omitempty"` // Half-time
	AwayScoreHT *int `json:"away_score_ht,omitempty"`
	HomeScoreFT *int `json:"home_score_ft,omitempty"` // Full-time
	AwayScoreFT *int `json:"away_score_ft,omitempty"`

	// Tennis sets
	HomeSets *int `json:"home_sets,omitempty"`
	AwaySets *int `json:"away_sets,omitempty"`
}

// Market represents a betting market for a match
type Market struct {
	ID          string       `json:"id" db:"id"`
	MatchID     string       `json:"match_id" db:"match_id"`
	Type        MarketType   `json:"type" db:"type"`
	Name        string       `json:"name" db:"name"`
	Outcomes    []Outcome    `json:"outcomes" db:"-"`
	Status      MarketStatus `json:"status" db:"status"`
	SuspendedAt *time.Time   `json:"suspended_at,omitempty" db:"suspended_at"`
}

type MarketType string

const (
	MarketTypeMatchWinner      MarketType = "MATCH_WINNER"
	MarketTypeHandicap         MarketType = "HANDICAP"
	MarketTypeTotalGoals       MarketType = "TOTAL_GOALS"
	MarketTypeBothTeamsScore   MarketType = "BOTH_TEAMS_SCORE"
	MarketTypeCorrectScore     MarketType = "CORRECT_SCORE"
	MarketTypeFirstGoalScorer  MarketType = "FIRST_GOAL_SCORER"
	MarketTypeHalfTimeFullTime MarketType = "HALF_TIME_FULL_TIME"
)

type MarketStatus string

const (
	MarketStatusOpen      MarketStatus = "OPEN"
	MarketStatusSuspended MarketStatus = "SUSPENDED"
	MarketStatusClosed    MarketStatus = "CLOSED"
	MarketStatusSettled   MarketStatus = "SETTLED"
)

// Outcome represents a specific outcome in a market
type Outcome struct {
	ID       string          `json:"id" db:"id"`
	MarketID string          `json:"market_id" db:"market_id"`
	Name     string          `json:"name" db:"name"`
	Odds     decimal.Decimal `json:"odds" db:"odds"`
	Price    decimal.Decimal `json:"price" db:"price"` // Alternative odds format
	Status   OutcomeStatus   `json:"status" db:"status"`
}

type OutcomeStatus string

const (
	OutcomeStatusPending OutcomeStatus = "PENDING"
	OutcomeStatusWon     OutcomeStatus = "WON"
	OutcomeStatusLost    OutcomeStatus = "LOST"
	OutcomeStatusVoid    OutcomeStatus = "VOID"
)

// IsLive checks if the match is currently in progress
func (m *Match) IsLive() bool {
	return m.Status == MatchStatusLive
}

// IsFinished checks if the match has ended
func (m *Match) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

// GetMarketByType finds a market by its type
func (m *Match) GetMarketByType(marketType MarketType) *Market {
	for _, market := range m.Markets {
		if market.Type == marketType {
			return &market
		}
	}
	return nil
}

// GetOutcomeByName finds an outcome by its name in a specific market
func (m *Match) GetOutcomeByName(marketType MarketType, outcomeName string) *Outcome {
	market := m.GetMarketByType(marketType)
	if market == nil {
		return nil
	}

	for _, outcome := range market.Outcomes {
		if outcome.Name == outcomeName {
			return &outcome
		}
	}
	return nil
}
