package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a betting platform user
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	PhoneNumber  string    `json:"phone_number" db:"phone_number"` // Primary identifier in Kenya
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CountryCode  string    `json:"country_code" db:"country_code"` // KE, NG, GH
	Currency     string    `json:"currency" db:"currency"`         // KES, NGN, GHS

	// KYC fields (BCLB requirement)
	NationalID  string    `json:"national_id,omitempty" db:"national_id"`
	KRAPin      string    `json:"kra_pin,omitempty" db:"kra_pin"`
	FullName    string    `json:"full_name" db:"full_name"`
	DateOfBirth time.Time `json:"date_of_birth" db:"date_of_birth"`
	IsVerified  bool      `json:"is_verified" db:"is_verified"`

	// Responsible Gaming (legal requirement)
	SelfExcluded      bool       `json:"self_excluded" db:"self_excluded"`
	SelfExcludedUntil *time.Time `json:"self_excluded_until,omitempty" db:"self_excluded_until"`
	DailyDepositLimit *int64     `json:"daily_deposit_limit,omitempty" db:"daily_deposit_limit"` // in cents

	// Metadata
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	Status      UserStatus `json:"status" db:"status"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
	UserStatusBanned    UserStatus = "BANNED"
	UserStatusPending   UserStatus = "PENDING_VERIFICATION"
)

// CanPlaceBet checks if user meets all requirements to place a bet
func (u *User) CanPlaceBet() bool {
	if u.Status != UserStatusActive {
		return false
	}
	if u.SelfExcluded {
		if u.SelfExcludedUntil != nil && time.Now().Before(*u.SelfExcludedUntil) {
			return false
		}
	}
	if !u.IsVerified {
		return false // BCLB requires KYC before betting
	}
	return true
}
