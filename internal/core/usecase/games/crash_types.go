package games

import (
	"context"
	"sync"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase"
	"github.com/betting-platform/internal/core/usecase/tax"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CrashGameEngine manages the game loop for Aviator-style crash games
type CrashGameEngine struct {
	hub           WebSocketHub
	fairService   *usecase.ProvablyFairService
	gameRepo      GameRepository
	betRepo       GameBetRepository
	walletService *wallet.Service
	taxEngine     *tax.Engine
	mu            sync.RWMutex
	currentGame   *domain.Game
	roundNumber   int64
	tickInterval  time.Duration
}

// GameRepository interface for game operations
type GameRepository interface {
	Create(ctx context.Context, game *domain.Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.GameStatus) error
}

// GameBetRepository interface for game bet operations
type GameBetRepository interface {
	Create(ctx context.Context, bet *domain.GameBet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.GameBet, error)
	GetActiveByGame(ctx context.Context, gameID uuid.UUID) ([]*domain.GameBet, error)
	UpdateCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.GameBetStatus) error
}

// WebSocketHub interface for WebSocket operations
type WebSocketHub interface {
	BroadcastGameState(state any)
	GetActivePlayerCount(gameID uuid.UUID) int
}

// GameState represents the current state of a crash game
type GameState struct {
	GameID        uuid.UUID         `json:"game_id"`
	RoundNumber   int64             `json:"round_number"`
	Status        domain.GameStatus `json:"status"`
	CurrentOdds   decimal.Decimal   `json:"current_odds"`
	MaxOdds       decimal.Decimal   `json:"max_odds"`
	StartedAt     time.Time         `json:"started_at"`
	NextTickAt    time.Time         `json:"next_tick_at"`
	TimeRemaining time.Duration     `json:"time_remaining"`
	ActivePlayers int               `json:"active_players"`
	TotalBets     int64             `json:"total_bets"`
	TotalStake    decimal.Decimal   `json:"total_stake"`
	IsCrashed     bool              `json:"is_crashed"`
	CrashOdds     decimal.Decimal   `json:"crash_odds"`
}

// BetRequest represents a bet request
type BetRequest struct {
	GameID        uuid.UUID        `json:"game_id"`
	UserID        uuid.UUID        `json:"user_id"`
	Amount        decimal.Decimal  `json:"amount"`
	AutoCashoutAt *decimal.Decimal `json:"auto_cashout_at,omitempty"`
}

// BetResponse represents a bet response
type BetResponse struct {
	Success       bool             `json:"success"`
	Message       string           `json:"message,omitempty"`
	BetID         uuid.UUID        `json:"bet_id,omitempty"`
	GameState     *GameState       `json:"game_state,omitempty"`
	AutoCashoutAt *decimal.Decimal `json:"auto_cashout_at,omitempty"`
}

// CashoutRequest represents a cashout request
type CashoutRequest struct {
	BetID  uuid.UUID `json:"bet_id"`
	UserID uuid.UUID `json:"user_id"`
}

// CashoutResponse represents a cashout response
type CashoutResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message,omitempty"`
	CashoutOdds decimal.Decimal `json:"cashout_odds"`
	Payout      decimal.Decimal `json:"payout"`
	GameState   *GameState      `json:"game_state,omitempty"`
}

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

// GameAnalytics represents game analytics data
type GameAnalytics struct {
	TimeRange          time.Time       `json:"time_range"`
	GamesPlayed        int64           `json:"games_played"`
	UniquePlayers      int64           `json:"unique_players"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	TotalRevenue       decimal.Decimal `json:"total_revenue"`
	AverageSessionTime time.Duration   `json:"average_session_time"`
	PlayerRetention    decimal.Decimal `json:"player_retention"`
	PopularTimes       []TimeSlot      `json:"popular_times"`
	OddsDistribution   []OddsRange     `json:"odds_distribution"`
	BetPatterns        []BetPattern    `json:"bet_patterns"`
}

// TimeSlot represents a time slot with activity data
type TimeSlot struct {
	Hour    int             `json:"hour"`
	Games   int64           `json:"games"`
	Players int64           `json:"players"`
	Volume  decimal.Decimal `json:"volume"`
}

// OddsRange represents an odds range with statistics
type OddsRange struct {
	MinOdds   decimal.Decimal `json:"min_odds"`
	MaxOdds   decimal.Decimal `json:"max_odds"`
	GameCount int64           `json:"game_count"`
	Frequency decimal.Decimal `json:"frequency"`
}

// BetPattern represents betting patterns
type BetPattern struct {
	Pattern    string          `json:"pattern"`
	Count      int64           `json:"count"`
	Percentage decimal.Decimal `json:"percentage"`
	AvgAmount  decimal.Decimal `json:"avg_amount"`
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

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank        int             `json:"rank"`
	UserID      uuid.UUID       `json:"user_id"`
	Username    string          `json:"username"`
	TotalProfit decimal.Decimal `json:"total_profit"`
	WinRate     decimal.Decimal `json:"win_rate"`
	GamesPlayed int64           `json:"games_played"`
	LastPlayed  time.Time       `json:"last_played"`
}

// GameLeaderboard represents the game leaderboard
type GameLeaderboard struct {
	Period       string              `json:"period"` // "daily", "weekly", "monthly", "all_time"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Entries      []*LeaderboardEntry `json:"entries"`
	TotalPlayers int64               `json:"total_players"`
}
