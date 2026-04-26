package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ApplyTx runs the movement against an existing transaction. The caller owns
// the transaction lifecycle (commit / rollback).
func (s *Service) ApplyTx(ctx context.Context, tx DBTX, m Movement) (*domain.Transaction, error) {
	if m.Amount.IsZero() {
		return nil, ErrInvalidAmount
	}

	// Lock the wallet row for the duration of the transaction.
	const selectSQL = `
		SELECT id, currency, balance, version
		FROM wallets
		WHERE user_id = $1
		FOR UPDATE`

	var (
		walletID uuid.UUID
		currency string
		balance  decimal.Decimal
		version  int64
	)
	if err := tx.QueryRowContext(ctx, selectSQL, m.UserID).Scan(&walletID, &currency, &balance, &version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("load wallet: %w", err)
	}

	newBalance := balance.Add(m.Amount)
	if newBalance.IsNegative() {
		return nil, ErrInsufficientFunds
	}

	now := time.Now().UTC()

	// Optimistic-lock guard: update ONLY if version hasn't changed.
	const updateSQL = `
		UPDATE wallets
		SET balance    = $1,
		    version    = version + 1,
		    updated_at = $2
		WHERE id = $3 AND version = $4`

	res, err := tx.ExecContext(ctx, updateSQL, newBalance, now, walletID, version)
	if err != nil {
		return nil, fmt.Errorf("update wallet: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if affected == 0 {
		return nil, ErrOptimisticConflict
	}

	rec := s.createTransactionRecord(walletID, currency, balance, newBalance, m, now)

	const insertTxnSQL = `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency,
			balance_before, balance_after, reference_id, reference_type,
			provider_txn_id, provider_name, status, description,
			created_at, completed_at, country_code
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`

	if _, err := tx.ExecContext(ctx, insertTxnSQL,
		rec.ID, rec.WalletID, rec.UserID, rec.Type, rec.Amount,
		rec.Currency, rec.BalanceBefore, rec.BalanceAfter, rec.ReferenceID,
		rec.ReferenceType, rec.ProviderTxnID, rec.ProviderName, rec.Status,
		rec.Description, rec.CreatedAt, rec.CompletedAt, rec.CountryCode,
	); err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	return rec, nil
}

// createTransactionRecord creates a transaction record with balance snapshots
func (s *Service) createTransactionRecord(
	walletID uuid.UUID,
	currency string,
	balanceBefore decimal.Decimal,
	balanceAfter decimal.Decimal,
	m Movement,
	now time.Time,
) *domain.Transaction {
	return &domain.Transaction{
		ID:            uuid.New(),
		WalletID:      walletID,
		UserID:        m.UserID,
		Type:          m.Type,
		Amount:        m.Amount.Abs(),
		Currency:      currency,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		ReferenceID:   m.ReferenceID,
		ReferenceType: m.ReferenceType,
		ProviderTxnID: m.ProviderTxnID,
		ProviderName:  m.ProviderName,
		Status:        domain.TransactionStatusCompleted,
		Description:   m.Description,
		CreatedAt:     now,
		CompletedAt:   &now,
		CountryCode:   m.CountryCode,
	}
}
