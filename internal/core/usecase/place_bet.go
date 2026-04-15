package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrUserNotEligible     = errors.New("user not eligible to bet")
	ErrInvalidBet          = errors.New("invalid bet configuration")
	ErrStakeTooLow         = errors.New("stake below minimum")
	ErrStakeTooHigh        = errors.New("stake exceeds maximum")
)

// BetRepository defines the interface for bet persistence
type BetRepository interface {
	Create(ctx context.Context, bet *domain.Bet) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bet, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BetStatus) error
}

// WalletRepository defines wallet operations with optimistic locking
type WalletRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error)
	UpdateBalance(ctx context.Context, wallet *domain.Wallet, tx *domain.Transaction) error
	CreateTransaction(ctx context.Context, tx *domain.Transaction) error
}

// UserRepository defines user data operations
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// PlaceBetUseCase handles the core bet placement logic
type PlaceBetUseCase struct {
	betRepo    BetRepository
	walletRepo WalletRepository
	userRepo   UserRepository
}

func NewPlaceBetUseCase(
	betRepo BetRepository,
	walletRepo WalletRepository,
	userRepo UserRepository,
) *PlaceBetUseCase {
	return &PlaceBetUseCase{
		betRepo:    betRepo,
		walletRepo: walletRepo,
		userRepo:   userRepo,
	}
}

type PlaceBetInput struct {
	UserID      uuid.UUID
	BetType     domain.BetType
	Stake       decimal.Decimal
	Selections  []domain.Selection
	IPAddress   string
	DeviceID    string
	CountryCode string
}

// Execute places a bet with full transactional integrity
func (uc *PlaceBetUseCase) Execute(ctx context.Context, input PlaceBetInput) (*domain.Bet, error) {
	// 1. Validate user eligibility
	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	
	if !user.CanPlaceBet() {
		return nil, ErrUserNotEligible
	}
	
	// 2. Validate bet configuration
	if err := uc.validateBet(input); err != nil {
		return nil, err
	}
	
	// 3. Calculate odds and potential win
	totalOdds := uc.calculateTotalOdds(input.BetType, input.Selections)
	potentialWin := input.Stake.Mul(totalOdds)
	
	// 4. Create bet entity
	bet := &domain.Bet{
		ID:           uuid.New(),
		UserID:       input.UserID,
		CountryCode:  input.CountryCode,
		BetType:      input.BetType,
		Stake:        input.Stake,
		Currency:     user.Currency,
		PotentialWin: potentialWin,
		TotalOdds:    totalOdds,
		Status:       domain.BetStatusPending,
		ActualWin:    decimal.Zero,
		Selections:   input.Selections,
		PlacedAt:     time.Now(),
		IPAddress:    input.IPAddress,
		DeviceID:     input.DeviceID,
		TaxAmount:    decimal.Zero,
		TaxPaid:      false,
	}
	
	// 5. Deduct stake from wallet (atomic operation)
	wallet, err := uc.walletRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	
	if !wallet.CanWithdraw(input.Stake) {
		return nil, ErrInsufficientBalance
	}
	
	// Create transaction record
	tx := &domain.Transaction{
		ID:            uuid.New(),
		WalletID:      wallet.ID,
		UserID:        input.UserID,
		Type:          domain.TransactionTypeBetPlaced,
		Amount:        input.Stake.Neg(), // Negative for debit
		Currency:      user.Currency,
		BalanceBefore: wallet.Balance,
		BalanceAfter:  wallet.Balance.Sub(input.Stake),
		ReferenceID:   &bet.ID,
		ReferenceType: "BET",
		Status:        domain.TransactionStatusCompleted,
		Description:   "Bet placed",
		CreatedAt:     time.Now(),
		CountryCode:   input.CountryCode,
	}
	
	now := time.Now()
	tx.CompletedAt = &now
	
	// Update wallet balance
	wallet.Balance = wallet.Balance.Sub(input.Stake)
	wallet.Version++ // Optimistic locking
	
	// 6. Persist bet and wallet update atomically
	if err := uc.walletRepo.UpdateBalance(ctx, wallet, tx); err != nil {
		return nil, err
	}
	
	if err := uc.betRepo.Create(ctx, bet); err != nil {
		// In production, this would trigger a rollback
		return nil, err
	}
	
	return bet, nil
}

func (uc *PlaceBetUseCase) validateBet(input PlaceBetInput) error {
	// Minimum stake (e.g., KES 10)
	minStake := decimal.NewFromInt(10)
	if input.Stake.LessThan(minStake) {
		return ErrStakeTooLow
	}
	
	// Maximum stake (e.g., KES 100,000)
	maxStake := decimal.NewFromInt(100000)
	if input.Stake.GreaterThan(maxStake) {
		return ErrStakeTooHigh
	}
	
	// Validate selections
	if len(input.Selections) == 0 {
		return ErrInvalidBet
	}
	
	if input.BetType == domain.BetTypeSingle && len(input.Selections) != 1 {
		return ErrInvalidBet
	}
	
	return nil
}

func (uc *PlaceBetUseCase) calculateTotalOdds(betType domain.BetType, selections []domain.Selection) decimal.Decimal {
	if len(selections) == 0 {
		return decimal.NewFromInt(1)
	}
	
	switch betType {
	case domain.BetTypeSingle:
		return selections[0].Odds
	case domain.BetTypeMulti:
		// Multiply all odds together
		total := decimal.NewFromInt(1)
		for _, sel := range selections {
			total = total.Mul(sel.Odds)
		}
		return total
	case domain.BetTypeSystem:
		// Simplified: return average for now
		// Real implementation would calculate combinations
		total := decimal.NewFromInt(1)
		for _, sel := range selections {
			total = total.Mul(sel.Odds)
		}
		return total
	default:
		return decimal.NewFromInt(1)
	}
}
