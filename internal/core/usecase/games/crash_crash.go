package games

import (
	"context"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// crashGame handles game crash
func (e *CrashGameEngine) crashGame(ctx context.Context, crashOdds decimal.Decimal) {
	e.mu.Lock()
	currentGame := e.currentGame
	if currentGame != nil {
		currentGame.Status = domain.GameStatusCrashed
		now := time.Now()
		currentGame.CrashedAt = &now
		e.currentGame = currentGame
	}
	e.mu.Unlock()

	if currentGame == nil {
		return
	}

	if err := e.gameRepo.UpdateStatus(ctx, currentGame.ID, domain.GameStatusCrashed); err != nil {
		log.Printf("Failed to update game status: %v", err)
	}

	gameState := &GameState{
		GameID:    currentGame.ID,
		Status:    domain.GameStatusCrashed,
		CrashOdds: crashOdds,
		IsCrashed: true,
	}
	e.hub.BroadcastGameState(gameState)

	e.mu.Lock()
	e.currentGame = nil
	e.mu.Unlock()
}

// stopGameInternal handles game stopping in the manager goroutine
func (e *CrashGameEngine) stopGameInternal() {
	if e.gameCancel != nil {
		e.gameCancel()
		e.gameCancel = nil
	}

	e.mu.Lock()
	e.currentGame = nil
	e.mu.Unlock()
}
