package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Wallet represents a user's balance with versioning for optimistic locking
type Wallet struct {
	ID       uuid.UUID       `json:"id" db:"id"`
	UserID   uuid.UUID       `json:"user_id" db:"user_id"`
	Currency string          `json:"currency" db:"currency"`
	Balance  decimal.Decimal `json:"balance" db:"balance"`
	Version  int64           `json:"version" db:"version"` // For optimistic locking

	// Bonus balance (separate from withdrawable)
	BonusBalance decimal.Decimal `json:"bonus_balance" db:"bonus_balance"`

	// Limits
	TodayDeposit     decimal.Decimal `json:"today_deposit" db:"today_deposit"`
	LastDepositReset time.Time       `json:"last_deposit_reset" db:"last_deposit_reset"`

	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Transaction represents any wallet movement
type Transaction struct {
	ID       uuid.UUID `json:"id" db:"id"`
	WalletID uuid.UUID `json:"wallet_id" db:"wallet_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`

	Type     TransactionType `json:"type" db:"type"`
	Amount   decimal.Decimal `json:"amount" db:"amount"`
	Currency string          `json:"currency" db:"currency"`

	// Balance snapshot (for auditing)
	BalanceBefore decimal.Decimal `json:"balance_before" db:"balance_before"`
	BalanceAfter  decimal.Decimal `json:"balance_after" db:"balance_after"`

	// Reference to source (bet, deposit, etc.)
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty" db:"reference_id"`
	ReferenceType string     `json:"reference_type" db:"reference_type"` // BET, DEPOSIT, WITHDRAWAL

	// Payment provider details (for deposits/withdrawals)
	ProviderTxnID string `json:"provider_txn_id,omitempty" db:"provider_txn_id"`
	ProviderName  string `json:"provider_name,omitempty" db:"provider_name"` // MPESA, AIRTEL

	Status      TransactionStatus `json:"status" db:"status"`
	Description string            `json:"description" db:"description"`

	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// Country-specific metadata
	CountryCode string `json:"country_code" db:"country_code"`
}

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeBetPlaced  TransactionType = "BET_PLACED"
	TransactionTypeBetWon     TransactionType = "BET_WON"
	TransactionTypeBetRefund  TransactionType = "BET_REFUND"
	TransactionTypeBonus      TransactionType = "BONUS"
	TransactionTypeTax        TransactionType = "TAX_DEDUCTION"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "PENDING"
	TransactionStatusCompleted TransactionStatus = "COMPLETED"
	TransactionStatusFailed    TransactionStatus = "FAILED"
	TransactionStatusCancelled TransactionStatus = "CANCELLED"
)

// CanWithdraw checks if wallet has sufficient balance for withdrawal
func (w *Wallet) CanWithdraw(amount decimal.Decimal) bool {
	return w.Balance.GreaterThanOrEqual(amount) && amount.GreaterThan(decimal.Zero)
}

// CanDeposit checks daily deposit limits (responsible gaming)
func (w *Wallet) CanDeposit(amount decimal.Decimal, limit *int64) bool {
	if limit == nil {
		return true
	}

	// Reset if new day
	if time.Since(w.LastDepositReset).Hours() >= 24 {
		return true
	}

	maxDeposit := decimal.NewFromInt(*limit)
	newTotal := w.TodayDeposit.Add(amount)

	return newTotal.LessThanOrEqual(maxDeposit)
}
