// Package africastalking provides Africa's Talking SMS and OTP integration
package africastalking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

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

// SendSMS sends an SMS message
func (c *AfricaTalkingClient) SendSMS(ctx context.Context, req *SMSRequest) (*SMSResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build request data
	data := url.Values{}
	data.Set("username", c.config.Username)
	data.Set("to", strings.Join(req.To, ","))
	data.Set("message", req.Message)

	if req.From != "" {
		data.Set("from", req.From)
	}

	// Make request
	resp, err := c.makeRequest(ctx, "POST", "/messaging", data)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	var smsResp SMSResponse
	if err := json.Unmarshal(resp, &smsResp); err != nil {
		return nil, fmt.Errorf("failed to parse SMS response: %w", err)
	}

	return &smsResp, nil
}

// SendOTP sends an OTP code
func (c *AfricaTalkingClient) SendOTP(ctx context.Context, req *OTPRequest) (*OTPResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build request data
	data := url.Values{}
	data.Set("username", c.config.Username)
	data.Set("phoneNumber", req.PhoneNumber)
	data.Set("length", fmt.Sprintf("%d", req.Length))
	data.Set("validity", fmt.Sprintf("%d", req.Validity))

	// Make request
	resp, err := c.makeRequest(ctx, "POST", "/otp/send", data)
	if err != nil {
		return nil, fmt.Errorf("failed to send OTP: %w", err)
	}

	var otpResp OTPResponse
	if err := json.Unmarshal(resp, &otpResp); err != nil {
		return nil, fmt.Errorf("failed to parse OTP response: %w", err)
	}

	return &otpResp, nil
}

// VerifyOTP verifies an OTP code
func (c *AfricaTalkingClient) VerifyOTP(ctx context.Context, req *OTPVerifyRequest) (*OTPVerifyResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build request data
	data := url.Values{}
	data.Set("username", c.config.Username)
	data.Set("phoneNumber", req.PhoneNumber)
	data.Set("otpCode", req.OTPCode)

	// Make request
	resp, err := c.makeRequest(ctx, "POST", "/otp/verify", data)
	if err != nil {
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	var verifyResp OTPVerifyResponse
	if err := json.Unmarshal(resp, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse OTP verification response: %w", err)
	}

	return &verifyResp, nil
}

// SendUSSD sends a USSD request
func (c *AfricaTalkingClient) SendUSSD(ctx context.Context, req *USSDRequest) (*USSDResponse, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build request data
	data := url.Values{}
	data.Set("username", c.config.Username)
	data.Set("phoneNumber", req.PhoneNumber)
	data.Set("text", req.Text)
	data.Set("sessionId", req.SessionID)

	// Make request
	resp, err := c.makeRequest(ctx, "POST", "/ussd", data)
	if err != nil {
		return nil, fmt.Errorf("failed to send USSD: %w", err)
	}

	var ussdResp USSDResponse
	if err := json.Unmarshal(resp, &ussdResp); err != nil {
		return nil, fmt.Errorf("failed to parse USSD response: %w", err)
	}

	return &ussdResp, nil
}

// makeRequest makes an HTTP request to Africa's Talking API
func (c *AfricaTalkingClient) makeRequest(ctx context.Context, method, endpoint string, data url.Values) ([]byte, error) {
	var reqBody *bytes.Buffer
	var contentType string

	if data != nil {
		reqBody = bytes.NewBufferString(data.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	url := c.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetMetrics returns client metrics
func (c *AfricaTalkingClient) GetMetrics(ctx context.Context) (*AfricaTalkingMetrics, error) {
	// In a real implementation, these would be calculated from actual data
	return &AfricaTalkingMetrics{
		TotalSMS:        10000,
		SuccessfulSMS:   9500,
		FailedSMS:       500,
		TotalOTP:        5000,
		SuccessfulOTP:   4800,
		FailedOTP:       200,
		TotalVoice:      1000,
		SuccessfulVoice: 980,
		FailedVoice:     20,
		LastActivity:    time.Now().Add(-1 * time.Hour),
	}, nil
}
