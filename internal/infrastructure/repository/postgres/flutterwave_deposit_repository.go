package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// FlutterwaveDeposit represents a Flutterwave deposit in the database
type FlutterwaveDeposit struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	UserID         uuid.UUID       `json:"user_id" db:"user_id"`
	DepositID      string          `json:"deposit_id" db:"deposit_id"`
	TransactionID  string          `json:"transaction_id" db:"transaction_id"`
	Reference      string          `json:"reference" db:"reference"`
	PaymentLink    string          `json:"payment_link" db:"payment_link"`
	Amount         decimal.Decimal `json:"amount" db:"amount"`
	Currency       string          `json:"currency" db:"currency"`
	Email          string          `json:"email" db:"email"`
	PhoneNumber    string          `json:"phone_number" db:"phone_number"`
	Status         string          `json:"status" db:"status"`
	PaymentMethod  string          `json:"payment_method" db:"payment_method"`
	ProviderName   string          `json:"provider_name" db:"provider_name"`
	ProviderTxnID  string          `json:"provider_txn_id" db:"provider_txn_id"`
	FlutterwaveRef string          `json:"flutterwave_ref" db:"flutterwave_ref"`
	Network        string          `json:"network" db:"network"`
	ProcessedAt    *time.Time      `json:"processed_at" db:"processed_at"`
	CompletedAt    *time.Time      `json:"completed_at" db:"completed_at"`
	FailureReason  *string         `json:"failure_reason" db:"failure_reason"`
	AppFee         decimal.Decimal `json:"app_fee" db:"app_fee"`
	MerchantFee    decimal.Decimal `json:"merchant_fee" db:"merchant_fee"`
	TotalFees      decimal.Decimal `json:"total_fees" db:"total_fees"`
	Meta           json.RawMessage `json:"meta" db:"meta"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

// FlutterwaveDepositRepository provides database operations for Flutterwave deposits
type FlutterwaveDepositRepository struct {
	db *sql.DB
}

// NewFlutterwaveDepositRepository creates a new Flutterwave deposit repository
func NewFlutterwaveDepositRepository(db *sql.DB) *FlutterwaveDepositRepository {
	return &FlutterwaveDepositRepository{
		db: db,
	}
}

// Create creates a new Flutterwave deposit
func (r *FlutterwaveDepositRepository) Create(ctx context.Context, deposit *FlutterwaveDeposit) error {
	query := `
		INSERT INTO flutterwave_deposits (
			id, user_id, deposit_id, transaction_id, reference, payment_link,
			amount, currency, email, phone_number, status, payment_method,
			provider_name, provider_txn_id, flutterwave_ref, network,
			app_fee, merchant_fee, meta
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	_, err := r.db.Exec(query,
		deposit.ID, deposit.UserID, deposit.DepositID, deposit.TransactionID,
		deposit.Reference, deposit.PaymentLink, deposit.Amount, deposit.Currency,
		deposit.Email, deposit.PhoneNumber, deposit.Status, deposit.PaymentMethod,
		deposit.ProviderName, deposit.ProviderTxnID, deposit.FlutterwaveRef,
		deposit.Network, deposit.AppFee, deposit.MerchantFee, deposit.Meta,
	)

	if err != nil {
		return fmt.Errorf("failed to create flutterwave deposit: %w", err)
	}

	return nil
}

// GetByID retrieves a Flutterwave deposit by ID
func (r *FlutterwaveDepositRepository) GetByID(ctx context.Context, id uuid.UUID) (*FlutterwaveDeposit, error) {
	query := `
		SELECT 
			id, user_id, deposit_id, transaction_id, reference, payment_link,
			amount, currency, email, phone_number, status, payment_method,
			provider_name, provider_txn_id, flutterwave_ref, network,
			processed_at, completed_at, failure_reason, app_fee, merchant_fee, total_fees,
			meta, created_at, updated_at
		FROM flutterwave_deposits
		WHERE id = $1
	`

	var deposit FlutterwaveDeposit
	err := r.db.QueryRow(query, id).Scan(
		&deposit.ID, &deposit.UserID, &deposit.DepositID, &deposit.TransactionID,
		&deposit.Reference, &deposit.PaymentLink, &deposit.Amount, &deposit.Currency,
		&deposit.Email, &deposit.PhoneNumber, &deposit.Status, &deposit.PaymentMethod,
		&deposit.ProviderName, &deposit.ProviderTxnID, &deposit.FlutterwaveRef,
		&deposit.Network, &deposit.ProcessedAt, &deposit.CompletedAt,
		&deposit.FailureReason, &deposit.AppFee, &deposit.MerchantFee,
		&deposit.TotalFees, &deposit.Meta, &deposit.CreatedAt, &deposit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flutterwave deposit not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get flutterwave deposit: %w", err)
	}

	return &deposit, nil
}

