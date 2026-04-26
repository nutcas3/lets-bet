package virtualsports

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/shopspring/decimal"
)

// VirtualSportsService manages virtual sports games and betting
type VirtualSportsService struct {
	matchRepo       postgres.MatchRepository
	marketRepo      postgres.BettingMarketRepository
	outcomeRepo     postgres.MarketOutcomeRepository
	betRepo         postgres.SportBetRepository
	walletService   WalletService
	eventBus        EventBus
	rng             *rand.Rand
	gameIDGenerator *id.SnowflakeGenerator
	mu              sync.RWMutex

	// Virtual game state
	games     map[string]*VirtualGame
	schedules map[string]*GameSchedule
}

// WalletService interface for wallet operations
type WalletService interface {
	Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
	Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
}

// EventBus interface for event publishing
type EventBus interface {
	Publish(topic string, event any) error
}

// Movement represents a wallet movement
type Movement struct {
	UserID        string                 `json:"user_id"`
	Amount        decimal.Decimal        `json:"amount"`
	Type          domain.TransactionType `json:"type"`
	ReferenceID   *string                `json:"reference_id,omitempty"`
	ReferenceType string                 `json:"reference_type"`
	Description   string                 `json:"description"`
	ProviderName  string                 `json:"provider_name"`
	ProviderTxnID string                 `json:"provider_txn_id"`
	CountryCode   string                 `json:"country_code"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID string `json:"id"`
}

// VirtualGameEvent represents an event in a virtual game
type VirtualGameEvent struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "goal", "card", "substitution", "whistle"
	Minute    int       `json:"minute"`
	Team      string    `json:"team"` // "home", "away"
	Player    string    `json:"player,omitempty"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// VirtualGameResult represents the final result of a virtual game
type VirtualGameResult struct {
	Winner        string          `json:"winner"` // "home", "away", "draw"
	HomeScore     int             `json:"home_score"`
	AwayScore     int             `json:"away_score"`
	Scorers       []GoalScorer    `json:"scorers"`
	Cards         []Card          `json:"cards"`
	Substitutions []Substitution  `json:"substitutions"`
	Statistics    *GameStatistics `json:"statistics"`
}

// GoalScorer represents a goal scorer
type GoalScorer struct {
	Player string `json:"player"`
	Team   string `json:"team"`
	Minute int    `json:"minute"`
	Type   string `json:"type"` // "regular", "penalty", "own_goal"
}

// Card represents a card in a virtual game
type Card struct {
	Player string `json:"player"`
	Team   string `json:"team"`
	Minute int    `json:"minute"`
	Type   string `json:"type"` // "yellow", "red"
}

// Substitution represents a substitution in a virtual game
type Substitution struct {
	PlayerOut string `json:"player_out"`
	PlayerIn  string `json:"player_in"`
	Team      string `json:"team"`
	Minute    int    `json:"minute"`
}

// GameStatistics represents game statistics
type GameStatistics struct {
	Possession  *TeamPossession  `json:"possession"`
	Shots       *TeamShots       `json:"shots"`
	Corners     *TeamCorners     `json:"corners"`
	Fouls       *TeamFouls       `json:"fouls"`
	Offsides    *TeamOffsides    `json:"offsides"`
	BallControl *TeamBallControl `json:"ball_control"`
}

// TeamPossession represents possession statistics
type TeamPossession struct {
	Home float64 `json:"home"`
	Away float64 `json:"away"`
}

// TeamShots represents shot statistics
type TeamShots struct {
	Home *ShotStats `json:"home"`
	Away *ShotStats `json:"away"`
}

// ShotStats represents shot statistics for a team
type ShotStats struct {
	Total     int `json:"total"`
	OnTarget  int `json:"on_target"`
	OffTarget int `json:"off_target"`
	Blocked   int `json:"blocked"`
}

// TeamCorners represents corner statistics
type TeamCorners struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// TeamFouls represents foul statistics
type TeamFouls struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// TeamOffsides represents offside statistics
type TeamOffsides struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// TeamBallControl represents ball control statistics
type TeamBallControl struct {
	Home float64 `json:"home"`
	Away float64 `json:"away"`
}

// GameSchedule represents a game schedule
type GameSchedule struct {
	GameID    string    `json:"game_id"`
	StartTime time.Time `json:"start_time"`
	Duration  int       `json:"duration"` // in minutes
	Interval  int       `json:"interval"` // time between games in minutes
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VirtualSport represents a virtual sport type
type VirtualSport struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	GameType    string        `json:"game_type"` // "football", "horse_racing", "dog_racing"
	Duration    time.Duration `json:"duration"`
	IsActive    bool          `json:"is_active"`
}

// BetRequest represents a betting request for virtual sports
type BetRequest struct {
	GameID    string          `json:"game_id"`
	UserID    string          `json:"user_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
}

// BetResponse represents a betting response
type BetResponse struct {
	BetID        string          `json:"bet_id"`
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
	Amount       decimal.Decimal `json:"amount"`
	Odds         decimal.Decimal `json:"odds"`
	PotentialWin decimal.Decimal `json:"potential_win"`
}

// VirtualSportsMetrics represents service metrics
type VirtualSportsMetrics struct {
	TotalGames        int64           `json:"total_games"`
	ActiveGames       int64           `json:"active_games"`
	CompletedGames    int64           `json:"completed_games"`
	TotalBets         int64           `json:"total_bets"`
	TotalVolume       decimal.Decimal `json:"total_volume"`
	ActiveUsers       int64           `json:"active_users"`
	AverageGameLength time.Duration   `json:"average_game_length"`
	LastActivity      time.Time       `json:"last_activity"`
}
