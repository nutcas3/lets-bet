package postgres

import (
	"context"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// SportBetRepositoryInterface implements sport bet repository using PostgreSQL
type SportBetRepositoryInterface interface {
	// Basic CRUD operations
	Create(ctx context.Context, bet *domain.SportBet) error
	GetByID(ctx context.Context, id string) (*domain.SportBet, error)
	Update(ctx context.Context, bet *domain.SportBet) error
	Delete(ctx context.Context, id string) error

	// Query operations
	GetByUserID(ctx context.Context, userID string, filters *BetFilters) ([]*domain.SportBet, error)
	GetByEventID(ctx context.Context, eventID string, filters *BetFilters) ([]*domain.SportBet, error)
	GetByMarketID(ctx context.Context, marketID string, filters *BetFilters) ([]*domain.SportBet, error)
	GetByStatus(ctx context.Context, status domain.BetStatus, filters *BetFilters) ([]*domain.SportBet, error)

	// Advanced queries
	GetPendingBets(ctx context.Context, filters *BetFilters) ([]*domain.SportBet, error)
	GetSettledBets(ctx context.Context, filters *BetFilters) ([]*domain.SportBet, error)
	GetActiveBets(ctx context.Context, filters *BetFilters) ([]*domain.SportBet, error)

	// Settlement operations
	SettleBets(ctx context.Context, bets []*domain.SportBet) error
	MarkAsWon(ctx context.Context, betID string, payout decimal.Decimal) error
	MarkAsLost(ctx context.Context, betID string) error
	MarkAsVoid(ctx context.Context, betID string, reason string) error

	// Analytics and reporting
	GetBetStats(ctx context.Context, filters *BetFilters) (*BetStats, error)
	GetUserBetHistory(ctx context.Context, userID string, period *TimePeriod) ([]*domain.SportBet, error)
	GetEventBetSummary(ctx context.Context, eventID string) (*EventBetSummary, error)

	// Batch operations
	BatchCreate(ctx context.Context, bets []*domain.SportBet) error
	BatchUpdate(ctx context.Context, bets []*domain.SportBet) error
	BatchSettle(ctx context.Context, betIDs []string, results []*SettlementResult) error
}

// BetFilters represents filters for bet queries
type BetFilters struct {
	UserID    *string           `json:"user_id,omitempty"`
	EventID   *string           `json:"event_id,omitempty"`
	MarketID  *string           `json:"market_id,omitempty"`
	OutcomeID *string           `json:"outcome_id,omitempty"`
	Status    *domain.BetStatus `json:"status,omitempty"`
	Currency  *string           `json:"currency,omitempty"`
	MinAmount *decimal.Decimal  `json:"min_amount,omitempty"`
	MaxAmount *decimal.Decimal  `json:"max_amount,omitempty"`
	MinOdds   *decimal.Decimal  `json:"min_odds,omitempty"`
	MaxOdds   *decimal.Decimal  `json:"max_odds,omitempty"`
	From      *time.Time        `json:"from,omitempty"`
	To        *time.Time        `json:"to,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
	OrderBy   string            `json:"order_by,omitempty"`
	OrderDir  string            `json:"order_dir,omitempty"`
}

// TimePeriod represents a time period for queries
type TimePeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// BetStats represents betting statistics
type BetStats struct {
	TotalBets        int64           `json:"total_bets"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	TotalPayout      decimal.Decimal `json:"total_payout"`
	TotalNetPayout   decimal.Decimal `json:"total_net_payout"`
	WinningBets      int64           `json:"winning_bets"`
	LosingBets       int64           `json:"losing_bets"`
	VoidBets         int64           `json:"void_bets"`
	PendingBets      int64           `json:"pending_bets"`
	WinRate          decimal.Decimal `json:"win_rate"`
	AverageBetAmount decimal.Decimal `json:"average_bet_amount"`
	AverageOdds      decimal.Decimal `json:"average_odds"`
	TotalProfit      decimal.Decimal `json:"total_profit"`
	TotalLoss        decimal.Decimal `json:"total_loss"`
	ProfitMargin     decimal.Decimal `json:"profit_margin"`
}

// EventBetSummary represents betting summary for an event
type EventBetSummary struct {
	EventID        string          `json:"event_id"`
	TotalBets      int64           `json:"total_bets"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	TotalPayout    decimal.Decimal `json:"total_payout"`
	TotalLiability decimal.Decimal `json:"total_liability"`
	BetsByMarket   []MarketBets    `json:"bets_by_market"`
	BetsByOutcome  []OutcomeBets   `json:"bets_by_outcome"`
	PendingBets    int64           `json:"pending_bets"`
	SettledBets    int64           `json:"settled_bets"`
}

// MarketBets represents bets grouped by market
type MarketBets struct {
	MarketID       string          `json:"market_id"`
	MarketName     string          `json:"market_name"`
	BetCount       int64           `json:"bet_count"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
	TotalLiability decimal.Decimal `json:"total_liability"`
}

// OutcomeBets represents bets grouped by outcome
type OutcomeBets struct {
	OutcomeID   string          `json:"outcome_id"`
	OutcomeName string          `json:"outcome_name"`
	BetCount    int64           `json:"bet_count"`
	TotalAmount decimal.Decimal `json:"total_amount"`
	Odds        decimal.Decimal `json:"odds"`
}

// SettlementResult represents the result of bet settlement
type SettlementResult struct {
	BetID     string           `json:"bet_id"`
	Status    domain.BetStatus `json:"status"`
	Payout    decimal.Decimal  `json:"payout"`
	NetPayout decimal.Decimal  `json:"net_payout"`
	Reason    string           `json:"reason"`
	SettledAt time.Time        `json:"settled_at"`
}

// BetMetrics represents betting metrics
type BetMetrics struct {
	TotalActiveBets    int64           `json:"total_active_bets"`
	TotalPendingBets   int64           `json:"total_pending_bets"`
	TotalSettledBets   int64           `json:"total_settled_bets"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	TotalLiability     decimal.Decimal `json:"total_liability"`
	AverageBetSize     decimal.Decimal `json:"average_bet_size"`
	LastBetTime        time.Time       `json:"last_bet_time"`
	LastSettlementTime time.Time       `json:"last_settlement_time"`
	ActiveUsers        int64           `json:"active_users"`
	ActiveEvents       int64           `json:"active_events"`
	ActiveMarkets      int64           `json:"active_markets"`
}

// UserBetSummary represents a user's betting summary
type UserBetSummary struct {
	UserID           string          `json:"user_id"`
	TotalBets        int64           `json:"total_bets"`
	TotalAmount      decimal.Decimal `json:"total_amount"`
	TotalPayout      decimal.Decimal `json:"total_payout"`
	TotalProfit      decimal.Decimal `json:"total_profit"`
	WinRate          decimal.Decimal `json:"win_rate"`
	AverageBetAmount decimal.Decimal `json:"average_bet_amount"`
	BiggestWin       decimal.Decimal `json:"biggest_win"`
	BiggestLoss      decimal.Decimal `json:"biggest_loss"`
	LongestWinStreak int             `json:"longest_win_streak"`
	CurrentStreak    int             `json:"current_streak"`
	FirstBetTime     time.Time       `json:"first_bet_time"`
	LastBetTime      time.Time       `json:"last_bet_time"`
}

// EventLiability represents liability for an event
type EventLiability struct {
	EventID        string            `json:"event_id"`
	EventName      string            `json:"event_name"`
	TotalLiability decimal.Decimal   `json:"total_liability"`
	TotalBets      int64             `json:"total_bets"`
	TotalAmount    decimal.Decimal   `json:"total_amount"`
	Markets        []MarketLiability `json:"markets"`
	CalculatedAt   time.Time         `json:"calculated_at"`
}

// MarketLiability represents liability for a market
type MarketLiability struct {
	MarketID       string             `json:"market_id"`
	MarketName     string             `json:"market_name"`
	TotalLiability decimal.Decimal    `json:"total_liability"`
	TotalBets      int64              `json:"total_bets"`
	TotalAmount    decimal.Decimal    `json:"total_amount"`
	Outcomes       []OutcomeLiability `json:"outcomes"`
}

// OutcomeLiability represents liability for an outcome
type OutcomeLiability struct {
	OutcomeID      string          `json:"outcome_id"`
	OutcomeName    string          `json:"outcome_name"`
	Odds           decimal.Decimal `json:"odds"`
	TotalLiability decimal.Decimal `json:"total_liability"`
	TotalBets      int64           `json:"total_bets"`
	TotalAmount    decimal.Decimal `json:"total_amount"`
}

// BetValidationResult represents the result of bet validation
type BetValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// BetRequest represents a bet creation request
type BetRequest struct {
	UserID    string          `json:"user_id"`
	EventID   string          `json:"event_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
	Currency  string          `json:"currency"`
}

// BetResponse represents a bet creation response
type BetResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	BetID   string           `json:"bet_id,omitempty"`
	Bet     *domain.SportBet `json:"bet,omitempty"`
	Errors  []string         `json:"errors,omitempty"`
}

// SettlementRequest represents a batch settlement request
type SettlementRequest struct {
	EventID   string                    `json:"event_id"`
	MarketID  string                    `json:"market_id,omitempty"`
	Outcomes  []OutcomeSettlementResult `json:"outcomes"`
	SettledAt time.Time                 `json:"settled_at"`
	SettledBy string                    `json:"settled_by"`
	Reason    string                    `json:"reason,omitempty"`
}

// OutcomeSettlementResult represents settlement result for an outcome
type OutcomeSettlementResult struct {
	OutcomeID string          `json:"outcome_id"`
	Winner    bool            `json:"winner"`
	Odds      decimal.Decimal `json:"odds"`
	Reason    string          `json:"reason,omitempty"`
}
