// Package gh provides Ghana-specific payment adapters
package gh

import (
	"time"

	"github.com/shopspring/decimal"
)

// FlutterwavePaymentResponse represents Flutterwave payment response
type FlutterwavePaymentResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID       int    `json:"id"`
		TxRef    string `json:"tx_ref"`
		FlwRef   string `json:"flw_ref"`
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
		Customer struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
			Name  string `json:"name"`
		} `json:"customer"`
		PaymentLink string `json:"link"`
	} `json:"data"`
}

// FlutterwaveMobileMoneyResponse represents Mobile Money response
type FlutterwaveMobileMoneyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID          int    `json:"id"`
		TxRef       string `json:"tx_ref"`
		FlwRef      string `json:"flw_ref"`
		Amount      string `json:"amount"`
		Currency    string `json:"currency"`
		Status      string `json:"status"`
		Network     string `json:"network"`
		PhoneNumber string `json:"phone_number"`
		CreatedAt   string `json:"created_at"`
	} `json:"data"`
}

// FlutterwavePayoutResponse represents Flutterwave payout response
type FlutterwavePayoutResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID              int    `json:"id"`
		Reference       string `json:"reference"`
		Amount          string `json:"amount"`
		Currency        string `json:"currency"`
		Status          string `json:"status"`
		BeneficiaryName string `json:"beneficiary_name"`
		BankCode        string `json:"bank_code"`
		AccountNumber   string `json:"account_number"`
		CreatedAt       string `json:"created_at"`
	} `json:"data"`
}

// FlutterwaveVerifyResponse represents Flutterwave transaction verification
type FlutterwaveVerifyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID            int            `json:"id"`
		TxRef         string         `json:"tx_ref"`
		FlwRef        string         `json:"flw_ref"`
		Amount        string         `json:"amount"`
		Currency      string         `json:"currency"`
		Status        string         `json:"status"`
		PaymentType   string         `json:"payment_type"`
		PaymentMethod string         `json:"payment_method"`
		Customer      CustomerInfo   `json:"customer"`
		Meta          map[string]any `json:"meta"`
		CreatedAt     string         `json:"created_at"`
		ChargedAmount string         `json:"charged_amount"`
		AppFee        string         `json:"app_fee"`
		MerchantFee   string         `json:"merchant_fee"`
	} `json:"data"`
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
