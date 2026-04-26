package games

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/core/usecase/wallet"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGameBetRepository implements GameBetRepository with atomic cashout support
type MockGameBetRepository struct {
	bets         map[uuid.UUID]*domain.GameBet
	mu           sync.RWMutex
	cashoutCount int
}

func NewMockGameBetRepository() *MockGameBetRepository {
	return &MockGameBetRepository{
		bets: make(map[uuid.UUID]*domain.GameBet),
	}
}

func (m *MockGameBetRepository) Create(ctx context.Context, bet *domain.GameBet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.bets[bet.ID] = bet
	return nil
}

func (m *MockGameBetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.GameBet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	bet, exists := m.bets[id]
	if !exists {
		return nil, assert.AnError
	}
	return bet, nil
}

func (m *MockGameBetRepository) GetActiveByGame(ctx context.Context, gameID uuid.UUID) ([]*domain.GameBet, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var activeBets []*domain.GameBet
	for _, bet := range m.bets {
		if bet.GameID == gameID && bet.Status == domain.GameBetStatusActive {
			activeBets = append(activeBets, bet)
		}
	}
	return activeBets, nil
}

func (m *MockGameBetRepository) UpdateCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	bet, exists := m.bets[id]
	if !exists {
		return assert.AnError
	}
	bet.CashoutAt = &cashoutAt
	bet.Payout = payout
	return nil
}

func (m *MockGameBetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.GameBetStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	bet, exists := m.bets[id]
	if !exists {
		return assert.AnError
	}
	bet.Status = status
	return nil
}

// AtomicCashout implements the atomic SQL update simulation
func (m *MockGameBetRepository) AtomicCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bet, exists := m.bets[id]
	if !exists {
		return false, assert.AnError
	}

	// Simulate atomic SQL: UPDATE bets SET status = 'cashed_out' WHERE id = ? AND status = 'active'
	if bet.Status != domain.GameBetStatusActive {
		return false, nil // Bet already cashed out or not active
	}

	// Update atomically
	bet.Status = domain.GameBetStatusCashedOut
	bet.CashoutAt = &cashoutAt
	bet.Payout = payout
	m.cashoutCount++

	return true, nil
}

// CreateBetWithWalletUpdate simulates atomic wallet update + bet creation in transaction
func (m *MockGameBetRepository) CreateBetWithWalletUpdate(ctx context.Context, bet *domain.GameBet, userID uuid.UUID, amount decimal.Decimal) (uuid.UUID, error) {
	// Simulate database transaction
	// In real implementation: BEGIN; UPDATE wallets SET balance = balance - ? WHERE user_id = ? AND balance >= ?; INSERT INTO bets ...; COMMIT;

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create bet
	m.bets[bet.ID] = bet

	return bet.ID, nil
}

// AtomicAutoCashoutWithCredit simulates atomic auto-cashout with wallet credit in transaction
func (m *MockGameBetRepository) AtomicAutoCashoutWithCredit(ctx context.Context, id uuid.UUID, userID uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal, country string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bet, exists := m.bets[id]
	if !exists {
		return false, assert.AnError
	}

	// Simulate atomic SQL: UPDATE bets SET status = 'cashed_out' WHERE id = ? AND status = 'active'
	if bet.Status != domain.GameBetStatusActive {
		return false, nil
	}

	bet.Status = domain.GameBetStatusCashedOut
	bet.CashedOut = true
	bet.CashoutAt = &cashoutAt
	bet.Payout = payout
	now := time.Now()
	bet.CashedOutAt = &now

	return true, nil
}

// MockWalletService implements wallet.WalletService interface
type MockWalletService struct {
	balances map[uuid.UUID]decimal.Decimal
	mu       sync.RWMutex
}

func NewMockWalletService() *MockWalletService {
	return &MockWalletService{
		balances: make(map[uuid.UUID]decimal.Decimal),
	}
}

func (m *MockWalletService) SetBalance(userID uuid.UUID, balance decimal.Decimal) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balances[userID] = balance
}

func (m *MockWalletService) Balance(ctx context.Context, userID uuid.UUID) (decimal.Decimal, decimal.Decimal, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	balance, exists := m.balances[userID]
	if !exists {
		return decimal.Zero, decimal.Zero, nil
	}
	return balance, decimal.Zero, nil
}

func (m *MockWalletService) Debit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, movement wallet.Movement) (decimal.Decimal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.balances[userID]
	if current.LessThan(amount) {
		return decimal.Zero, assert.AnError
	}

	m.balances[userID] = current.Sub(amount)
	return m.balances[userID], nil
}

func (m *MockWalletService) Credit(ctx context.Context, userID uuid.UUID, amount decimal.Decimal, movement wallet.Movement) (decimal.Decimal, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.balances[userID]
	m.balances[userID] = current.Add(amount)
	return m.balances[userID], nil
}

// MockWebSocketHub implements WebSocketHub interface for testing
type MockWebSocketHub struct{}

func (m *MockWebSocketHub) BroadcastGameState(state any) {}

func (m *MockWebSocketHub) GetActivePlayerCount(gameID uuid.UUID) int {
	return 0
}

