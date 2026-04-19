// Package africastalking provides Africa's Talking SMS and OTP integration
package africastalking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

// NewAfricaTalkingClient creates a new Africa's Talking client
func NewAfricaTalkingClient(config *AfricaTalkingConfig) *AfricaTalkingClient {
	if config == nil {
		config = DefaultAfricaTalkingConfig()
	}

	return &AfricaTalkingClient{
		config:      config,
		httpClient:  &http.Client{Timeout: config.Timeout},
		rateLimiter: NewRateLimiter(config.RateLimit, time.Second),
	}
}

// SMSRequest represents an SMS request
type SMSRequest struct {
	To          []string `json:"to"`
	Message     string   `json:"message"`
	From        string   `json:"from,omitempty"`
	BulkSMSMode string   `json:"bulkSMSMode,omitempty"`
}

// SMSResponse represents SMS response
type SMSResponse struct {
	SMSMessageData struct {
		Message    string `json:"message"`
		Recipients []struct {
			StatusCode string `json:"statusCode"`
			Number     string `json:"number"`
			MessageID  string `json:"messageId"`
			Cost       string `json:"cost"`
			Status     string `json:"status"`
		} `json:"recipients"`
	} `json:"SMSMessageData"`
}

// OTPRequest represents an OTP request
type OTPRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	BrandName   string `json:"brandName"`
	Length      int    `json:"length"`     // OTP length
	TimeToLive  int    `json:"timeToLive"` // TTL in seconds
}

// OTPResponse represents OTP response
type OTPResponse struct {
	TransactionID string `json:"transactionId"`
	PhoneNumber   string `json:"phoneNumber"`
	Status        string `json:"status"`
	Description   string `json:"description"`
}

// OTPVerifyRequest represents OTP verification request
type OTPVerifyRequest struct {
	TransactionID string `json:"transactionId"`
	OTPCode       string `json:"otpCode"`
}

// OTPVerifyResponse represents OTP verification response
type OTPVerifyResponse struct {
	TransactionID string `json:"transactionId"`
	PhoneNumber   string `json:"phoneNumber"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	ValidUntil    string `json:"validUntil"`
}

// USSDRequest represents a USSD request
type USSDRequest struct {
	PhoneNumber string `json:"phoneNumber"`
	Text        string `json:"text"`
	SessionID   string `json:"sessionId"`
}

// USSDResponse represents USSD response
type USSDResponse struct {
	StatusCode string `json:"statusCode"`
	Message    string `json:"message"`
	SessionID  string `json:"sessionId"`
}

// VoiceRequest represents a voice call request
type VoiceRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
	Record  bool   `json:"record"`
}

// VoiceResponse represents voice response
type VoiceResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	CallID   string `json:"callId"`
	Duration string `json:"duration"`
}

// RateLimiter provides rate limiting for API requests
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
	for {
		if r.tryConsume() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval / time.Duration(r.maxTokens)):
			// Wait for a short period before retrying
		}
	}
}

// tryConsume tries to consume a token
func (r *RateLimiter) tryConsume() bool {
	now := time.Now()
	// Refill tokens based on time elapsed
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.interval)
	if tokensToAdd > 0 {
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// SendSMS sends SMS messages
func (a *AfricaTalkingClient) SendSMS(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Set default sender name
	if req.From == "" {
		req.From = a.config.SenderName
	}

	// Prepare form data
	data := url.Values{}
	data.Set("username", a.config.Username)
	data.Set("to", strings.Join(req.To, ","))
	data.Set("message", req.Message)
	data.Set("from", req.From)

	if req.BulkSMSMode != "" {
		data.Set("bulkSMSMode", req.BulkSMSMode)
	}

	// Create request
	url := fmt.Sprintf("%s/version1/messaging", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var smsResp SMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &smsResp, nil
}

// SendOTP sends an OTP code
func (a *AfricaTalkingClient) SendOTP(ctx context.Context, req *OTPRequest) (*OTPResponse, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Set default values
	if req.BrandName == "" {
		req.BrandName = a.config.SenderName
	}
	if req.Length == 0 {
		req.Length = 6
	}
	if req.TimeToLive == 0 {
		req.TimeToLive = 300 // 5 minutes
	}

	// Prepare request body
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/version1/otp/send", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var otpResp OTPResponse
	if err := json.NewDecoder(resp.Body).Decode(&otpResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &otpResp, nil
}

// VerifyOTP verifies an OTP code
func (a *AfricaTalkingClient) VerifyOTP(ctx context.Context, req *OTPVerifyRequest) (*OTPVerifyResponse, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Prepare request body
	requestBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/version1/otp/verify", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var verifyResp OTPVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &verifyResp, nil
}

// SendUSSD sends USSD request
func (a *AfricaTalkingClient) SendUSSD(ctx context.Context, req *USSDRequest) (*USSDResponse, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Prepare form data
	data := url.Values{}
	data.Set("username", a.config.Username)
	data.Set("phone", req.PhoneNumber)
	data.Set("text", req.Text)
	if req.SessionID != "" {
		data.Set("sessionId", req.SessionID)
	}

	// Create request
	url := fmt.Sprintf("%s/version1/ussd", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var ussdResp USSDResponse
	if err := json.NewDecoder(resp.Body).Decode(&ussdResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ussdResp, nil
}

// MakeVoiceCall makes a voice call
func (a *AfricaTalkingClient) MakeVoiceCall(ctx context.Context, req *VoiceRequest) (*VoiceResponse, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	// Prepare form data
	data := url.Values{}
	data.Set("username", a.config.Username)
	data.Set("from", req.From)
	data.Set("to", req.To)
	data.Set("message", req.Message)
	if req.Record {
		data.Set("record", "true")
	}

	// Create request
	url := fmt.Sprintf("%s/version1/call", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var voiceResp VoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&voiceResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &voiceResp, nil
}

// GetBalance retrieves account balance
func (a *AfricaTalkingClient) GetBalance(ctx context.Context) (decimal.Decimal, error) {
	// Wait for rate limiter
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return decimal.Zero, fmt.Errorf("rate limit error: %w", err)
	}

	// Create request
	url := fmt.Sprintf("%s/version1/user", a.config.BaseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("apiKey", a.config.APIKey)

	// Make request
	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("Africa's Talking API error: %d", resp.StatusCode)
	}

	var balanceResp struct {
		UserData struct {
			Balance string `json:"balance"`
		} `json:"userData"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&balanceResp); err != nil {
		return decimal.Zero, fmt.Errorf("failed to decode response: %w", err)
	}

	balance, err := decimal.NewFromString(balanceResp.UserData.Balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to parse balance: %w", err)
	}

	return balance, nil
}

