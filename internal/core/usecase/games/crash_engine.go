package games

import (
	"context"
	"fmt"
	"log"
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
	}
}

// StartGame starts a new crash game
func (e *CrashGameEngine) StartGame(ctx context.Context) error {
	if e.currentGame != nil && e.currentGame.Status == domain.GameStatusRunning {
		return fmt.Errorf("game already in progress")
	}

	e.roundNumber++
	gameID := uuid.New()

	// Generate provably fair crash point
	seed, err := e.fairService.GenerateServerSeed()
	if err != nil {
		return fmt.Errorf("failed to generate server seed: %w", err)
	}
	hash := e.fairService.HashServerSeed(seed)
	crashPoint := e.fairService.CalculateCrashPoint(seed, "", e.roundNumber)

	// Create new game
	game := &domain.Game{
		ID:             gameID,
		GameType:       domain.GameTypeCrash,
		RoundNumber:    e.roundNumber,
		Status:         domain.GameStatusRunning,
		StartedAt:      time.Now(),
		ServerSeed:     seed,
		ServerSeedHash: hash,
		CrashPoint:     crashPoint,
	}

	if err := e.gameRepo.Create(ctx, game); err != nil {
		return fmt.Errorf("failed to create game: %w", err)
	}

	e.currentGame = game

	// Start game loop
	go e.gameLoop(ctx)

	return nil
}

// PlaceBet places a bet on the current game
func (e *CrashGameEngine) PlaceBet(ctx context.Context, req *BetRequest) (*BetResponse, error) {
	if e.currentGame == nil || e.currentGame.Status != domain.GameStatusRunning {
		return &BetResponse{
			Success: false,
			Message: "No active game to bet on",
		}, nil
	}

	// Validate bet amount
	if req.Amount.LessThan(decimal.NewFromFloat(10)) || req.Amount.GreaterThan(decimal.NewFromFloat(10000)) {
		return &BetResponse{
			Success: false,
			Message: "Bet amount must be between 10 and 10000",
		}, nil
	}

	// Check if user has sufficient balance
	balance, _, err := e.walletService.Balance(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	if balance.LessThan(req.Amount) {
		return &BetResponse{
			Success: false,
			Message: "Insufficient balance",
		}, nil
	}

	// Create bet
	bet := &domain.GameBet{
		ID:       uuid.New(),
		GameID:   e.currentGame.ID,
		UserID:   req.UserID,
		Amount:   req.Amount,
		Status:   domain.GameBetStatusActive,
		PlacedAt: time.Now(),
	}

	if err := e.betRepo.Create(ctx, bet); err != nil {
		return nil, fmt.Errorf("failed to create bet: %w", err)
	}

	// Deduct from wallet
	movement := wallet.Movement{
		UserID: req.UserID,
		Amount: req.Amount,
		Type:   domain.TransactionTypeBetPlaced,
	}
	if _, err := e.walletService.Debit(ctx, req.UserID, req.Amount, movement); err != nil {
		return nil, fmt.Errorf("failed to deduct bet amount: %w", err)
	}

	// Broadcast updated game state
	gameState := e.getGameState()
	e.hub.BroadcastGameState(gameState)

	return &BetResponse{
		Success:       true,
		Message:       "Bet placed successfully",
		BetID:         bet.ID,
		GameState:     gameState,
		AutoCashoutAt: req.AutoCashoutAt,
	}, nil
}

// Cashout cashes out a bet at the current odds
func (e *CrashGameEngine) Cashout(ctx context.Context, req *CashoutRequest) (*CashoutResponse, error) {
	if e.currentGame == nil {
		return &CashoutResponse{
			Success: false,
			Message: "No active game",
		}, nil
	}

	// Get bet
	bet, err := e.betRepo.GetByID(ctx, req.BetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bet: %w", err)
	}

	// Validate bet
	if bet.UserID != req.UserID {
		return &CashoutResponse{
			Success: false,
			Message: "Unauthorized",
		}, nil
	}

	if bet.Status != domain.GameBetStatusActive {
		return &CashoutResponse{
			Success: false,
			Message: "Bet is not active",
		}, nil
	}

	if bet.GameID != e.currentGame.ID {
		return &CashoutResponse{
			Success: false,
			Message: "Bet is not for current game",
		}, nil
	}

	// Calculate cashout
	currentOdds := e.getCurrentOdds()
	payout := bet.Amount.Mul(currentOdds)

	// Calculate tax
	payoutBreakdown := e.taxEngine.ApplyPayoutTax("KE", payout, bet.Amount)
	netPayout := payoutBreakdown.NetPayout

	// Update bet
	if err := e.betRepo.UpdateCashout(ctx, bet.ID, currentOdds, netPayout); err != nil {
		return nil, fmt.Errorf("failed to update bet: %w", err)
	}

	// Add to wallet
	movement := wallet.Movement{
		UserID: req.UserID,
		Amount: netPayout,
		Type:   domain.TransactionTypeBetWon,
	}
	if _, err := e.walletService.Credit(ctx, req.UserID, netPayout, movement); err != nil {
		return nil, fmt.Errorf("failed to add payout: %w", err)
	}

	// Broadcast updated game state
	gameState := e.getGameState()
	e.hub.BroadcastGameState(gameState)

	return &CashoutResponse{
		Success:     true,
		Message:     "Cashout successful",
		CashoutOdds: currentOdds,
		Payout:      netPayout,
		GameState:   gameState,
	}, nil
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

// gameLoop runs the main game loop
func (e *CrashGameEngine) gameLoop(ctx context.Context) {
	if e.currentGame == nil {
		return
	}

	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	currentOdds := decimal.NewFromFloat(1.0)
	maxOdds := e.currentGame.CrashPoint.Div(decimal.NewFromFloat(100.0))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Increment odds
			currentOdds = currentOdds.Add(decimal.NewFromFloat(0.01))

			// Check if crashed
			if currentOdds.GreaterThanOrEqual(maxOdds) {
				e.crashGame(ctx, currentOdds)
				return
			}

			// Check for auto cashouts
			e.checkAutoCashouts(ctx, currentOdds)

			// Broadcast game state
			gameState := e.getGameState()
			gameState.CurrentOdds = currentOdds
			gameState.MaxOdds = maxOdds
			e.hub.BroadcastGameState(gameState)
		}
	}
}