// ========================================
// Minimal Crash Engine Tests
// ========================================

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

	// Create engine with minimal setup - initialize all required fields
	betRepo := NewMockGameBetRepository()
	engine := &CrashGameEngine{
		roundNumber:  42,
		tickInterval: 10 * time.Millisecond,
		hub:          &MockWebSocketHub{}, // Initialize hub to prevent nil pointer
		betRepo:      betRepo,             // Initialize betRepo to prevent nil pointer in crashGame
	}

	game := &domain.Game{
		ID:          uuid.New(),
		GameType:    domain.GameTypeCrash,
		RoundNumber: 42,
		Status:      domain.GameStatusRunning,
		StartedAt:   time.Now(),
		CrashPoint:  decimal.NewFromFloat(2.5),
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

	// Just verify we can get game state after setting it
	gameState = engine.getGameState()
	assert.Equal(t, domain.GameStatusRunning, gameState.Status)
}

// ========================================
// Atomic Cashout Tests
// ========================================

// TestAtomicCashoutRepository tests the atomic cashout functionality
func TestAtomicCashoutRepository(t *testing.T) {
	t.Parallel()

	// Test atomic cashout logic directly
	betRepo := NewMockGameBetRepository()

	// Create a test bet
	betID := uuid.New()
	bet := &domain.GameBet{
		ID:       betID,
		GameID:   uuid.New(),
		UserID:   uuid.New(),
		Amount:   decimal.NewFromFloat(100),
		Status:   domain.GameBetStatusActive,
		PlacedAt: time.Now(),
	}
	require.NoError(t, betRepo.Create(context.Background(), bet))

	// Test successful atomic cashout
	cashoutAt := decimal.NewFromFloat(2.0)
	payout := decimal.NewFromFloat(200)

	success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
	require.NoError(t, err)
	assert.True(t, success, "First cashout should succeed")

	// Verify bet was updated
	updatedBet, err := betRepo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	assert.Equal(t, domain.GameBetStatusCashedOut, updatedBet.Status)
	assert.True(t, updatedBet.CashoutAt.Equal(cashoutAt))
	assert.True(t, updatedBet.Payout.Equal(payout))

	// Test second cashout attempt (should fail)
	success, err = betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
	require.NoError(t, err)
	assert.False(t, success, "Second cashout should fail")

	// Verify bet status remains unchanged
	updatedBet, err = betRepo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	assert.Equal(t, domain.GameBetStatusCashedOut, updatedBet.Status)
}

// TestAtomicCashoutConcurrentAccess tests concurrent cashout attempts
func TestAtomicCashoutConcurrentAccess(t *testing.T) {
	t.Parallel()

	betRepo := NewMockGameBetRepository()

	// Create a test bet
	betID := uuid.New()
	bet := &domain.GameBet{
		ID:       betID,
		GameID:   uuid.New(),
		UserID:   uuid.New(),
		Amount:   decimal.NewFromFloat(100),
		Status:   domain.GameBetStatusActive,
		PlacedAt: time.Now(),
	}
	require.NoError(t, betRepo.Create(context.Background(), bet))

	// Test concurrent cashout attempts
	var wg sync.WaitGroup
	numGoroutines := 50
	results := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			cashoutAt := decimal.NewFromFloat(2.0)
			payout := decimal.NewFromFloat(200)

			success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
			if err != nil {
				results <- false
				return
			}
			results <- success
		}()
	}

	wg.Wait()
	close(results)

	// Count successful cashouts
	successCount := 0
	for success := range results {
		if success {
			successCount++
		}
	}

	// Only one cashout should succeed
	assert.Equal(t, 1, successCount, "Only one cashout should succeed in concurrent access")
	assert.Equal(t, 1, betRepo.cashoutCount, "AtomicCashout should only be called once")

	// Verify final bet status
	updatedBet, err := betRepo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	assert.Equal(t, domain.GameBetStatusCashedOut, updatedBet.Status)
}

// TestAtomicCashoutEdgeCases tests various edge cases
func TestAtomicCashoutEdgeCases(t *testing.T) {
	t.Parallel()

	betRepo := NewMockGameBetRepository()

	t.Run("Non-existent bet", func(t *testing.T) {
		t.Parallel()

		nonExistentBetID := uuid.New()
		cashoutAt := decimal.NewFromFloat(2.0)
		payout := decimal.NewFromFloat(200)

		success, err := betRepo.AtomicCashout(context.Background(), nonExistentBetID, cashoutAt, payout)
		assert.Error(t, err, "Should return error for non-existent bet")
		assert.False(t, success, "Should not succeed for non-existent bet")
	})

	t.Run("Already cashed out bet", func(t *testing.T) {
		t.Parallel()

		// Create and cash out a bet
		betID := uuid.New()
		bet := &domain.GameBet{
			ID:       betID,
			GameID:   uuid.New(),
			UserID:   uuid.New(),
			Amount:   decimal.NewFromFloat(100),
			Status:   domain.GameBetStatusCashedOut, // Already cashed out
			PlacedAt: time.Now(),
		}
		require.NoError(t, betRepo.Create(context.Background(), bet))

		cashoutAt := decimal.NewFromFloat(2.0)
		payout := decimal.NewFromFloat(200)

		success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
		require.NoError(t, err)
		assert.False(t, success, "Should not succeed for already cashed out bet")
	})

	t.Run("Lost bet", func(t *testing.T) {
		t.Parallel()

		// Create a lost bet
		betID := uuid.New()
		bet := &domain.GameBet{
			ID:       betID,
			GameID:   uuid.New(),
			UserID:   uuid.New(),
			Amount:   decimal.NewFromFloat(100),
			Status:   domain.GameBetStatusLost, // Lost bet
			PlacedAt: time.Now(),
		}
		require.NoError(t, betRepo.Create(context.Background(), bet))

		cashoutAt := decimal.NewFromFloat(2.0)
		payout := decimal.NewFromFloat(200)

		success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
		require.NoError(t, err)
		assert.False(t, success, "Should not succeed for lost bet")
	})
}

