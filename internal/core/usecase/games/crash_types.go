package games

import (
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GameHistory represents game history
type GameHistory struct {
	GameID      uuid.UUID       `json:"game_id"`
	RoundNumber int64           `json:"round_number"`
	StartedAt   time.Time       `json:"started_at"`
	CrashedAt   time.Time       `json:"crashed_at"`
	CrashOdds   decimal.Decimal `json:"crash_odds"`
	MaxOdds     decimal.Decimal `json:"max_odds"`
	TotalBets   int64           `json:"total_bets"`
	TotalStake  decimal.Decimal `json:"total_stake"`
	TotalPayout decimal.Decimal `json:"total_payout"`
	Profit      decimal.Decimal `json:"profit"`
}

// PlayerStats represents player statistics
type PlayerStats struct {
	UserID           uuid.UUID       `json:"user_id"`
	TotalGames       int64           `json:"total_games"`
	TotalBets        int64           `json:"total_bets"`
	TotalStake       decimal.Decimal `json:"total_stake"`
	TotalPayout      decimal.Decimal `json:"total_payout"`
	TotalProfit      decimal.Decimal `json:"total_profit"`
	WinRate          decimal.Decimal `json:"win_rate"`
	AverageBetSize   decimal.Decimal `json:"average_bet_size"`
	BiggestWin       decimal.Decimal `json:"biggest_win"`
	BiggestLoss      decimal.Decimal `json:"biggest_loss"`
	CurrentStreak    int             `json:"current_streak"`
	LongestWinStreak int             `json:"longest_win_streak"`
	LastPlayedAt     time.Time       `json:"last_played_at"`
}

// GameMetrics represents game metrics
type GameMetrics struct {
	TotalGames         int64           `json:"total_games"`
	ActiveGames        int64           `json:"active_games"`
	TotalPlayers       int64           `json:"total_players"`
	TotalBets          int64           `json:"total_bets"`
	TotalStake         decimal.Decimal `json:"total_stake"`
	TotalPayout        decimal.Decimal `json:"total_payout"`
	TotalProfit        decimal.Decimal `json:"total_profit"`
	AverageOdds        decimal.Decimal `json:"average_odds"`
	HighestOdds        decimal.Decimal `json:"highest_odds"`
	LowestOdds         decimal.Decimal `json:"lowest_odds"`
	AverageBetsPerGame decimal.Decimal `json:"average_bets_per_game"`
	AverageStakePerBet decimal.Decimal `json:"average_stake_per_bet"`
	ProfitMargin       decimal.Decimal `json:"profit_margin"`
	LastGameTime       time.Time       `json:"last_game_time"`
	NextGameTime       time.Time       `json:"next_game_time"`
}

// GameConfig represents game configuration
type GameConfig struct {
	MinBetAmount       decimal.Decimal `json:"min_bet_amount"`
	MaxBetAmount       decimal.Decimal `json:"max_bet_amount"`
	MaxMultiplier      decimal.Decimal `json:"max_multiplier"`
	MinMultiplier      decimal.Decimal `json:"min_multiplier"`
	HouseEdge          decimal.Decimal `json:"house_edge"`
	TickInterval       time.Duration   `json:"tick_interval"`
	MaxGameDuration    time.Duration   `json:"max_game_duration"`
	BetTimeout         time.Duration   `json:"bet_timeout"`
	CashoutDelay       time.Duration   `json:"cashout_delay"`
	EnableAutoCashout  bool            `json:"enable_auto_cashout"`
	MinAutoCashoutOdds decimal.Decimal `json:"min_auto_cashout_odds"`
	EnableStatistics   bool            `json:"enable_statistics"`
	EnableHistory      bool            `json:"enable_history"`
}

// GameEvent represents a game event
type GameEvent struct {
	Type        string    `json:"type"`
	GameID      uuid.UUID `json:"game_id"`
	RoundNumber int64     `json:"round_number"`
	Timestamp   time.Time `json:"timestamp"`
	Data        any       `json:"data"`
}

// BetEvent represents a bet event
type BetEvent struct {
	BetID     uuid.UUID            `json:"bet_id"`
	GameID    uuid.UUID            `json:"game_id"`
	UserID    uuid.UUID            `json:"user_id"`
	Amount    decimal.Decimal      `json:"amount"`
	Odds      decimal.Decimal      `json:"odds"`
	Status    domain.GameBetStatus `json:"status"`
	Timestamp time.Time            `json:"timestamp"`
}

// CashoutEvent represents a cashout event
type CashoutEvent struct {
	BetID       uuid.UUID       `json:"bet_id"`
	GameID      uuid.UUID       `json:"game_id"`
	UserID      uuid.UUID       `json:"user_id"`
	CashoutOdds decimal.Decimal `json:"cashout_odds"`
	Payout      decimal.Decimal `json:"payout"`
	Profit      decimal.Decimal `json:"profit"`
	Timestamp   time.Time       `json:"timestamp"`
}

// GameResult represents the result of a completed game
type GameResult struct {
	GameID         uuid.UUID         `json:"game_id"`
	RoundNumber    int64             `json:"round_number"`
	Status         domain.GameStatus `json:"status"`
	CrashOdds      decimal.Decimal   `json:"crash_odds"`
	MaxOdds        decimal.Decimal   `json:"max_odds"`
	StartedAt      time.Time         `json:"started_at"`
	CrashedAt      time.Time         `json:"crashed_at"`
	Duration       time.Duration     `json:"duration"`
	TotalBets      int64             `json:"total_bets"`
	TotalStake     decimal.Decimal   `json:"total_stake"`
	TotalPayout    decimal.Decimal   `json:"total_payout"`
	Profit         decimal.Decimal   `json:"profit"`
	HouseProfit    decimal.Decimal   `json:"house_profit"`
	WinningBets    int64             `json:"winning_bets"`
	LosingBets     int64             `json:"losing_bets"`
	AutoCashouts   int64             `json:"auto_cashouts"`
	ManualCashouts int64             `json:"manual_cashouts"`
}

// PlayerSession represents a player's game session
type PlayerSession struct {
	SessionID   uuid.UUID       `json:"session_id"`
	UserID      uuid.UUID       `json:"user_id"`
	GameID      uuid.UUID       `json:"game_id"`
	StartedAt   time.Time       `json:"started_at"`
	EndedAt     *time.Time      `json:"ended_at,omitempty"`
	TotalBets   int64           `json:"total_bets"`
	TotalStake  decimal.Decimal `json:"total_stake"`
	TotalPayout decimal.Decimal `json:"total_payout"`
	Profit      decimal.Decimal `json:"profit"`
	IsActive    bool            `json:"is_active"`
}

// FairnessVerification represents provably fair verification
type FairnessVerification struct {
	GameID     uuid.UUID `json:"game_id"`
	Seed       string    `json:"seed"`
	Hash       string    `json:"hash"`
	ServerSeed string    `json:"server_seed"`
	ClientSeed string    `json:"client_seed"`
	CrashPoint int       `json:"crash_point"`
	IsVerified bool      `json:"is_verified"`
	VerifiedAt time.Time `json:"verified_at"`
}

// GameSettings represents adjustable game settings
type GameSettings struct {
	Enabled            bool            `json:"enabled"`
	MinBetAmount       decimal.Decimal `json:"min_bet_amount"`
	MaxBetAmount       decimal.Decimal `json:"max_bet_amount"`
	MaxMultiplier      decimal.Decimal `json:"max_multiplier"`
	MinMultiplier      decimal.Decimal `json:"min_multiplier"`
	HouseEdge          decimal.Decimal `json:"house_edge"`
	TickInterval       time.Duration   `json:"tick_interval"`
	MaxGameDuration    time.Duration   `json:"max_game_duration"`
	BetTimeout         time.Duration   `json:"bet_timeout"`
	CashoutDelay       time.Duration   `json:"cashout_delay"`
	EnableAutoCashout  bool            `json:"enable_auto_cashout"`
	MinAutoCashoutOdds decimal.Decimal `json:"min_auto_cashout_odds"`
	EnableStatistics   bool            `json:"enable_statistics"`
	EnableHistory      bool            `json:"enable_history"`
	EnableFairness     bool            `json:"enable_fairness"`
	FairnessAlgorithm  string          `json:"fairness_algorithm"`
	MaxPlayersPerGame  int             `json:"max_players_per_game"`
	EnableChat         bool            `json:"enable_chat"`
	EnableLeaderboard  bool            `json:"enable_leaderboard"`
}
