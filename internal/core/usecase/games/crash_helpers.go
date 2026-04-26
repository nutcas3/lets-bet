package games

import (
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// getGameState creates the current game state using atomic.Value for lock-free reads
func (e *CrashGameEngine) getGameState() *GameState {
	if state := e.latestState.Load(); state != nil {
		return state.(*GameState)
	}

	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil {
		return &GameState{Status: domain.GameStatusWaiting}
	}

	activePlayerCount := e.hub.GetActivePlayerCount(game.ID)

	return &GameState{
		GameID:        game.ID,
		RoundNumber:   e.roundNumber,
		Status:        game.Status,
		StartedAt:     game.StartedAt,
		CurrentOdds:   e.calculateCurrentOdds(game.StartedAt),
		ActivePlayers: activePlayerCount,
	}
}

// calculateCurrentOdds calculates the current odds based on elapsed time
func (e *CrashGameEngine) calculateCurrentOdds(startedAt time.Time) decimal.Decimal {
	elapsed := time.Since(startedAt).Milliseconds()
	seconds := decimal.NewFromInt(elapsed).Div(decimal.NewFromInt(1000))
	return decimal.NewFromFloat(1.0).Add(seconds.Mul(decimal.NewFromFloat(0.1)))
}

// getCurrentOdds returns the current odds for the active game
func (e *CrashGameEngine) getCurrentOdds() decimal.Decimal {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil {
		return decimal.NewFromFloat(1.0)
	}

	return e.calculateCurrentOdds(game.StartedAt)
}
