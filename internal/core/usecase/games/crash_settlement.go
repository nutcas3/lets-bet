package games

import (
	"context"
	"log"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// checkAutoCashouts checks for automatic cashouts
func (e *CrashGameEngine) checkAutoCashouts(ctx context.Context, currentOdds decimal.Decimal) {
	e.mu.RLock()
	currentGame := e.currentGame
	if currentGame == nil {
		e.mu.RUnlock()
		return
	}
	// Copy ID while holding the lock
	gameID := currentGame.ID
	e.mu.RUnlock()

	// Get active bets
	bets, err := e.betRepo.GetActiveByGame(ctx, gameID)
	if err != nil {
		log.Printf("Failed to get active bets: %v", err)
		return
	}

	for _, bet := range bets {
		if bet.CashoutAt != nil && currentOdds.GreaterThanOrEqual(*bet.CashoutAt) {
			// Auto cashout
			payout := bet.Amount.Mul(*bet.CashoutAt)
			payoutBreakdown, err := e.taxEngine.ApplyPayoutTax("KE", payout, bet.Amount)
			if err != nil {
				log.Printf("failed to calculate payout tax for bet %s: %v", bet.ID, err)
				continue
			}
			netPayout := payoutBreakdown.NetPayout

			// Use atomic transaction to prevent fund loss bug
			success, err := e.betRepo.AtomicAutoCashoutWithCredit(ctx, bet.ID, bet.UserID, *bet.CashoutAt, netPayout, "KE")
			if err != nil {
				log.Printf("Failed to process auto cashout for bet %s: %v", bet.ID, err)
				continue
			}
			if !success {
				log.Printf("Auto cashout skipped for bet %s (already processed)", bet.ID)
				continue
			}
		}
	}
}

// settleBets settles remaining active bets after crash
func (e *CrashGameEngine) settleBets(ctx context.Context) {
	e.mu.RLock()
	currentGame := e.currentGame
	if currentGame == nil {
		e.mu.RUnlock()
		return
	}
	// Copy ID while holding the lock
	gameID := currentGame.ID
	e.mu.RUnlock()

	// Get active bets
	bets, err := e.betRepo.GetActiveByGame(ctx, gameID)
	if err != nil {
		log.Printf("Failed to get active bets: %v", err)
		return
	}

	for _, bet := range bets {
		// Mark as lost (no payout)
		if err := e.betRepo.UpdateStatus(ctx, bet.ID, domain.GameBetStatusLost); err != nil {
			log.Printf("Failed to update bet status: %v", err)
		}
	}
}
