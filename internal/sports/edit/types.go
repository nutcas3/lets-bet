package edit

import (
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/shopspring/decimal"
)

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

// EditBetService handles bet editing operations
type EditBetService struct {
	betRepo       *postgres.SportBetRepository
	matchRepo     *postgres.MatchRepository
	marketRepo    *postgres.BettingMarketRepository
	outcomeRepo   *postgres.MarketOutcomeRepository
	walletService WalletService
	eventBus      EventBus
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data any) error
}

// EditBetRequest represents a bet edit request
type EditBetRequest struct {
	BetID        string          `json:"bet_id"`
	UserID       string          `json:"user_id"`
	NewAmount    decimal.Decimal `json:"new_amount"`
	NewOdds      decimal.Decimal `json:"new_odds"`
	NewOutcomeID string          `json:"new_outcome_id,omitempty"`
	Reason       string          `json:"reason"`
}

// EditBetResponse represents a bet edit response
type EditBetResponse struct {
	OriginalBet      *domain.SportBet `json:"original_bet"`
	EditedBet        *domain.SportBet `json:"edited_bet"`
	RefundAmount     decimal.Decimal  `json:"refund_amount"`
	AdditionalAmount decimal.Decimal  `json:"additional_amount"`
	EditedAt         time.Time        `json:"edited_at"`
	Reason           string           `json:"reason"`
}
