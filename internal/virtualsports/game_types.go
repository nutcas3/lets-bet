package virtualsports

import (
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// VirtualGame represents a virtual sports game
type VirtualGame struct {
	ID          string           `json:"id"`
	Sport       domain.Sport     `json:"sport"`
	HomeTeam    *VirtualTeam     `json:"home_team"`
	AwayTeam    *VirtualTeam     `json:"away_team"`
	Status      GameStatus       `json:"status"`
	Score       *VirtualScore    `json:"score"`
	Odds        []*VirtualOdds   `json:"odds"`
	Outcomes    []*VirtualOutcome `json:"outcomes"`
	Events      []*VirtualGameEvent `json:"events"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     *time.Time       `json:"end_time"`
	Statistics  *GameStatistics  `json:"statistics"`
}

// VirtualTeam represents a virtual sports team
type VirtualTeam struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	ShortName  string         `json:"short_name"`
	Strength   decimal.Decimal `json:"strength"`
	Form       []string       `json:"form"`
	Statistics *TeamStatistics `json:"statistics"`
}

// TeamStatistics represents team performance statistics
type TeamStatistics struct {
	MatchesPlayed int           `json:"matches_played"`
	Wins          int           `json:"wins"`
	Draws         int           `json:"draws"`
	Losses        int           `json:"losses"`
	GoalsScored   int           `json:"goals_scored"`
	GoalsConceded int           `json:"goals_conceded"`
	WinRate       decimal.Decimal `json:"win_rate"`
	AverageGoals  decimal.Decimal `json:"average_goals"`
}

// VirtualScore represents the score in a virtual game
type VirtualScore struct {
	HomeScore int `json:"home_score"`
	AwayScore int `json:"away_score"`
	Period    int `json:"period"`
}

// VirtualOdds represents betting odds for a virtual game
type VirtualOdds struct {
	MarketID  string          `json:"market_id"`
	MarketName string         `json:"market_name"`
	HomeOdds  decimal.Decimal `json:"home_odds"`
	DrawOdds  decimal.Decimal `json:"draw_odds"`
	AwayOdds  decimal.Decimal `json:"away_odds"`
}

// VirtualOutcome represents a betting outcome
type VirtualOutcome struct {
	ID      string          `json:"id"`
	MarketID string         `json:"market_id"`
	Name    string          `json:"name"`
	Odds    decimal.Decimal `json:"odds"`
	Status  OutcomeStatus   `json:"status"`
}

// GameStatus represents the status of a virtual game
type GameStatus string

const (
	GameStatusScheduled GameStatus = "SCHEDULED"
	GameStatusInProgress GameStatus = "IN_PROGRESS"
	GameStatusCompleted GameStatus = "COMPLETED"
	GameStatusCancelled GameStatus = "CANCELLED"
	GameStatusPostponed GameStatus = "POSTPONED"
)

// OutcomeStatus represents the status of a betting outcome
type OutcomeStatus string

const (
	OutcomeStatusPending OutcomeStatus = "PENDING"
	OutcomeStatusWon OutcomeStatus = "WON"
	OutcomeStatusLost OutcomeStatus = "LOST"
	OutcomeStatusVoid OutcomeStatus = "VOID"
)
