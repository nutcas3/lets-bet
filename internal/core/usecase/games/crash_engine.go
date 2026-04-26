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

// handlePlaceBet processes bet requests in the manager goroutine
func (e *CrashGameEngine) handlePlaceBet(req BetRequest) {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil || game.Status != domain.GameStatusRunning {
		req.Resp <- fmt.Errorf("game not active")
		return
	}

	// Create bet with atomic wallet update in transaction
	// This is the ONLY place where game state changes
	// No Mutex needed inside this block!
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

	// Get bet to validate and calculate payout
	bet, err := e.betRepo.GetByID(context.Background(), req.BetID)
	if err != nil {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Bet not found",
		}
		return
	}

	if bet.Status != domain.GameBetStatusActive {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Bet is not active",
		}
		return
	}

	if bet.GameID != game.ID {
		req.Resp <- &CashoutResponse{
			Success: false,
			Message: "Bet is not for current game",
		}
		return
	}

	// Calculate cashout
	currentOdds := e.calculateCurrentOdds(game.StartedAt)
	payout := bet.Amount.Mul(currentOdds)

	// Atomic SQL update for double-cashout prevention
	// UPDATE bets SET status = 'cashed_out', cashout_at = ?, payout = ?
	// WHERE id = ? AND status = 'active'
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
			Message: "Bet already cashed out or game crashed",
		}
		return
	}

	// Credit wallet atomically
	movement := wallet.Movement{
		UserID: req.UserID,
		Amount: payout,
		Type:   domain.TransactionTypeBetWon,
	}
	if _, err := e.walletService.Credit(context.Background(), req.UserID, payout, movement); err != nil {
		// Log error but don't fail the cashout since bet was already updated
		log.Printf("Failed to credit wallet for user %s: %v", req.UserID, err)
	}

	// Return success response
	req.Resp <- &CashoutResponse{
		Success:     true,
		Message:     "Cashout successful",
		CashoutOdds: currentOdds,
		Payout:      payout,
	}
}

// handleCommand processes control commands
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

// tick updates game state and broadcasts
func (e *CrashGameEngine) tick() {
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil || game.Status != domain.GameStatusRunning {
		return
	}

	// Calculate current odds
	currentOdds := e.calculateCurrentOdds(game.StartedAt)
	maxOdds := game.CrashPoint.Div(decimal.NewFromFloat(100.0))

	// Check if crashed
	if currentOdds.GreaterThanOrEqual(maxOdds) {
		e.commandChan <- "CRASH"
		return
	}

	// Generate and store immutable snapshot for lock-free reads
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

	// Broadcast to WebSocket hub
	e.hub.BroadcastGameState(state)
}

// startGameInternal handles game startup in the manager goroutine
func (e *CrashGameEngine) startGameInternal() {
	// Cancel any existing game loop
	if e.gameCancel != nil {
		e.gameCancel()
	}

	e.mu.Lock()
	e.roundNumber++
	roundNumber := e.roundNumber
	e.mu.Unlock()

	gameID := uuid.New()

	// Generate provably fair crash point
	seed, err := e.fairService.GenerateServerSeed()
	if err != nil {
		log.Printf("Failed to generate server seed: %v", err)
		return
	}
	hash := e.fairService.HashServerSeed(seed)
	crashPoint := e.fairService.CalculateCrashPoint(seed, "", roundNumber)

	// Create new game
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
		return
	}

	// Set current game
	e.mu.Lock()
	e.currentGame = game
	e.mu.Unlock()
}