// SendWelcomeSMS sends a welcome SMS to new users
func (a *AfricaTalkingClient) SendWelcomeSMS(ctx context.Context, phoneNumber, userName string) error {
	message := fmt.Sprintf("Welcome to %s! Your account has been successfully created. Happy betting!", a.config.SenderName)
	if userName != "" {
		message = fmt.Sprintf("Welcome %s to %s! Your account has been successfully created. Happy betting!", userName, a.config.SenderName)
	}

	req := &SMSRequest{
		To:      []string{phoneNumber},
		Message: message,
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send welcome SMS: %w", err)
	}

	return nil
}

// SendVerificationOTP sends OTP for phone verification
func (a *AfricaTalkingClient) SendVerificationOTP(ctx context.Context, phoneNumber string) (*OTPResponse, error) {
	req := &OTPRequest{
		PhoneNumber: phoneNumber,
		BrandName:   a.config.SenderName,
		Length:      6,
		TimeToLive:  300, // 5 minutes
	}

	return a.SendOTP(ctx, req)
}

// SendDepositConfirmationSMS sends SMS confirmation for deposits
func (a *AfricaTalkingClient) SendDepositConfirmationSMS(ctx context.Context, phoneNumber, amount, currency string) error {
	message := fmt.Sprintf("Your deposit of %s %s has been successfully processed and credited to your account. Thank you for choosing %s!", amount, currency, a.config.SenderName)

	req := &SMSRequest{
		To:      []string{phoneNumber},
		Message: message,
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send deposit confirmation SMS: %w", err)
	}

	return nil
}

// SendWithdrawalConfirmationSMS sends SMS confirmation for withdrawals
func (a *AfricaTalkingClient) SendWithdrawalConfirmationSMS(ctx context.Context, phoneNumber, amount, currency string) error {
	message := fmt.Sprintf("Your withdrawal of %s %s has been successfully processed. The funds should reflect in your account shortly. Thank you for choosing %s!", amount, currency, a.config.SenderName)

	req := &SMSRequest{
		To:      []string{phoneNumber},
		Message: message,
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send withdrawal confirmation SMS: %w", err)
	}

	return nil
}

// SendBetConfirmationSMS sends SMS confirmation for bets
func (a *AfricaTalkingClient) SendBetConfirmationSMS(ctx context.Context, phoneNumber, betID, amount, odds string) error {
	message := fmt.Sprintf("Bet %s placed successfully! Amount: %s, Odds: %s. Good luck from %s!", betID, amount, odds, a.config.SenderName)

	req := &SMSRequest{
		To:      []string{phoneNumber},
		Message: message,
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send bet confirmation SMS: %w", err)
	}

	return nil
}

// SendWinNotificationSMS sends SMS notification for winning bets
func (a *AfricaTalkingClient) SendWinNotificationSMS(ctx context.Context, phoneNumber, betID, payout string) error {
	message := fmt.Sprintf("Congratulations! Your bet %s has won! Payout: %s. Claim your winnings now! - %s", betID, payout, a.config.SenderName)

	req := &SMSRequest{
		To:      []string{phoneNumber},
		Message: message,
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send win notification SMS: %w", err)
	}

	return nil
}

// SendPromotionalSMS sends promotional SMS messages
func (a *AfricaTalkingClient) SendPromotionalSMS(ctx context.Context, phoneNumbers []string, message string) error {
	req := &SMSRequest{
		To:          phoneNumbers,
		Message:     message,
		BulkSMSMode: "1", // Enable bulk SMS mode
	}

	_, err := a.SendSMS(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send promotional SMS: %w", err)
	}

	return nil
}
