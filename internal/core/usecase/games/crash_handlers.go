package games

import (
	"context"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
)

// handlePlaceBet processes bet requests in the manager goroutine
func (e *CrashGameEngine) handlePlaceBet(req BetRequest) {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil || game.Status != domain.GameStatusRunning {
		req.Resp <- fmt.Errorf("game not active")
		return
	}

	bet := &domain.GameBet{
		ID:       uuid.New(),
		GameID:   game.ID,
		UserID:   req.UserID,
		Amount:   req.Amount,
		Status:   domain.GameBetStatusActive,
		PlacedAt: time.Now(),
	}

	_, err := e.betRepo.CreateBetWithWalletUpdate(context.Background(), bet, req.UserID, req.Amount)
	if err != nil {
		req.Resp <- fmt.Errorf("failed to place bet: %w", err)
		return
	}

	req.Resp <- nil
}

// handleCashout processes cashout requests in the manager goroutine
func (e *CrashGameEngine) handleCashout(req CashoutRequest) {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "No active game",
		}
		return
	}

	currentOdds := e.getCurrentOdds()
	bet, err := e.betRepo.GetByID(context.Background(), req.BetID)
	if err != nil {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Bet not found",
		}
		return
	}

	payout := bet.Amount.Mul(currentOdds)

	success, err := e.betRepo.AtomicCashout(context.Background(), req.BetID, currentOdds, payout)
	if err != nil {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Database error during cashout",
		}
		return
	}

	if !success {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Bet already cashed out or not found",
		}
		return
	}

	req.Resp <- &CashoutResponse{
		Success:     true,
		Message:     "Cashout processed",
		CashoutOdds: currentOdds,
		Payout:      payout,
	}
}

// handleCommand processes command messages in the manager goroutine
func (e *CrashGameEngine) handleCommand(cmd string) {
	switch cmd {
	case "START":
		e.startGameInternal()
	case "CRASH":
		e.crashGameInternal()
	case "STOP":
		e.stopGameInternal()
	}
}
