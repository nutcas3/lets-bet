package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Bet represents a single betting slip
type Bet struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CountryCode string    `json:"country_code" db:"country_code"`

	// Bet details
	BetType      BetType         `json:"bet_type" db:"bet_type"`
	Stake        decimal.Decimal `json:"stake" db:"stake"` // Amount wagered
	Currency     string          `json:"currency" db:"currency"`
	PotentialWin decimal.Decimal `json:"potential_win" db:"potential_win"`
	TotalOdds    decimal.Decimal `json:"total_odds" db:"total_odds"` // Multiplied odds

	// Status and settlement
	Status    BetStatus       `json:"status" db:"status"`
	ActualWin decimal.Decimal `json:"actual_win" db:"actual_win"`
	SettledAt *time.Time      `json:"settled_at,omitempty" db:"settled_at"`

	// Selections (JSON in DB, but we'll use a separate table in production)
	Selections []Selection `json:"selections" db:"-"`

	// Metadata
	PlacedAt  time.Time `json:"placed_at" db:"placed_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	DeviceID  string    `json:"device_id" db:"device_id"`

	// Tax tracking (BCLB compliance)
	TaxAmount decimal.Decimal `json:"tax_amount" db:"tax_amount"`
	TaxPaid   bool            `json:"tax_paid" db:"tax_paid"`
}

type BetType string

const (
	BetTypeSingle BetType = "SINGLE" // One selection
	BetTypeMulti  BetType = "MULTI"  // All must win
	BetTypeSystem BetType = "SYSTEM" // Partial wins (e.g., 3/5)
)

type BetStatus string

const (
	BetStatusPending BetStatus = "PENDING"
	BetStatusWon     BetStatus = "WON"
	BetStatusLost    BetStatus = "LOST"
	BetStatusVoid    BetStatus = "VOID"
	BetStatusCashout BetStatus = "CASHED_OUT"
)

// Selection represents a single pick in a bet
type Selection struct {
	ID          uuid.UUID       `json:"id"`
	BetID       uuid.UUID       `json:"bet_id"`
	MarketID    string          `json:"market_id"`    // From odds provider
	EventID     string          `json:"event_id"`     // e.g., "match_12345"
	EventName   string          `json:"event_name"`   // e.g., "Arsenal vs Chelsea"
	MarketName  string          `json:"market_name"`  // e.g., "Match Winner"
	OutcomeName string          `json:"outcome_name"` // e.g., "Arsenal"
	Odds        decimal.Decimal `json:"odds"`
	Status      SelectionStatus `json:"status"`
	SettledAt   *time.Time      `json:"settled_at,omitempty"`
}

type SelectionStatus string

const (
	SelectionStatusPending SelectionStatus = "PENDING"
	SelectionStatusWon     SelectionStatus = "WON"
	SelectionStatusLost    SelectionStatus = "LOST"
	SelectionStatusVoid    SelectionStatus = "VOID"
)

// CalculatePotentialWin computes the potential payout based on bet type
func (b *Bet) CalculatePotentialWin() decimal.Decimal {
	switch b.BetType {
	case BetTypeSingle:
		return b.Stake.Mul(b.TotalOdds)
	case BetTypeMulti:
		return b.Stake.Mul(b.TotalOdds)
	case BetTypeSystem:
		// System bet calculation is more complex
		// For now, simplified version
		return b.Stake.Mul(b.TotalOdds)
	default:
		return decimal.Zero
	}
}

// CalculateTax applies Kenya's 20% withholding tax on winnings
func (b *Bet) CalculateTax(countryCode string) decimal.Decimal {
	if countryCode != "KE" {
		return decimal.Zero // Different tax rules per country
	}

	// Kenya: 20% tax on (winnings - stake)
	profit := b.ActualWin.Sub(b.Stake)
	if profit.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	return profit.Mul(decimal.NewFromFloat(0.20))
}

// SportBet represents a sports betting bet
type SportBet struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	CountryCode      string          `json:"country_code"`
	EventID          string          `json:"event_id"`
	MarketID         string          `json:"market_id"`
	OutcomeID        string          `json:"outcome_id"`
	Amount           decimal.Decimal `json:"amount"`
	Odds             decimal.Decimal `json:"odds"`
	Currency         string          `json:"currency"`
	Status           BetStatus       `json:"status"`
	Payout           decimal.Decimal `json:"payout"`
	NetPayout        decimal.Decimal `json:"net_payout"`
	SettledAt        *time.Time      `json:"settled_at,omitempty"`
	SettlementReason string          `json:"settlement_reason,omitempty"`
	PlacedAt         time.Time       `json:"placed_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}
