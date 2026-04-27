package edit

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/shopspring/decimal"
)

// WalletService interface for wallet operations
type WalletService interface {
	Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
	Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
}

// NewEditBetService creates a new edit bet service
func NewEditBetService(
	betRepo *postgres.SportBetRepository,
	matchRepo *postgres.MatchRepository,
	marketRepo *postgres.BettingMarketRepository,
	outcomeRepo *postgres.MarketOutcomeRepository,
	walletService WalletService,
	eventBus EventBus,
) *EditBetService {
	idGenerator, err := id.ServiceTypeGenerator("edit")
	if err != nil {
		panic(fmt.Sprintf("Failed to create edit ID generator: %v", err))
	}

	return &EditBetService{
		betRepo:       betRepo,
		matchRepo:     matchRepo,
		marketRepo:    marketRepo,
		outcomeRepo:   outcomeRepo,
		walletService: walletService,
		eventBus:      eventBus,
		idGenerator:   idGenerator,
	}
}

// EditBet edits an existing bet
func (s *EditBetService) EditBet(ctx context.Context, req *EditBetRequest) (*EditBetResponse, error) {
	// Get the original bet
	originalBet, err := s.betRepo.GetByID(ctx, req.BetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bet: %w", err)
	}

	// Validate ownership
	if originalBet.UserID != req.UserID {
		return nil, fmt.Errorf("unauthorized: bet does not belong to user")
	}

	// Check if bet can be edited
	if !s.canEditBet(originalBet) {
		return nil, fmt.Errorf("bet cannot be edited: invalid status or match has started")
	}

	// Get the match to check timing
	match, err := s.matchRepo.GetByID(ctx, originalBet.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	// Check if match has started
	if time.Now().After(match.StartTime) {
		return nil, fmt.Errorf("cannot edit bet: match has already started")
	}

	// Calculate refund or additional amount
	refundAmount, additionalAmount := s.calculateAmountDifference(originalBet, req)

	// Process refund if needed
	if refundAmount.GreaterThan(decimal.Zero) {
		err = s.processRefund(ctx, originalBet, refundAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to process refund: %w", err)
		}
	}

	// Process additional payment if needed
	if additionalAmount.GreaterThan(decimal.Zero) {
		err = s.processAdditionalPayment(ctx, originalBet, additionalAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to process additional payment: %w", err)
		}
	}

	// Create edited bet
	editedBet := s.createEditedBet(originalBet, req)

	// Save the edited bet
	err = s.betRepo.Update(ctx, editedBet)
	if err != nil {
		return nil, fmt.Errorf("failed to update bet: %w", err)
	}

	// Publish edit event
	s.publishEditEvent(originalBet, editedBet, req.Reason)

	// Create response
	response := &EditBetResponse{
		OriginalBet:      originalBet,
		EditedBet:        editedBet,
		RefundAmount:     refundAmount,
		AdditionalAmount: additionalAmount,
		EditedAt:         time.Now(),
		Reason:           req.Reason,
	}

	return response, nil
}

// canEditBet checks if a bet can be edited
func (s *EditBetService) canEditBet(bet *domain.SportBet) bool {
	// Can only edit pending bets
	if bet.Status != domain.BetStatusPending {
		return false
	}

	// Can only edit bets placed within the last 5 minutes
	editWindow := 5 * time.Minute
	if time.Since(bet.PlacedAt) > editWindow {
		return false
	}

	return true
}

// calculateAmountDifference calculates the refund or additional amount needed
func (s *EditBetService) calculateAmountDifference(originalBet *domain.SportBet, req *EditBetRequest) (decimal.Decimal, decimal.Decimal) {
	originalStake := originalBet.Amount
	newStake := req.NewAmount

	if newStake.LessThan(originalStake) {
		// Refund the difference
		refundAmount := originalStake.Sub(newStake)
		return refundAmount, decimal.Zero
	} else if newStake.GreaterThan(originalStake) {
		// Charge additional amount
		additionalAmount := newStake.Sub(originalStake)
		return decimal.Zero, additionalAmount
	}

	// No change in amount
	return decimal.Zero, decimal.Zero
}

// processRefund processes a refund for bet reduction
func (s *EditBetService) processRefund(ctx context.Context, bet *domain.SportBet, refundAmount decimal.Decimal) error {
	movement := Movement{
		UserID:        bet.UserID,
		Amount:        refundAmount,
		Type:          domain.TransactionTypeBetRefund,
		ReferenceID:   &bet.ID,
		ReferenceType: "sport_bet",
		Description:   fmt.Sprintf("Bet edit refund for bet %s", bet.ID),
		ProviderName:  "edit_bet",
		ProviderTxnID: fmt.Sprintf("refund-%s", s.idGenerator.GenerateID()),
		CountryCode:   bet.CountryCode,
	}

	_, err := s.walletService.Credit(ctx, bet.UserID, refundAmount, movement)
	if err != nil {
		return fmt.Errorf("failed to credit wallet: %w", err)
	}

	log.Printf("Refunded %s to user %s for bet edit %s", refundAmount.String(), bet.UserID, bet.ID)
	return nil
}

// processAdditionalPayment processes additional payment for bet increase
func (s *EditBetService) processAdditionalPayment(ctx context.Context, bet *domain.SportBet, additionalAmount decimal.Decimal) error {
	movement := Movement{
		UserID:        bet.UserID,
		Amount:        additionalAmount,
		Type:          domain.TransactionTypeBetPlaced,
		ReferenceID:   &bet.ID,
		ReferenceType: "sport_bet",
		Description:   fmt.Sprintf("Bet edit additional payment for bet %s", bet.ID),
		ProviderName:  "edit_bet",
		ProviderTxnID: fmt.Sprintf("additional-%s", s.idGenerator.GenerateID()),
		CountryCode:   bet.CountryCode,
	}

	_, err := s.walletService.Debit(ctx, bet.UserID, additionalAmount, movement)
	if err != nil {
		return fmt.Errorf("failed to debit wallet: %w", err)
	}

	log.Printf("Charged additional %s from user %s for bet edit %s", additionalAmount.String(), bet.UserID, bet.ID)
	return nil
}

// createEditedBet creates the edited version of the bet
func (s *EditBetService) createEditedBet(originalBet *domain.SportBet, req *EditBetRequest) *domain.SportBet {
	editedBet := *originalBet // Copy the original bet

	// Update the editable fields
	editedBet.Amount = req.NewAmount
	editedBet.Odds = req.NewOdds
	editedBet.UpdatedAt = time.Now()

	// Update outcome if specified
	if req.NewOutcomeID != "" {
		editedBet.OutcomeID = req.NewOutcomeID
	}

	return &editedBet
}

// publishEditEvent publishes a bet edit event
func (s *EditBetService) publishEditEvent(originalBet, editedBet *domain.SportBet, reason string) {
	event := map[string]any{
		"event_type":      "bet_edited",
		"bet_id":          editedBet.ID,
		"user_id":         editedBet.UserID,
		"original_amount": originalBet.Amount,
		"new_amount":      editedBet.Amount,
		"original_odds":   originalBet.Odds,
		"new_odds":        editedBet.Odds,
		"reason":          reason,
		"edited_at":       time.Now(),
	}

	err := s.eventBus.Publish("bet.edited", event)
	if err != nil {
		log.Printf("Failed to publish bet edit event: %v", err)
	}
}

// GetEditableBets returns bets that can be edited for a user
func (s *EditBetService) GetEditableBets(ctx context.Context, userID string) ([]*domain.SportBet, error) {
	// Get recent pending bets for the user
	status := domain.BetStatusPending
	filters := &postgres.BetFilters{Status: &status}
	bets, err := s.betRepo.GetByUserID(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bets: %w", err)
	}

	// Filter bets that can be edited
	var editableBets []*domain.SportBet
	for _, bet := range bets {
		if s.canEditBet(bet) {
			// Check if match hasn't started
			match, err := s.matchRepo.GetByID(ctx, bet.EventID)
			if err != nil {
				log.Printf("Failed to get match for bet %s: %v", bet.ID, err)
				continue
			}

			if time.Now().Before(match.StartTime) {
				editableBets = append(editableBets, bet)
			}
		}
	}

	return editableBets, nil
}