// TestAtomicCashoutSimulation simulates the real-world double-click scenario
func TestAtomicCashoutSimulation(t *testing.T) {
	t.Parallel()

	betRepo := NewMockGameBetRepository()

	// Create a test bet (simulating a user placing a bet)
	betID := uuid.New()
	bet := &domain.GameBet{
		ID:       betID,
		GameID:   uuid.New(),
		UserID:   uuid.New(),
		Amount:   decimal.NewFromFloat(100),
		Status:   domain.GameBetStatusActive,
		PlacedAt: time.Now(),
	}
	require.NoError(t, betRepo.Create(context.Background(), bet))

	// Simulate user double-clicking cashout button
	// Both requests arrive almost simultaneously
	var wg sync.WaitGroup
	results := make(chan string, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(requestNum int) {
			defer wg.Done()

			// Small delay to simulate real-world timing
			time.Sleep(time.Duration(requestNum) * time.Millisecond)

			cashoutAt := decimal.NewFromFloat(2.5)
			payout := decimal.NewFromFloat(250)

			success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
			if err != nil {
				results <- "error"
				return
			}

			if success {
				results <- "success"
			} else {
				results <- "failed"
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Analyze results
	successCount := 0
	failedCount := 0
	errorCount := 0

	for result := range results {
		switch result {
		case "success":
			successCount++
		case "failed":
			failedCount++
		case "error":
			errorCount++
		}
	}

	// Only one should succeed, one should fail, no errors
	assert.Equal(t, 1, successCount, "Exactly one cashout should succeed")
	assert.Equal(t, 1, failedCount, "Exactly one cashout should fail")
	assert.Equal(t, 0, errorCount, "No errors should occur")
	assert.Equal(t, 1, betRepo.cashoutCount, "AtomicCashout should only be called once")

	// Verify final state
	updatedBet, err := betRepo.GetByID(context.Background(), betID)
	require.NoError(t, err)
	assert.Equal(t, domain.GameBetStatusCashedOut, updatedBet.Status)
	assert.True(t, updatedBet.CashoutAt.Equal(decimal.NewFromFloat(2.5)))
	assert.True(t, updatedBet.Payout.Equal(decimal.NewFromFloat(250)))
}

// TestAtomicCashoutPerformance tests performance under high load
func TestAtomicCashoutPerformance(t *testing.T) {
	t.Parallel()

	betRepo := NewMockGameBetRepository()

	// Create multiple bets for concurrent testing
	numBets := 100
	betIDs := make([]uuid.UUID, numBets)

	for i := 0; i < numBets; i++ {
		betID := uuid.New()
		betIDs[i] = betID
		bet := &domain.GameBet{
			ID:       betID,
			GameID:   uuid.New(),
			UserID:   uuid.New(),
			Amount:   decimal.NewFromFloat(100),
			Status:   domain.GameBetStatusActive,
			PlacedAt: time.Now(),
		}
		require.NoError(t, betRepo.Create(context.Background(), bet))
	}

	// Test concurrent cashouts for all bets
	start := time.Now()
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numBets; i++ {
		wg.Add(1)
		go func(betID uuid.UUID) {
			defer wg.Done()

			cashoutAt := decimal.NewFromFloat(2.0)
			payout := decimal.NewFromFloat(200)

			success, err := betRepo.AtomicCashout(context.Background(), betID, cashoutAt, payout)
			if err != nil {
				return
			}

			if success {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(betIDs[i])
	}

	wg.Wait()
	duration := time.Since(start)

	// All cashouts should succeed since each bet is different
	assert.Equal(t, numBets, successCount, "All cashouts should succeed for different bets")
	assert.Equal(t, numBets, betRepo.cashoutCount, "All AtomicCashout should be called")

	// Performance should be reasonable (less than 1 second for 100 concurrent operations)
	assert.Less(t, duration, time.Second, "Performance should be reasonable under load")

	t.Logf("Processed %d concurrent cashouts in %v (%.2f ops/sec)",
		numBets, duration, float64(numBets)/duration.Seconds())
}