// GetByDepositID retrieves a Flutterwave deposit by deposit ID
func (r *FlutterwaveDepositRepository) GetByDepositID(ctx context.Context, depositID string) (*FlutterwaveDeposit, error) {
	query := `
		SELECT 
			id, user_id, deposit_id, transaction_id, reference, payment_link,
			amount, currency, email, phone_number, status, payment_method,
			provider_name, provider_txn_id, flutterwave_ref, network,
			processed_at, completed_at, failure_reason, app_fee, merchant_fee, total_fees,
			meta, created_at, updated_at
		FROM flutterwave_deposits
		WHERE deposit_id = $1
	`

	var deposit FlutterwaveDeposit
	err := r.db.QueryRow(query, depositID).Scan(
		&deposit.ID, &deposit.UserID, &deposit.DepositID, &deposit.TransactionID,
		&deposit.Reference, &deposit.PaymentLink, &deposit.Amount, &deposit.Currency,
		&deposit.Email, &deposit.PhoneNumber, &deposit.Status, &deposit.PaymentMethod,
		&deposit.ProviderName, &deposit.ProviderTxnID, &deposit.FlutterwaveRef,
		&deposit.Network, &deposit.ProcessedAt, &deposit.CompletedAt,
		&deposit.FailureReason, &deposit.AppFee, &deposit.MerchantFee,
		&deposit.TotalFees, &deposit.Meta, &deposit.CreatedAt, &deposit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flutterwave deposit not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get flutterwave deposit: %w", err)
	}

	return &deposit, nil
}

// GetByTransactionID retrieves a Flutterwave deposit by Flutterwave transaction ID
func (r *FlutterwaveDepositRepository) GetByTransactionID(ctx context.Context, transactionID string) (*FlutterwaveDeposit, error) {
	query := `
		SELECT 
			id, user_id, deposit_id, transaction_id, reference, payment_link,
			amount, currency, email, phone_number, status, payment_method,
			provider_name, provider_txn_id, flutterwave_ref, network,
			processed_at, completed_at, failure_reason, app_fee, merchant_fee, total_fees,
			meta, created_at, updated_at
		FROM flutterwave_deposits
		WHERE transaction_id = $1
	`

	var deposit FlutterwaveDeposit
	err := r.db.QueryRow(query, transactionID).Scan(
		&deposit.ID, &deposit.UserID, &deposit.DepositID, &deposit.TransactionID,
		&deposit.Reference, &deposit.PaymentLink, &deposit.Amount, &deposit.Currency,
		&deposit.Email, &deposit.PhoneNumber, &deposit.Status, &deposit.PaymentMethod,
		&deposit.ProviderName, &deposit.ProviderTxnID, &deposit.FlutterwaveRef,
		&deposit.Network, &deposit.ProcessedAt, &deposit.CompletedAt,
		&deposit.FailureReason, &deposit.AppFee, &deposit.MerchantFee,
		&deposit.TotalFees, &deposit.Meta, &deposit.CreatedAt, &deposit.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flutterwave deposit not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get flutterwave deposit: %w", err)
	}

	return &deposit, nil
}

// UpdateStatus updates the status of a Flutterwave deposit
func (r *FlutterwaveDepositRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE flutterwave_deposits
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update flutterwave deposit status: %w", err)
	}

	return nil
}

// MarkCompleted updates a deposit to completed status
func (r *FlutterwaveDepositRepository) MarkCompleted(ctx context.Context, id uuid.UUID, flutterwaveRef string, fees decimal.Decimal) error {
	query := `
		UPDATE flutterwave_deposits
		SET status = 'COMPLETED', 
			flutterwave_ref = $2,
			app_fee = $3,
			completed_at = NOW(),
			processed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, flutterwaveRef, fees)
	if err != nil {
		return fmt.Errorf("failed to mark flutterwave deposit as completed: %w", err)
	}

	return nil
}

// MarkFailed updates a deposit to failed status
func (r *FlutterwaveDepositRepository) MarkFailed(ctx context.Context, id uuid.UUID, reason string) error {
	query := `
		UPDATE flutterwave_deposits
		SET status = 'FAILED', 
			failure_reason = $2,
			processed_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(query, id, reason)
	if err != nil {
		return fmt.Errorf("failed to mark flutterwave deposit as failed: %w", err)
	}

	return nil
}

// Delete deletes a Flutterwave deposit
func (r *FlutterwaveDepositRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM flutterwave_deposits WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flutterwave deposit: %w", err)
	}

	return nil
}
