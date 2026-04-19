package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
)

// MPesaDepositRepository implements M-Pesa deposit repository using PostgreSQL
type MPesaDepositRepository struct {
	db *sql.DB
}

func NewMPesaDepositRepository(db *sql.DB) *MPesaDepositRepository {
	return &MPesaDepositRepository{db: db}
}

// Create stores a new M-Pesa deposit
func (r *MPesaDepositRepository) Create(ctx context.Context, deposit *domain.MPesaDeposit) error {
	query := `
		INSERT INTO mpesa_deposits (
			id, merchant_request_id, checkout_request_id, user_id, phone_number,
			amount, currency, status, reference, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		deposit.ID, deposit.MerchantRequestID, deposit.CheckoutRequestID,
		deposit.UserID, deposit.PhoneNumber, deposit.Amount, deposit.Currency,
		deposit.Status, deposit.Reference, deposit.Description,
	)

	if err != nil {
		log.Printf("Error creating M-Pesa deposit: %v", err)
		return err
	}

	return nil
}

// GetByCheckoutRequestID retrieves a deposit by checkout request ID
func (r *MPesaDepositRepository) GetByCheckoutRequestID(ctx context.Context, checkoutRequestID string) (*domain.MPesaDeposit, error) {
	query := `
		SELECT id, merchant_request_id, checkout_request_id, user_id, phone_number,
			   amount, currency, status, reference, description,
			   mpesa_receipt_number, transaction_date, result_code, result_desc,
			   created_at, updated_at, completed_at
		FROM mpesa_deposits WHERE checkout_request_id = $1
	`

	var deposit domain.MPesaDeposit
	var receiptNumber, resultCode, resultDesc sql.NullString
	var transactionDate, completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, checkoutRequestID).Scan(
		&deposit.ID, &deposit.MerchantRequestID, &deposit.CheckoutRequestID,
		&deposit.UserID, &deposit.PhoneNumber, &deposit.Amount, &deposit.Currency,
		&deposit.Status, &deposit.Reference, &deposit.Description,
		&receiptNumber, &transactionDate, &resultCode, &resultDesc,
		&deposit.CreatedAt, &deposit.UpdatedAt, &completedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deposit not found: %w", err)
		}
		return nil, err
	}

	if receiptNumber.Valid {
		deposit.MpesaReceiptNumber = &receiptNumber.String
	}
	if resultCode.Valid {
		deposit.ResultCode = &resultCode.String
	}
	if resultDesc.Valid {
		deposit.ResultDesc = &resultDesc.String
	}
	if transactionDate.Valid {
		deposit.TransactionDate = &transactionDate.Time
	}
	if completedAt.Valid {
		deposit.CompletedAt = &completedAt.Time
	}

	return &deposit, nil
}

// GetByUserID retrieves all deposits for a user
func (r *MPesaDepositRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.MPesaDeposit, error) {
	query := `
		SELECT id, merchant_request_id, checkout_request_id, user_id, phone_number,
			   amount, currency, status, reference, description,
			   mpesa_receipt_number, transaction_date, result_code, result_desc,
			   created_at, updated_at, completed_at
		FROM mpesa_deposits WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*domain.MPesaDeposit

	for rows.Next() {
		var deposit domain.MPesaDeposit
		var receiptNumber, resultCode, resultDesc sql.NullString
		var transactionDate, completedAt sql.NullTime

		err := rows.Scan(
			&deposit.ID, &deposit.MerchantRequestID, &deposit.CheckoutRequestID,
			&deposit.UserID, &deposit.PhoneNumber, &deposit.Amount, &deposit.Currency,
			&deposit.Status, &deposit.Reference, &deposit.Description,
			&receiptNumber, &transactionDate, &resultCode, &resultDesc,
			&deposit.CreatedAt, &deposit.UpdatedAt, &completedAt,
		)

		if err != nil {
			return nil, err
		}

		if receiptNumber.Valid {
			deposit.MpesaReceiptNumber = &receiptNumber.String
		}
		if resultCode.Valid {
			deposit.ResultCode = &resultCode.String
		}
		if resultDesc.Valid {
			deposit.ResultDesc = &resultDesc.String
		}
		if transactionDate.Valid {
			deposit.TransactionDate = &transactionDate.Time
		}
		if completedAt.Valid {
			deposit.CompletedAt = &completedAt.Time
		}

		deposits = append(deposits, &deposit)
	}

	return deposits, nil
}

// UpdateStatus updates the deposit status and related fields
func (r *MPesaDepositRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.MPesaDepositStatus, receiptNumber *string, transactionDate *time.Time, resultCode *string, resultDesc *string) error {
	query := `
		UPDATE mpesa_deposits 
		SET status = $1, 
		    mpesa_receipt_number = $2,
		    transaction_date = $3,
		    result_code = $4,
		    result_desc = $5,
		    completed_at = CASE WHEN $1 IN ('COMPLETED', 'FAILED', 'CANCELLED') THEN CURRENT_TIMESTAMP ELSE completed_at END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $6
	`

	_, err := r.db.ExecContext(ctx, query, status, receiptNumber, transactionDate, resultCode, resultDesc, id)
	if err != nil {
		log.Printf("Error updating M-Pesa deposit status: %v", err)
		return err
	}

	return nil
}

// GetPendingDeposits retrieves all pending deposits older than the specified duration
func (r *MPesaDepositRepository) GetPendingDeposits(ctx context.Context, olderThan time.Duration) ([]*domain.MPesaDeposit, error) {
	query := `
		SELECT id, merchant_request_id, checkout_request_id, user_id, phone_number,
			   amount, currency, status, reference, description,
			   mpesa_receipt_number, transaction_date, result_code, result_desc,
			   created_at, updated_at, completed_at
		FROM mpesa_deposits 
		WHERE status = 'PENDING' AND created_at < $1
		ORDER BY created_at ASC
	`

	cutoff := time.Now().Add(-olderThan)
	rows, err := r.db.QueryContext(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*domain.MPesaDeposit

	for rows.Next() {
		var deposit domain.MPesaDeposit
		var receiptNumber, resultCode, resultDesc sql.NullString
		var transactionDate, completedAt sql.NullTime

		err := rows.Scan(
			&deposit.ID, &deposit.MerchantRequestID, &deposit.CheckoutRequestID,
			&deposit.UserID, &deposit.PhoneNumber, &deposit.Amount, &deposit.Currency,
			&deposit.Status, &deposit.Reference, &deposit.Description,
			&receiptNumber, &transactionDate, &resultCode, &resultDesc,
			&deposit.CreatedAt, &deposit.UpdatedAt, &completedAt,
		)

		if err != nil {
			return nil, err
		}

		if receiptNumber.Valid {
			deposit.MpesaReceiptNumber = &receiptNumber.String
		}
		if resultCode.Valid {
			deposit.ResultCode = &resultCode.String
		}
		if resultDesc.Valid {
			deposit.ResultDesc = &resultDesc.String
		}
		if transactionDate.Valid {
			deposit.TransactionDate = &transactionDate.Time
		}
		if completedAt.Valid {
			deposit.CompletedAt = &completedAt.Time
		}

		deposits = append(deposits, &deposit)
	}

	return deposits, nil
}

// CountByUserID returns the total number of deposits for a user
func (r *MPesaDepositRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM mpesa_deposits WHERE user_id = $1`

	var count int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
