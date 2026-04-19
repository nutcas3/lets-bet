// Package ng provides Nigeria-specific payment adapters
package ng

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// FlutterwaveConfig provides configuration for Flutterwave adapter
type FlutterwaveConfig struct {
	SecretKey     string `json:"secret_key"`
	PublicKey     string `json:"public_key"`
	Environment   string `json:"environment"` // "live" or "test"
	BaseURL       string `json:"base_url"`
	WebhookSecret string `json:"webhook_secret"`
	Currency      string `json:"currency"` // NGN for Nigeria
}

// DefaultFlutterwaveConfig returns default configuration
func DefaultFlutterwaveConfig() *FlutterwaveConfig {
	return &FlutterwaveConfig{
		Environment: "test",
		BaseURL:     "https://api.flutterwave.com/v3",
		Currency:    "NGN",
	}
}

// FlutterwaveAdapter provides Flutterwave payment processing for Nigeria
type FlutterwaveAdapter struct {
	config *FlutterwaveConfig
	client *http.Client
}

// PaymentRequest represents a Flutterwave payment request
type PaymentRequest struct {
	TxRef         string          `json:"tx_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	CustomerEmail string          `json:"customer_email"`
	CustomerPhone string          `json:"customer_phone"`
	CustomerName  string          `json:"customer_name"`
	PaymentType   string          `json:"payment_type"` // "card", "banktransfer", "mobilemoney"
	RedirectURL   string          `json:"redirect_url"`
}

// PaymentResponse represents a Flutterwave payment response
type PaymentResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    PaymentData `json:"data"`
}

// PaymentData contains payment response data
type PaymentData struct {
	ID            string          `json:"id"`
	TxRef         string          `json:"tx_ref"`
	FlwRef        string          `json:"flw_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	PaymentType   string          `json:"payment_type"`
	Customer      CustomerInfo    `json:"customer"`
	CreatedAt     time.Time       `json:"created_at"`
	ChargedAmount decimal.Decimal `json:"charged_amount"`
	AppFee        decimal.Decimal `json:"app_fee"`
	MerchantFee   decimal.Decimal `json:"merchant_fee"`
}

// CustomerInfo contains customer information
type CustomerInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// TransactionVerificationRequest represents a verification request
type TransactionVerificationRequest struct {
	ID string `json:"id"`
}

// TransactionVerificationResponse represents a verification response
type TransactionVerificationResponse struct {
	Status  string                      `json:"status"`
	Message string                      `json:"message"`
	Data    TransactionVerificationData `json:"data"`
}

// TransactionVerificationData contains verification data
type TransactionVerificationData struct {
	ID            string          `json:"id"`
	TxRef         string          `json:"tx_ref"`
	FlwRef        string          `json:"flw_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	PaymentType   string          `json:"payment_type"`
	Customer      CustomerInfo    `json:"customer"`
	CreatedAt     time.Time       `json:"created_at"`
	ChargedAmount decimal.Decimal `json:"charged_amount"`
	AppFee        decimal.Decimal `json:"app_fee"`
	MerchantFee   decimal.Decimal `json:"merchant_fee"`
}

// ProcessResponse contains processing response details
type ProcessResponse struct {
	Response string `json:"response"`
}

// RefundRequest represents a refund request
type RefundRequest struct {
	ID     string          `json:"id"`
	Amount decimal.Decimal `json:"amount"`
	Reason string          `json:"reason"`
}

// RefundResponse represents a refund response
type RefundResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Data    RefundData `json:"data"`
}

// RefundData contains refund response data
type RefundData struct {
	ID        string          `json:"id"`
	TxRef     string          `json:"tx_ref"`
	FlwRef    string          `json:"flw_ref"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	Status    string          `json:"status"`
	RefundRef string          `json:"refund_ref"`
	CreatedAt time.Time       `json:"created_at"`
}

// WebhookEvent represents a Flutterwave webhook event
type WebhookEvent struct {
	Event string      `json:"event"`
	Data  WebhookData `json:"data"`
}

// WebhookData contains webhook event data
type WebhookData struct {
	ID            string          `json:"id"`
	TxRef         string          `json:"tx_ref"`
	FlwRef        string          `json:"flw_ref"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	PaymentType   string          `json:"payment_type"`
	Customer      CustomerInfo    `json:"customer"`
	CreatedAt     time.Time       `json:"created_at"`
	ChargedAmount decimal.Decimal `json:"charged_amount"`
	AppFee        decimal.Decimal `json:"app_fee"`
	MerchantFee   decimal.Decimal `json:"merchant_fee"`
}

// FlutterwaveMetrics represents Flutterwave adapter metrics
type FlutterwaveMetrics struct {
	TotalTransactions      int64           `json:"total_transactions"`
	SuccessfulTransactions int64           `json:"successful_transactions"`
	FailedTransactions     int64           `json:"failed_transactions"`
	TotalAmount            decimal.Decimal `json:"total_amount"`
	SuccessRate            decimal.Decimal `json:"success_rate"`
	AverageAmount          decimal.Decimal `json:"average_amount"`
	LastTransactionTime    time.Time       `json:"last_transaction_time"`
}
