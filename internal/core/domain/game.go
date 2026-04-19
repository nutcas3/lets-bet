package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Game represents a crash game instance (Aviator-style)
type Game struct {
	ID          uuid.UUID `json:"id" db:"id"`
	GameType    GameType  `json:"game_type" db:"game_type"`
	RoundNumber int64     `json:"round_number" db:"round_number"`

	// Provably Fair seeds
	ServerSeed     string `json:"-" db:"server_seed"`                     // Hidden until game ends
	ServerSeedHash string `json:"server_seed_hash" db:"server_seed_hash"` // Public
	ClientSeed     string `json:"client_seed" db:"client_seed"`

	// Game result
	CrashPoint decimal.Decimal `json:"crash_point" db:"crash_point"` // e.g., 2.45x

	// State
	Status    GameStatus `json:"status" db:"status"`
	StartedAt time.Time  `json:"started_at" db:"started_at"`
	CrashedAt *time.Time `json:"crashed_at,omitempty" db:"crashed_at"`

	// Metadata
	CountryCode   string          `json:"country_code" db:"country_code"`
	MinBet        decimal.Decimal `json:"min_bet" db:"min_bet"`
	MaxBet        decimal.Decimal `json:"max_bet" db:"max_bet"`
	MaxMultiplier decimal.Decimal `json:"max_multiplier" db:"max_multiplier"`
}

type GameType string

const (
	GameTypeCrash   GameType = "CRASH"   // Aviator-style
	GameTypeVirtual GameType = "VIRTUAL" // Virtual sports
	GameTypeSlot    GameType = "SLOT"
)

type GameStatus string

const (
	GameStatusWaiting GameStatus = "WAITING" // Betting phase
	GameStatusRunning GameStatus = "RUNNING" // Flight in progress
	GameStatusCrashed GameStatus = "CRASHED" // Round ended
)

// GameBet represents a player's bet in a crash game
type GameBet struct {
	ID     uuid.UUID `json:"id" db:"id"`
	GameID uuid.UUID `json:"game_id" db:"game_id"`
	UserID uuid.UUID `json:"user_id" db:"user_id"`

	Amount   decimal.Decimal `json:"amount" db:"amount"`
	Currency string          `json:"currency" db:"currency"`

	// Cashout
	CashedOut bool             `json:"cashed_out" db:"cashed_out"`
	CashoutAt *decimal.Decimal `json:"cashout_at,omitempty" db:"cashout_at"` // Multiplier when cashed out
	Payout    decimal.Decimal  `json:"payout" db:"payout"`

	Status      GameBetStatus `json:"status" db:"status"`
	PlacedAt    time.Time     `json:"placed_at" db:"placed_at"`
	CashedOutAt *time.Time    `json:"cashed_out_at,omitempty" db:"cashed_out_at"`
}

type GameBetStatus string

const (
	GameBetStatusActive    GameBetStatus = "ACTIVE"
	GameBetStatusWon       GameBetStatus = "WON"
	GameBetStatusLost      GameBetStatus = "LOST"
	GameBetStatusCashedOut GameBetStatus = "CASHED_OUT"
)

// CalculatePayout computes the payout based on cashout multiplier
func (gb *GameBet) CalculatePayout(multiplier decimal.Decimal) decimal.Decimal {
	return gb.Amount.Mul(multiplier)
}
