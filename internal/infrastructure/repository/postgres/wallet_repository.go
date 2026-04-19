package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// WalletRepository implements wallet repository using PostgreSQL
type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	query := `
		INSERT INTO wallets (
			id, user_id, currency, balance, version, bonus_balance,
			today_deposit, last_deposit_reset, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID, wallet.UserID, wallet.Currency, wallet.Balance, wallet.Version, wallet.BonusBalance,
		wallet.TodayDeposit, wallet.LastDepositReset, wallet.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating wallet: %v", err)
		return err
	}

	return nil
}

func (r *WalletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	query := `
		SELECT id, user_id, currency, balance, version, bonus_balance,
			   today_deposit, last_deposit_reset, updated_at
		FROM wallets WHERE user_id = $1
	`

	var wallet domain.Wallet

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Currency, &wallet.Balance, &wallet.Version, &wallet.BonusBalance,
		&wallet.TodayDeposit, &wallet.LastDepositReset, &wallet.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, userID uuid.UUID, newBalance, newBonusBalance decimal.Decimal) error {
	query := `
		UPDATE wallets SET 
			balance = $2, bonus_balance = $3, updated_at = $4
		WHERE user_id = $1
	`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, userID, newBalance, newBonusBalance, now)
	if err != nil {
		log.Printf("Error updating wallet balance: %v", err)
		return err
	}

	return nil
}

func (r *WalletRepository) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency,
			balance_before, balance_after, reference_id, reference_type,
			provider_txn_id, provider_name, status, description,
			created_at, completed_at, country_code
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.WalletID, transaction.UserID, transaction.Type, transaction.Amount,
		transaction.Currency, transaction.BalanceBefore, transaction.BalanceAfter, transaction.ReferenceID, transaction.ReferenceType,
		transaction.ProviderTxnID, transaction.ProviderName, transaction.Status, transaction.Description,
		transaction.CreatedAt, transaction.CompletedAt, transaction.CountryCode,
	)

	if err != nil {
		log.Printf("Error creating transaction: %v", err)
		return err
	}

	return nil
}

func (r *WalletRepository) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, wallet_id, user_id, type, amount, currency,
			   balance_before, balance_after, reference_id, reference_type,
			   provider_txn_id, provider_name, status, description,
			   created_at, completed_at, country_code
		FROM transactions WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction

	for rows.Next() {
		var transaction domain.Transaction
		var referenceID sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&transaction.ID, &transaction.WalletID, &transaction.UserID, &transaction.Type,
			&transaction.Amount, &transaction.Currency, &transaction.BalanceBefore, &transaction.BalanceAfter,
			&referenceID, &transaction.ReferenceType, &transaction.ProviderTxnID, &transaction.ProviderName,
			&transaction.Status, &transaction.Description, &transaction.CreatedAt, &completedAt, &transaction.CountryCode,
		)

		if referenceID.Valid {
			refUUID, err := uuid.Parse(referenceID.String)
			if err == nil {
				transaction.ReferenceID = &refUUID
			}
		}

		if completedAt.Valid {
			transaction.CompletedAt = &completedAt.Time
		}

		if err != nil {
			return nil, err
		}

		transactions = append(transactions, &transaction)
	}

	return transactions, nil
}

func (r *WalletRepository) GetBalance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	query := `
		SELECT balance, bonus_balance FROM wallets WHERE user_id = $1
	`

	var balance, bonusBalance decimal.Decimal

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance, &bonusBalance)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}

	return balance, bonusBalance, nil
}

func (r *WalletRepository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status domain.TransactionStatus) error {
	query := `UPDATE transactions SET status = $2, updated_at = $3 WHERE id = $1`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, id, status, now)
	if err != nil {
		log.Printf("Error updating transaction status: %v", err)
		return err
	}

	return nil
}
