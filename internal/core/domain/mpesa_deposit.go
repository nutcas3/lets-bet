package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MPesaDeposit represents an M-Pesa deposit transaction
type MPesaDeposit struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	
	// M-Pesa transaction identifiers
	MerchantRequestID string     `json:"merchant_request_id" db:"merchant_request_id"`
	CheckoutRequestID string     `json:"checkout_request_id" db:"checkout_request_id"`
	
	// User and amount details
	UserID      uuid.UUID       `json:"user_id" db:"user_id"`
	PhoneNumber string          `json:"phone_number" db:"phone_number"`
	Amount      decimal.Decimal `json:"amount" db:"amount"`
	Currency    string          `json:"currency" db:"currency"`
	
	// Transaction status and metadata
	Status      MPesaDepositStatus `json:"status" db:"status"`
	Reference   string             `json:"reference" db:"reference"`
	Description string             `json:"description" db:"description"`
	
	// M-Pesa callback data
	MpesaReceiptNumber *string    `json:"mpesa_receipt_number,omitempty" db:"mpesa_receipt_number"`
	TransactionDate    *time.Time `json:"transaction_date,omitempty" db:"transaction_date"`
	ResultCode         *string    `json:"result_code,omitempty" db:"result_code"`
	ResultDesc         *string    `json:"result_desc,omitempty" db:"result_desc"`
	
	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type MPesaDepositStatus string

const (
	MPesaDepositStatusPending   MPesaDepositStatus = "PENDING"
	MPesaDepositStatusCompleted MPesaDepositStatus = "COMPLETED"
	MPesaDepositStatusFailed    MPesaDepositStatus = "FAILED"
	MPesaDepositStatusCancelled MPesaDepositStatus = "CANCELLED"
)

// IsCompleted returns true if the deposit is in a final state
func (d *MPesaDeposit) IsCompleted() bool {
	return d.Status == MPesaDepositStatusCompleted || 
		   d.Status == MPesaDepositStatusFailed || 
		   d.Status == MPesaDepositStatusCancelled
}

// IsSuccessful returns true if the deposit was completed successfully
func (d *MPesaDeposit) IsSuccessful() bool {
	return d.Status == MPesaDepositStatusCompleted
}

// MarkCompleted updates the deposit to completed status with M-Pesa details
func (d *MPesaDeposit) MarkCompleted(receiptNumber string, transactionDate time.Time) {
	d.Status = MPesaDepositStatusCompleted
	d.MpesaReceiptNumber = &receiptNumber
	d.TransactionDate = &transactionDate
	now := time.Now()
	d.CompletedAt = &now
	d.UpdatedAt = now
}

// MarkFailed updates the deposit to failed status
func (d *MPesaDeposit) MarkFailed(resultCode, resultDesc string) {
	d.Status = MPesaDepositStatusFailed
	d.ResultCode = &resultCode
	d.ResultDesc = &resultDesc
	now := time.Now()
	d.CompletedAt = &now
	d.UpdatedAt = now
}