// crashGame handles game crash
func (e *CrashGameEngine) crashGame(ctx context.Context, crashOdds decimal.Decimal) {
	if e.currentGame == nil {
		return
	}

	// Update game status
	e.currentGame.Status = domain.GameStatusCrashed
	now := time.Now()
	e.currentGame.CrashedAt = &now

	if err := e.gameRepo.UpdateStatus(ctx, e.currentGame.ID, domain.GameStatusCrashed); err != nil {
		log.Printf("Failed to update game status: %v", err)
	}

	// Settle remaining active bets
	e.settleBets(ctx, crashOdds)

	// Broadcast final game state
	gameState := e.getGameState()
	gameState.Status = domain.GameStatusCrashed
	gameState.IsCrashed = true
	gameState.CrashOdds = crashOdds
	e.hub.BroadcastGameState(gameState)

	// Clear current game
	e.currentGame = nil
}

// checkAutoCashouts checks for automatic cashouts
func (e *CrashGameEngine) checkAutoCashouts(ctx context.Context, currentOdds decimal.Decimal) {
	if e.currentGame == nil {
		return
	}

	// Get active bets
	bets, err := e.betRepo.GetActiveByGame(ctx, e.currentGame.ID)
	if err != nil {
		log.Printf("Failed to get active bets: %v", err)
		return
	}

	for _, bet := range bets {
		if bet.CashoutAt != nil && currentOdds.GreaterThanOrEqual(*bet.CashoutAt) {
			// Auto cashout
			payout := bet.Amount.Mul(*bet.CashoutAt)
			payoutBreakdown := e.taxEngine.ApplyPayoutTax("KE", payout, bet.Amount)
			netPayout := payoutBreakdown.NetPayout

			if err := e.betRepo.UpdateCashout(ctx, bet.ID, *bet.CashoutAt, netPayout); err != nil {
				log.Printf("Failed to update auto cashout bet: %v", err)
				continue
			}

			movement := wallet.Movement{
				UserID: bet.UserID,
				Amount: netPayout,
				Type:   domain.TransactionTypeBetWon,
			}
			if _, err := e.walletService.Credit(ctx, bet.UserID, netPayout, movement); err != nil {
				log.Printf("Failed to add auto cashout payout: %v", err)
			}
		}
	}
}

// settleBets settles remaining active bets after crash
func (e *CrashGameEngine) settleBets(ctx context.Context, crashOdds decimal.Decimal) {
	if e.currentGame == nil {
		return
	}

	// Get active bets
	bets, err := e.betRepo.GetActiveByGame(ctx, e.currentGame.ID)
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

// getGameState creates the current game state
func (e *CrashGameEngine) getGameState() *GameState {
	if e.currentGame == nil {
		return &GameState{
			Status: domain.GameStatusWaiting,
		}
	}

	activePlayerCount := 0
	if e.currentGame != nil {
		activePlayerCount = e.hub.GetActivePlayerCount(e.currentGame.ID)
	}

	return &GameState{
		GameID:        e.currentGame.ID,
		RoundNumber:   e.roundNumber,
		Status:        e.currentGame.Status,
		StartedAt:     e.currentGame.StartedAt,
		ActivePlayers: activePlayerCount,
	}
}

// getCurrentOdds returns the current odds
func (e *CrashGameEngine) getCurrentOdds() decimal.Decimal {
	if e.currentGame == nil {
		return decimal.NewFromFloat(1.0)
	}

	// Calculate current odds based on elapsed time
	elapsed := time.Since(e.currentGame.StartedAt)
	ticks := elapsed.Milliseconds() / e.tickInterval.Milliseconds()
	odds := decimal.NewFromFloat(1.0 + float64(ticks)*0.01)

	return odds
}
