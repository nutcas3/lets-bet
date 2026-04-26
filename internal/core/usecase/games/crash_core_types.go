package games

import (
	"context"
	"sync"
	"sync/atomic"
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

	// Channel-based state management
	betChan     chan BetRequest
	cashoutChan chan CashoutRequest
	commandChan chan string // "START", "STOP", "CRASH"
	gameCancel  context.CancelFunc
	latestState atomic.Value // Stores latest GameState for lock-free reads
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
	AtomicCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) (bool, error)
	CreateBetWithWalletUpdate(ctx context.Context, bet *domain.GameBet, userID uuid.UUID, amount decimal.Decimal) (uuid.UUID, error)
	AtomicAutoCashoutWithCredit(ctx context.Context, id uuid.UUID, userID uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal, country string) (bool, error)
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
	Resp          chan error       `json:"-"` // Response channel for manager pattern
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
	BetID  uuid.UUID             `json:"bet_id"`
	UserID uuid.UUID             `json:"user_id"`
	Resp   chan *CashoutResponse `json:"-"` // Response channel for manager pattern
}

// CashoutResponse represents a cashout response
type CashoutResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message,omitempty"`
	CashoutOdds decimal.Decimal `json:"cashout_odds"`
	Payout      decimal.Decimal `json:"payout"`
	GameState   *GameState      `json:"game_state,omitempty"`
}
