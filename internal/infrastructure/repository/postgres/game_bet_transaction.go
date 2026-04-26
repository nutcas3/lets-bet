package postgres

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// CreateBetWithWalletUpdate implements atomic wallet update + bet creation in transaction
func (r *GameBetRepository) CreateBetWithWalletUpdate(ctx context.Context, bet *domain.GameBet, userID uuid.UUID, amount decimal.Decimal) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return uuid.Nil, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	walletUpdateQuery := `
		UPDATE wallets 
		SET balance = balance - $1, updated_at = $2
		WHERE user_id = $3 AND balance >= $1
		RETURNING balance
	`

	var newBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, walletUpdateQuery, amount, time.Now(), userID).Scan(&newBalance)
	if err != nil {
		log.Printf("Error updating wallet balance: %v", err)
		return uuid.Nil, fmt.Errorf("insufficient balance or wallet error: %w", err)
	}

	betCreateQuery := `
		INSERT INTO game_bets (
			id, game_id, user_id, amount, currency, cashed_out, cashout_at,
			payout, status, placed_at, cashed_out_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var returnedID uuid.UUID
	err = tx.QueryRowContext(ctx, betCreateQuery,
		bet.ID, bet.GameID, bet.UserID, bet.Amount, bet.Currency,
		bet.CashedOut, bet.CashoutAt, bet.Payout, bet.Status,
		bet.PlacedAt, bet.CashedOutAt,
	).Scan(&returnedID)

	if err != nil {
		log.Printf("Error creating game bet: %v", err)
		return uuid.Nil, fmt.Errorf("bet creation failed: %w", err)
	}

	transactionQuery := `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency, 
			balance_before, balance_after, reference_id, reference_type,
			status, description, created_at, country_code
		) 
		SELECT $1, w.id, $2, $3, $4, $5, 
			   w.balance + $4, w.balance, $6, $7,
			   $8, $9, $10, $11
		FROM wallets w WHERE w.user_id = $2
	`

	transactionID := uuid.New()
	_, err = tx.ExecContext(ctx, transactionQuery,
		transactionID, userID, "BET_PLACED", amount, bet.Currency,
		bet.ID, "BET", "COMPLETED", "Bet placed", time.Now(), "US",
	)
	if err != nil {
		log.Printf("Error creating transaction record: %v", err)
		return uuid.Nil, fmt.Errorf("transaction recording failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return uuid.Nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	err = nil
	log.Printf("Successfully created bet %s and updated wallet for user %s, new balance: %s",
		returnedID.String(), userID.String(), newBalance.String())

	return returnedID, nil
}

// AtomicAutoCashoutWithCredit implements atomic auto-cashout with wallet credit in a single transaction
func (r *GameBetRepository) AtomicAutoCashoutWithCredit(ctx context.Context, id uuid.UUID, userID uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal, country string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction for auto cashout: %v", err)
		return false, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	betUpdateQuery := `
		UPDATE game_bets
		SET cashed_out = true, cashout_at = $1, payout = $2, status = $3, cashed_out_at = $4
		WHERE id = $5 AND status = $6
	`

	now := time.Now()
	result, err := tx.ExecContext(ctx, betUpdateQuery, cashoutAt, payout, domain.GameBetStatusCashedOut, now, id, domain.GameBetStatusActive)
	if err != nil {
		log.Printf("Error updating auto cashout bet: %v", err)
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false, err
	}

	if rowsAffected == 0 {
		return false, nil
	}

	walletUpdateQuery := `
		UPDATE wallets
		SET balance = balance + $1, bonus_balance = bonus_balance + $2, updated_at = $3
		WHERE user_id = $4
		RETURNING balance
	`

	var newBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, walletUpdateQuery, payout, decimal.Zero, now, userID).Scan(&newBalance)
	if err != nil {
		log.Printf("Error crediting wallet for auto cashout bet %s: %v", id, err)
		return false, fmt.Errorf("wallet credit failed: %w", err)
	}

	transactionQuery := `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency,
			balance_before, balance_after, reference_id, reference_type,
			status, description, created_at, country_code
		)
		SELECT $1, w.id, $2, $3, $4, 'KES',
			   w.balance - $4, w.balance, $5, $6,
			   $7, $8, $9, $10
		FROM wallets w WHERE w.user_id = $2
	`

	transactionID := uuid.New()
	_, err = tx.ExecContext(ctx, transactionQuery,
		transactionID, userID, domain.TransactionTypeBetWon, payout,
		id, "BET", "COMPLETED", "Auto cashout payout", now, country,
	)
	if err != nil {
		log.Printf("Error creating transaction record for auto cashout: %v", err)
		return false, fmt.Errorf("transaction recording failed: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing auto cashout transaction: %v", err)
		return false, fmt.Errorf("transaction commit failed: %w", err)
	}

	err = nil
	log.Printf("Successfully processed auto cashout for bet %s, credited %s to user %s, new balance: %s",
		id.String(), payout.String(), userID.String(), newBalance.String())

	return true, nil
}
