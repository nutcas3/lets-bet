package games

import (
	"context"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase"
	"github.com/betting-platform/internal/core/usecase/tax"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func NewCrashGameEngine(
	hub WebSocketHub,
	fairService *usecase.ProvablyFairService,
	gameRepo GameRepository,
	betRepo GameBetRepository,
	walletService *wallet.Service,
	taxEngine *tax.Engine,
) *CrashGameEngine {
	return &CrashGameEngine{
		hub:           hub,
		fairService:   fairService,
		gameRepo:      gameRepo,
		betRepo:       betRepo,
		walletService: walletService,
		taxEngine:     taxEngine,
		roundNumber:   0,
		tickInterval:  100 * time.Millisecond,
		betChan:       make(chan BetRequest, 100),
		cashoutChan:   make(chan CashoutRequest, 100),
		commandChan:   make(chan string, 10),
	}
}

// Run starts the manager goroutine that handles all state changes
func (e *CrashGameEngine) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-e.betChan:
			e.handlePlaceBet(req)
		case req := <-e.cashoutChan:
			e.handleCashout(req)
		case cmd := <-e.commandChan:
			e.handleCommand(cmd)
		case <-time.After(e.tickInterval):
			e.tick()
		}
	}
}

// StartGame starts a new crash game using channel-based approach
func (e *CrashGameEngine) StartGame(ctx context.Context) error {
	e.mu.RLock()
	currentGame := e.currentGame
	e.mu.RUnlock()

	if currentGame != nil && currentGame.Status == domain.GameStatusRunning {
		return fmt.Errorf("game already in progress")
	}

	// Send start command to manager goroutine
	select {
	case e.commandChan <- "START":
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout starting game")
	}
}

// PlaceBet places a bet on the current game using channel-based approach
func (e *CrashGameEngine) PlaceBet(ctx context.Context, req *BetRequest) (*BetResponse, error) {
	// Validate bet amount
	if req.Amount.LessThan(decimal.NewFromFloat(10)) || req.Amount.GreaterThan(decimal.NewFromFloat(10000)) {
		return &BetResponse{
			Success: false,
			Message: "Bet amount must be between 10 and 10000",
		}, nil
	}

	// Create response channel
	respChan := make(chan error, 1)
	req.Resp = respChan

	// Send bet request to manager goroutine
	select {
	case e.betChan <- *req:
		// Wait for response
		select {
		case err := <-respChan:
			if err != nil {
				return &BetResponse{
					Success: false,
					Message: err.Error(),
				}, nil
			}
			return &BetResponse{
				Success:   true,
				Message:   "Bet placed successfully",
				GameState: e.getGameState(),
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return &BetResponse{
				Success: false,
				Message: "Timeout processing bet",
			}, nil
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return &BetResponse{
			Success: false,
			Message: "Timeout submitting bet",
		}, nil
	}
}

// Cashout cashes out a bet at the current odds using channel-based approach
func (e *CrashGameEngine) Cashout(ctx context.Context, req *CashoutRequest) (*CashoutResponse, error) {
	// Create response channel
	respChan := make(chan *CashoutResponse, 1)
	req.Resp = respChan

	// Send cashout request to manager goroutine
	select {
	case e.cashoutChan <- *req:
		// Wait for response
		select {
		case resp := <-respChan:
			return resp, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return &CashoutResponse{
				Success: false,
				Message: "Timeout processing cashout",
			}, nil
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
		return &CashoutResponse{
			Success: false,
			Message: "Timeout submitting cashout",
		}, nil
	}
}

// GetGameState returns the current game state
func (e *CrashGameEngine) GetGameState() *GameState {
	return e.getGameState()
}

// GetGameHistory returns game history
func (e *CrashGameEngine) GetGameHistory(ctx context.Context, limit int) ([]*GameHistory, error) {
	// This would typically query the repository for game history
	// For now, return empty slice
	return []*GameHistory{}, nil
}

// GetPlayerStats returns player statistics
func (e *CrashGameEngine) GetPlayerStats(ctx context.Context, userID uuid.UUID) (*PlayerStats, error) {
	// This would typically query the repository for player stats
	// For now, return empty stats
	return &PlayerStats{
		UserID: userID,
	}, nil
}
