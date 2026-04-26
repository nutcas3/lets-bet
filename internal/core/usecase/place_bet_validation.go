package usecase

import (
	"errors"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// PlaceBet errors.
var (
	ErrUserNotEligible = errors.New("user not eligible to bet")
	ErrInvalidBet      = errors.New("invalid bet configuration")
	ErrStakeTooLow     = errors.New("stake below minimum")
	ErrStakeTooHigh    = errors.New("stake exceeds maximum")
)

// PlaceBetValidator handles validation logic for bet placement
type PlaceBetValidator struct {
	minStake decimal.Decimal
	maxStake decimal.Decimal
}

// NewPlaceBetValidator creates a new validator with configured limits
func NewPlaceBetValidator(minStake, maxStake decimal.Decimal) *PlaceBetValidator {
	return &PlaceBetValidator{
		minStake: minStake,
		maxStake: maxStake,
	}
}

// Validate checks if a bet placement request is valid
func (v *PlaceBetValidator) Validate(in PlaceBetInput) error {
	if in.Stake.LessThan(v.minStake) {
		return ErrStakeTooLow
	}
	if in.Stake.GreaterThan(v.maxStake) {
		return ErrStakeTooHigh
	}
	if len(in.Selections) == 0 {
		return ErrInvalidBet
	}
	if in.BetType == domain.BetTypeSingle && len(in.Selections) != 1 {
		return ErrInvalidBet
	}
	return nil
}

// ValidateUserEligibility checks if a user is eligible to place bets
func ValidateUserEligibility(user *domain.User) error {
	if !user.CanPlaceBet() {
		return ErrUserNotEligible
	}
	return nil
}
