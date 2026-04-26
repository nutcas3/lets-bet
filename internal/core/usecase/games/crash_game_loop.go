package games

import (
	"context"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// tick updates game state and broadcasts
func (e *CrashGameEngine) tick() {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil || game.Status != domain.GameStatusRunning {
		return
	}

	currentOdds := e.calculateCurrentOdds(game.StartedAt)
	maxOdds := game.CrashPoint.Div(decimal.NewFromFloat(100.0))

	if currentOdds.GreaterThanOrEqual(maxOdds) {
		e.commandChan <- "CRASH"
		return
	}

	e.checkAutoCashouts(context.Background(), currentOdds)

	e.mu.RLock()
	roundNumber := e.roundNumber
	e.mu.RUnlock()

	state := &GameState{
		GameID:        game.ID,
		RoundNumber:   roundNumber,
		Status:        game.Status,
		StartedAt:     game.StartedAt,
		CurrentOdds:   currentOdds,
		MaxOdds:       maxOdds,
		ActivePlayers: e.hub.GetActivePlayerCount(game.ID),
	}
	e.latestState.Store(state)
	e.hub.BroadcastGameState(state)
}

// startGameInternal handles game startup in the manager goroutine
func (e *CrashGameEngine) startGameInternal() {
	if e.gameCancel != nil {
		e.gameCancel()
	}

	e.mu.Lock()
	e.roundNumber++
	roundNumber := e.roundNumber
	e.mu.Unlock()

	game := e.createGame(roundNumber)
	if game == nil {
		return
	}

	e.mu.Lock()
	e.currentGame = game
	e.mu.Unlock()

	gameCtx, cancel := context.WithCancel(context.Background())
	e.gameCancel = cancel
	go e.gameLoop(gameCtx)

	state := &GameState{
		GameID:      game.ID,
		RoundNumber: roundNumber,
		Status:      domain.GameStatusRunning,
		StartedAt:   game.StartedAt,
		CurrentOdds: decimal.NewFromFloat(1.0),
		MaxOdds:     game.CrashPoint.Div(decimal.NewFromFloat(100.0)),
	}
	e.latestState.Store(state)
	e.hub.BroadcastGameState(state)
}

// createGame creates a new game with provably fair crash point
func (e *CrashGameEngine) createGame(roundNumber int64) *domain.Game {
	gameID := uuid.New()

	seed, err := e.fairService.GenerateServerSeed()
	if err != nil {
		log.Printf("Failed to generate server seed: %v", err)
		return nil
	}
	hash := e.fairService.HashServerSeed(seed)
	crashPoint := e.fairService.CalculateCrashPoint(seed, "", roundNumber)

	game := &domain.Game{
		ID:             gameID,
		GameType:       domain.GameTypeCrash,
		RoundNumber:    roundNumber,
		Status:         domain.GameStatusRunning,
		StartedAt:      time.Now(),
		ServerSeed:     seed,
		ServerSeedHash: hash,
		CrashPoint:     crashPoint,
	}

	if err := e.gameRepo.Create(context.Background(), game); err != nil {
		log.Printf("Failed to create game: %v", err)
		return nil
	}

	return game
}

// crashGameInternal handles game crash in the manager goroutine
func (e *CrashGameEngine) crashGameInternal() {
	e.mu.RLock()
	currentGame := e.currentGame
	e.mu.RUnlock()

	if currentGame == nil {
		return
	}

	currentOdds := e.calculateCurrentOdds(currentGame.StartedAt)
	e.crashGame(context.Background(), currentOdds)
	e.settleBets(context.Background())

	state := &GameState{
		Status: domain.GameStatusWaiting,
	}
	e.latestState.Store(state)
	e.hub.BroadcastGameState(state)

	if e.gameCancel != nil {
		e.gameCancel()
		e.gameCancel = nil
	}

	e.mu.Lock()
	e.currentGame = nil
	e.mu.Unlock()
}

// gameLoop runs the main game loop
func (e *CrashGameEngine) gameLoop(ctx context.Context) {
	e.mu.RLock()
	crashPoint := e.currentGame.CrashPoint
	e.mu.RUnlock()

	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick()
		}

		if e.shouldCrash(crashPoint) {
			e.crashGame(ctx, crashPoint)
			return
		}
	}
}

// shouldCrash checks if the game should crash based on current odds
func (e *CrashGameEngine) shouldCrash(crashPoint decimal.Decimal) bool {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil || game.Status != domain.GameStatusRunning {
		return true
	}

	return e.calculateCurrentOdds(game.StartedAt).GreaterThanOrEqual(crashPoint)
}
