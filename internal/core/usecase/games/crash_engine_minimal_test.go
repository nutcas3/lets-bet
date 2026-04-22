package games

import (
	"context"
	"testing"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// TestCrashEngineMinimal tests basic functionality without extreme concurrency
func TestCrashEngineMinimal(t *testing.T) {
	t.Parallel()

	// Create engine with minimal setup
	engine := &CrashGameEngine{
		roundNumber:  0,
		tickInterval: 10 * time.Millisecond,
	}

	// Test initial state
	gameState := engine.getGameState()
	assert.Equal(t, domain.GameStatusWaiting, gameState.Status)

	odds := engine.getCurrentOdds()
	assert.Equal(t, decimal.NewFromFloat(1.0), odds)

	// Test with no game - should not panic
	engine.crashGame(context.Background(), decimal.NewFromFloat(2.0))
	gameState = engine.getGameState()
	assert.Equal(t, domain.GameStatusWaiting, gameState.Status)
}

// TestCrashEngineBasicGame tests basic game operations
func TestCrashEngineBasicGame(t *testing.T) {
	t.Parallel()

	// Create engine with minimal setup
	engine := &CrashGameEngine{
		roundNumber:  42,
		tickInterval: 10 * time.Millisecond,
	}

	game := &domain.Game{
		ID:         uuid.New(),
		GameType:   domain.GameTypeCrash,
		RoundNumber: 42,
		Status:     domain.GameStatusRunning,
		StartedAt:  time.Now(),
		CrashPoint: decimal.NewFromFloat(2.5),
	}

	// Set current game
	engine.mu.Lock()
	engine.currentGame = game
	engine.mu.Unlock()

	// Test game state
	gameState := engine.getGameState()
	assert.Equal(t, game.ID, gameState.GameID)
	assert.Equal(t, domain.GameStatusRunning, gameState.Status)
	assert.Equal(t, int64(42), gameState.RoundNumber)

	// Test odds calculation
	odds := engine.getCurrentOdds()
	assert.True(t, odds.GreaterThanOrEqual(decimal.NewFromFloat(1.0)))

	// Test crash game
	ctx := context.Background()
	engine.crashGame(ctx, decimal.NewFromFloat(2.0))

	// Verify game is cleared
	gameState = engine.getGameState()
	assert.Equal(t, domain.GameStatusWaiting, gameState.Status)
}
