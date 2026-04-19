// Package africastalking provides Africa's Talking SMS and OTP integration
package africastalking

import (
	"context"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// AfricaTalkingConfig provides configuration for Africa's Talking client
type AfricaTalkingConfig struct {
	Username    string        `json:"username"`
	APIKey      string        `json:"api_key"`
	SenderName  string        `json:"sender_name"`
	Environment string        `json:"environment"` // "sandbox", "production"
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	RateLimit   int           `json:"rate_limit"` // requests per second
}

// DefaultAfricaTalkingConfig returns default configuration
func DefaultAfricaTalkingConfig() *AfricaTalkingConfig {
	return &AfricaTalkingConfig{
		Environment: "sandbox",
		BaseURL:     "https://api.africastalking.com",
		Timeout:     30 * time.Second,
		RateLimit:   10, // 10 requests per second
		SenderName:  "BettingPlatform",
	}
}

// AfricaTalkingClient provides Africa's Talking SMS and OTP integration
type AfricaTalkingClient struct {
	config      *AfricaTalkingConfig
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// RateLimiter implements rate limiting for API requests
type RateLimiter struct {
	tokens     int
	maxTokens  int
	interval   time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// Wait waits until a token is available
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.refill()

	for r.tokens <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval):
			r.refill()
		}
	}

	r.tokens--
	return nil
}

// refill adds tokens based on elapsed time
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	if elapsed >= r.interval {
		tokensToAdd := int(elapsed / r.interval)
		r.tokens += tokensToAdd
		if r.tokens > r.maxTokens {
			r.tokens = r.maxTokens
		}
		r.lastRefill = now
	}
}

// SMSRequest represents an SMS request
type SMSRequest struct {
	To      []string `json:"to"`
	Message string   `json:"message"`
	From    string   `json:"from,omitempty"`
}

// SMSResponse represents an SMS response
type SMSResponse struct {
	SMSMessageData SMSMessageData `json:"SMSMessageData"`
}

// SMSMessageData contains SMS message data
type SMSMessageData struct {
	Message    string   `json:"Message"`
	Recipients []string `json:"Recipients"`
}

// OTPRequest represents an OTP request
type OTPRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Length      int    `json:"length"`
	Validity    int    `json:"validity"`
}

// OTPResponse represents an OTP response
type OTPResponse struct {
	OTPCode     string `json:"otpCode"`
	PhoneNumber string `json:"phoneNumber"`
	Validity    int    `json:"validity"`
}

// OTPVerifyRequest represents an OTP verification request
type OTPVerifyRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	OTPCode     string `json:"otpCode"`
}

// OTPVerifyResponse represents an OTP verification response
type OTPVerifyResponse struct {
	Valid       bool   `json:"valid"`
	PhoneNumber string `json:"phoneNumber"`
	OTPCode     string `json:"otpCode"`
}

// VoiceRequest represents a voice call request
type VoiceRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Message   string `json:"message"`
	VoiceType string `json:"voiceType"` // "male", "female"
}

// VoiceResponse represents a voice call response
type VoiceResponse struct {
	ErrorMessage string `json:"errorMessage"`
	Status       string `json:"status"`
}

// USSDRequest represents a USSD request
type USSDRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Text        string `json:"text"`
	SessionID   string `json:"sessionId"`
}

// USSDResponse represents a USSD response
type USSDResponse struct {
	Response  string `json:"response"`
	SessionID string `json:"sessionId"`
	ShouldEnd bool   `json:"shouldEnd"`
}

// AirtimeRequest represents an airtime request
type AirtimeRequest struct {
	Recipients []AirtimeRecipient `json:"recipients"`
}

// AirtimeRecipient represents an airtime recipient
type AirtimeRecipient struct {
	PhoneNumber string          `json:"phoneNumber"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
}

// AirtimeResponse represents an airtime response
type AirtimeResponse struct {
	Responses []AirtimeRecipientResponse `json:"responses"`
}

// AirtimeRecipientResponse represents an airtime recipient response
type AirtimeRecipientResponse struct {
	PhoneNumber  string `json:"phoneNumber"`
	Amount       string `json:"amount"`
	Status       string `json:"status"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// PaymentRequest represents a mobile payment request
type PaymentRequest struct {
	ProductName  string          `json:"productName"`
	CurrencyCode string          `json:"currencyCode"`
	Amount       decimal.Decimal `json:"amount"`
	PhoneNumber  string          `json:"phoneNumber"`
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

// PaymentResponse represents a mobile payment response
type PaymentResponse struct {
	Status        string          `json:"status"`
	TransactionID string          `json:"transactionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	PhoneNumber   string          `json:"phoneNumber"`
	Message       string          `json:"message"`
}

// PaymentStatusRequest represents a payment status request
type PaymentStatusRequest struct {
	TransactionID string `json:"transactionId"`
}

// PaymentStatusResponse represents a payment status response
type PaymentStatusResponse struct {
	Status        string          `json:"status"`
	TransactionID string          `json:"transactionId"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	PhoneNumber   string          `json:"phoneNumber"`
	Message       string          `json:"message"`
	CreatedAt     time.Time       `json:"createdAt"`
}

// AfricaTalkingMetrics represents client metrics
type AfricaTalkingMetrics struct {
	TotalSMS        int64           `json:"total_sms"`
	SuccessfulSMS   int64           `json:"successful_sms"`
	FailedSMS       int64           `json:"failed_sms"`
	TotalOTP        int64           `json:"total_otp"`
	SuccessfulOTP   int64           `json:"successful_otp"`
	FailedOTP       int64           `json:"failed_otp"`
	TotalVoice      int64           `json:"total_voice"`
	SuccessfulVoice int64           `json:"successful_voice"`
	FailedVoice     int64           `json:"failed_voice"`
	TotalCost       decimal.Decimal `json:"total_cost"`
	LastActivity    time.Time       `json:"last_activity"`
}

// AfricaTalkingError represents an API error
type AfricaTalkingError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
