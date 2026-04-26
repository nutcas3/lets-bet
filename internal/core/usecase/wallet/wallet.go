// Package wallet implements the atomic-money-movement service for our platform.
//
// Every balance change goes through [Service] which:
//
//   - opens a single DB transaction,
//   - reads the wallet row FOR UPDATE (pessimistic lock against the same row
//     while also incrementing a `version` column for optimistic-lock-based
//     readers),
//   - writes the new balance with a WHERE clause on the old version,
//   - writes a Transaction audit row with BalanceBefore / BalanceAfter,
//   - commits.
//
// If anything fails mid-flight the transaction rolls back and the wallet is
// untouched. Callers never mutate the wallet directly.
package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Common errors surfaced to callers.
var (
	ErrWalletNotFound     = errors.New("wallet not found")
	ErrInsufficientFunds  = errors.New("insufficient funds")
	ErrInvalidAmount      = errors.New("amount must be positive")
	ErrOptimisticConflict = errors.New("wallet was modified concurrently")
)

// Service is the atomic wallet operations service.
type Service struct {
	db *sql.DB
}

// New constructs a wallet service bound to the given DB handle.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// Movement describes a balance change that should be applied atomically.
// Positive Amount credits the wallet; negative Amount debits.
type Movement struct {
	UserID        uuid.UUID
	Amount        decimal.Decimal // signed; positive = credit, negative = debit
	Type          domain.TransactionType
	ReferenceID   *uuid.UUID
	ReferenceType string
	Description   string
	ProviderName  string
	ProviderTxnID string
	CountryCode   string
}

// DBTX is the minimal *sql.Tx / *sql.DB subset the wallet movement needs.
// Using an interface lets [Service.ApplyTx] run against a caller-managed
// transaction so additional domain writes (e.g. bet insert) can be atomic with
// the wallet movement.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Apply runs the movement inside a single DB transaction it owns. Returns the
// persisted Transaction record (with before/after balance snapshots) on
// success.
func (s *Service) Apply(ctx context.Context, m Movement) (*domain.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit

	rec, err := s.ApplyTx(ctx, tx, m)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return rec, nil
}

// DB returns the underlying handle so callers can BeginTx when they want to
// span a wallet movement with other domain writes.
func (s *Service) DB() *sql.DB { return s.db }

// Debit removes amount from the wallet. Amount must be positive.
func (s *Service) Debit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, m Movement) (*domain.Transaction, error) {
	if !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}
	m.UserID = userID
	m.Amount = amount.Neg()
	return s.Apply(ctx, m)
}

// Credit adds amount to the wallet. Amount must be positive.
func (s *Service) Credit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, m Movement) (*domain.Transaction, error) {
	if !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}
	m.UserID = userID
	m.Amount = amount
	return s.Apply(ctx, m)
}

// Balance returns the wallet's current balance and bonus balance.
func (s *Service) Balance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	const q = `SELECT balance, bonus_balance FROM wallets WHERE user_id = $1`
	var balance, bonus decimal.Decimal
	if err := s.db.QueryRowContext(ctx, q, userID).Scan(&balance, &bonus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.Zero, decimal.Zero, ErrWalletNotFound
		}
		return decimal.Zero, decimal.Zero, err
	}
	return balance, bonus, nil
}