// crashGameInternal handles game crash in the manager goroutine
func (e *CrashGameEngine) crashGameInternal() {
	e.mu.Lock()
	currentGame := e.currentGame
	if currentGame != nil {
		// Update game status
		currentGame.Status = domain.GameStatusCrashed
		now := time.Now()
		currentGame.CrashedAt = &now
		e.currentGame = currentGame
	}
	e.mu.Unlock()

	if currentGame == nil {
		return
	}

	if err := e.gameRepo.UpdateStatus(context.Background(), currentGame.ID, domain.GameStatusCrashed); err != nil {
		log.Printf("Failed to update game status: %v", err)
	}

	// Settle remaining active bets
	e.settleBets(context.Background(), decimal.NewFromFloat(2.0))

	// Clear current game
	e.mu.Lock()
	e.currentGame = nil
	e.mu.Unlock()

	// Store final state
	state := &GameState{
		Status: domain.GameStatusWaiting,
	}
	e.latestState.Store(state)
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

// gameLoop runs the main game loop
func (e *CrashGameEngine) gameLoop(ctx context.Context) {
	// Hold RLock for the entire loop body to prevent race condition
	e.mu.RLock()
	currentGame := e.currentGame
	if currentGame == nil {
		e.mu.RUnlock()
		return
	}
	// Copy CrashPoint while holding the lock
	crashPoint := currentGame.CrashPoint
	e.mu.RUnlock()

	ticker := time.NewTicker(e.tickInterval)
	defer ticker.Stop()

	currentOdds := decimal.NewFromFloat(1.0)
	maxOdds := crashPoint.Div(decimal.NewFromFloat(100.0))

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
	e.mu.Lock()
	currentGame := e.currentGame
	var gameID uuid.UUID
	if currentGame != nil {
		// Update game status
		currentGame.Status = domain.GameStatusCrashed
		now := time.Now()
		currentGame.CrashedAt = &now
		gameID = currentGame.ID
		e.currentGame = currentGame
	}
	e.mu.Unlock()

	if currentGame == nil {
		return
	}

	if err := e.gameRepo.UpdateStatus(ctx, gameID, domain.GameStatusCrashed); err != nil {
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
	e.mu.Lock()
	e.currentGame = nil
	e.mu.Unlock()
}

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
			// This wraps bet update and wallet credit in a single database transaction
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
func (e *CrashGameEngine) settleBets(ctx context.Context, crashOdds decimal.Decimal) {
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

// getGameState creates the current game state using atomic.Value for lock-free reads
func (e *CrashGameEngine) getGameState() *GameState {
	// Try atomic load first for maximum performance
	if state := e.latestState.Load(); state != nil {
		return state.(*GameState)
	}

	// Fallback to atomic snapshot pattern if no state stored
	e.mu.RLock()
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil {
		return &GameState{Status: domain.GameStatusWaiting}
	}

	// Use the local 'game' variable. Even if e.currentGame is wiped
	// by the gameLoop thread, this local pointer remains valid.
	activePlayerCount := e.hub.GetActivePlayerCount(game.ID)

	e.mu.RLock()
	roundNumber := e.roundNumber
	e.mu.RUnlock()

	return &GameState{
		GameID:        game.ID,
		RoundNumber:   roundNumber,
		Status:        game.Status,
		StartedAt:     game.StartedAt,
		CurrentOdds:   e.calculateCurrentOdds(game.StartedAt),
		ActivePlayers: activePlayerCount,
	}
}

// calculateCurrentOdds calculates odds based on elapsed time to eliminate drift
func (e *CrashGameEngine) calculateCurrentOdds(startedAt time.Time) decimal.Decimal {
	elapsed := time.Since(startedAt).Seconds()
	// Example: 1.0 + 0.1x per second (adjust multiplier for your curve)
	odds := decimal.NewFromFloat(1.0 + (elapsed * 0.1))
	return odds
}

// getCurrentOdds returns the current odds using atomic snapshot pattern
func (e *CrashGameEngine) getCurrentOdds() decimal.Decimal {
	e.mu.RLock()
	// Capture the pointer and immediately release the lock
	game := e.currentGame
	e.mu.RUnlock()

	if game == nil {
		return decimal.NewFromFloat(1.0)
	}

	// Use the local 'game' variable for time-based calculation
	return e.calculateCurrentOdds(game.StartedAt)
}
