package kyc

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/config"
	smileid "github.com/nutcas3/smileid-go"
)

// ErrKYCNotConfigured is returned when Smile ID credentials are missing.
var ErrKYCNotConfigured = errors.New("smile id is not configured")

// Provider is the KYC provider contract used by the user/wallet services.
type Provider interface {
	VerifyUser(ctx context.Context, user *domain.User, idType, idNumber string) (*VerifyResult, error)
	VerifyID(ctx context.Context, countryCode, idType, idNumber string) (*VerifyResult, error)
}

// VerifyResult is the normalized KYC verification result surfaced to callers.
type VerifyResult struct {
	Verified     bool   `json:"verified"`
	Status       string `json:"status"`
	FullName     string `json:"full_name,omitempty"`
	IDNumber     string `json:"id_number,omitempty"`
	IDType       string `json:"id_type,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// SmileIDProvider wraps the Smile ID SDK with our domain-shaped API.
type SmileIDProvider struct {
	client *smileid.Client
}

// NewSmileIDProvider constructs a Smile ID provider from application config.
func NewSmileIDProvider(cfg config.SmileIDConfig) (*SmileIDProvider, error) {
	if cfg.APIKey == "" || cfg.PartnerID == "" {
		return nil, ErrKYCNotConfigured
	}
	env := strings.ToLower(cfg.Env)
	if env == "" {
		env = "sandbox"
	}
	client := smileid.NewClient(smileid.Config{
		APIKey:    cfg.APIKey,
		PartnerID: cfg.PartnerID,
		Env:       env,
		Timeout:   15 * time.Second,
	})
	return &SmileIDProvider{client: client}, nil
}

// VerifyUser runs a digital KYC check for a registered user.
func (p *SmileIDProvider) VerifyUser(ctx context.Context, user *domain.User, idType, idNumber string) (*VerifyResult, error) {
	req := smileid.KYCRequest{
		CountryCode: user.CountryCode,
		IDType:      idType,
		IDNumber:    idNumber,
		FirstName:   firstName(user.FullName),
		LastName:    lastName(user.FullName),
	}
	resp, err := p.client.KYC.VerifyUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return &VerifyResult{
		Verified:     resp.Verified,
		Status:       statusFor(resp.Verified),
		FullName:     resp.FullName,
		IDNumber:     resp.IDNumber,
		IDType:       resp.IDType,
		CountryCode:  resp.CountryCode,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

// VerifyID runs an identity verification check for a national ID or passport.
func (p *SmileIDProvider) VerifyID(ctx context.Context, countryCode, idType, idNumber string) (*VerifyResult, error) {
	req := smileid.VerificationRequest{
		CountryCode: countryCode,
		IDType:      idType,
		IDNumber:    idNumber,
	}
	resp, err := p.client.Identity.VerifyID(ctx, req)
	if err != nil {
		return nil, err
	}
	return &VerifyResult{
		Verified:     resp.Verified,
		Status:       statusFor(resp.Verified),
		IDNumber:     resp.IDNumber,
		IDType:       resp.IDType,
		CountryCode:  resp.CountryCode,
		ErrorMessage: resp.ErrorMessage,
	}, nil
}

func statusFor(verified bool) string {
	if verified {
		return "verified"
	}
	return "failed"
}

func firstName(full string) string {
	if i := strings.IndexByte(full, ' '); i > 0 {
		return full[:i]
	}
	return full
}

func lastName(full string) string {
	if i := strings.LastIndexByte(full, ' '); i > 0 {
		return full[i+1:]
	}
	return ""
}
