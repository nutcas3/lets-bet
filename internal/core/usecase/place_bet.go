package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase/tax"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BetInserter persists a new bet inside the caller's transaction.
// It is implemented by the postgres BetRepository against an *sql.Tx.
type BetInserter interface {
	InsertTx(ctx context.Context, tx wallet.DBTX, bet *domain.Bet) error
}

// UserRepository defines user data operations.
type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// PlaceBetUseCase implements the PlaceBet state machine:
//
//	validate → reserve funds → persist → (later) settle.
//
// The reserve + persist steps run in a single DB transaction so a failure
// anywhere in the flow never debits a player without a matching bet row.
type PlaceBetUseCase struct {
	bets    BetInserter
	wallets *wallet.Service
	users   UserRepository
	tax     *tax.Engine

	minStake decimal.Decimal
	maxStake decimal.Decimal
}

// NewPlaceBetUseCase constructs the usecase with sensible defaults.
func NewPlaceBetUseCase(
	bets BetInserter,
	wallets *wallet.Service,
	users UserRepository,
	taxEngine *tax.Engine,
) *PlaceBetUseCase {
	return &PlaceBetUseCase{
		bets:     bets,
		wallets:  wallets,
		users:    users,
		tax:      taxEngine,
		minStake: decimal.NewFromInt(10),
		maxStake: decimal.NewFromInt(100000),
	}
}

// PlaceBetInput is the input contract for [PlaceBetUseCase.Execute].
type PlaceBetInput struct {
	UserID      uuid.UUID
	BetType     domain.BetType
	Stake       decimal.Decimal
	Selections  []domain.Selection
	IPAddress   string
	DeviceID    string
	CountryCode string
}

// Execute runs the bet placement state machine.
func (uc *PlaceBetUseCase) Execute(ctx context.Context, in PlaceBetInput) (*domain.Bet, error) {
	// 1. Validate user eligibility.
	user, err := uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	if !user.CanPlaceBet() {
		return nil, ErrUserNotEligible
	}

	// 2. Validate bet shape using validator
	validator := NewPlaceBetValidator(uc.minStake, uc.maxStake)
	if err := validator.Validate(in); err != nil {
		return nil, err
	}

	// 3. Apply stake tax (collected up front).
	stakeBreak, err := uc.tax.ApplyStakeTax(in.CountryCode, in.Stake)
	if err != nil {
		return nil, fmt.Errorf("apply stake tax: %w", err)
	}

	// 4. Compute odds and potential win from the post-tax net stake, so the
	// advertised odds apply to the actual money at risk.
	totalOdds := uc.calculateTotalOdds(in.BetType, in.Selections)
	potentialWin := stakeBreak.NetStake.Mul(totalOdds).Round(2)

	bet := &domain.Bet{
		ID:           uuid.New(),
		UserID:       in.UserID,
		CountryCode:  in.CountryCode,
		BetType:      in.BetType,
		Stake:        in.Stake,
		Currency:     user.Currency,
		PotentialWin: potentialWin,
		TotalOdds:    totalOdds,
		Status:       domain.BetStatusPending,
		ActualWin:    decimal.Zero,
		Selections:   in.Selections,
		PlacedAt:     time.Now().UTC(),
		IPAddress:    in.IPAddress,
		DeviceID:     in.DeviceID,
		TaxAmount:    stakeBreak.StakeTax,
		TaxPaid:      true,
	}

	// 5. Reserve funds + persist bet atomically.
	dbTx, err := uc.wallets.DB().BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = dbTx.Rollback() }()

	// Debit only the net stake (amount actually wagered)
	if _, err := uc.wallets.ApplyTx(ctx, dbTx, wallet.Movement{
		UserID:        in.UserID,
		Amount:        stakeBreak.NetStake.Neg(),
		Type:          domain.TransactionTypeBetPlaced,
		ReferenceID:   &bet.ID,
		ReferenceType: "BET",
		Description:   fmt.Sprintf("bet %s", bet.ID),
		CountryCode:   in.CountryCode,
	}); err != nil {
		return nil, fmt.Errorf("debit stake: %w", err)
	}

	// Record tax deduction as separate wallet movement for audit trail
	if stakeBreak.StakeTax.GreaterThan(decimal.Zero) {
		if _, err := uc.wallets.ApplyTx(ctx, dbTx, wallet.Movement{
			UserID:        in.UserID,
			Amount:        stakeBreak.StakeTax.Neg(),
			Type:          domain.TransactionTypeTax,
			ReferenceID:   &bet.ID,
			ReferenceType: "BET",
			Description:   fmt.Sprintf("stake tax for bet %s", bet.ID),
			CountryCode:   in.CountryCode,
		}); err != nil {
			return nil, err
		}
	}

	if err := uc.bets.InsertTx(ctx, dbTx, bet); err != nil {
		return nil, fmt.Errorf("insert bet: %w", err)
	}

	if err := dbTx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return bet, nil
}

func (uc *PlaceBetUseCase) calculateTotalOdds(betType domain.BetType, selections []domain.Selection) decimal.Decimal {
	switch betType {
	case domain.BetTypeSingle:
		if len(selections) == 0 {
			return decimal.NewFromInt(1)
		}
		return selections[0].Odds
	case domain.BetTypeMulti, domain.BetTypeSystem:
		total := decimal.NewFromInt(1)
		for _, sel := range selections {
			total = total.Mul(sel.Odds)
		}
		return total
	default:
		return decimal.NewFromInt(1)
	}
}
