package edit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testNow = time.Now()

// MockBetRepository is a mock implementation of the bet repository
type MockBetRepository struct {
	mock.Mock
}

func (m *MockBetRepository) GetByID(ctx context.Context, betID string) (*domain.SportBet, error) {
	args := m.Called(ctx, betID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SportBet), args.Error(1)
}

func (m *MockBetRepository) Update(ctx context.Context, bet *domain.SportBet) error {
	args := m.Called(ctx, bet)
	return args.Error(0)
}

func (m *MockBetRepository) GetByUserID(ctx context.Context, userID string, filters *MockBetFilters) ([]*domain.SportBet, error) {
	args := m.Called(ctx, userID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SportBet), args.Error(1)
}

// MockMatchRepository is a mock implementation of the match repository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) GetByID(ctx context.Context, matchID string) (*domain.Match, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Match), args.Error(1)
}

// MockMarketRepository is a mock implementation of the market repository
type MockMarketRepository struct {
	mock.Mock
}

func (m *MockMarketRepository) GetByID(ctx context.Context, marketID string) (*domain.Market, error) {
	args := m.Called(ctx, marketID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Market), args.Error(1)
}

// MockOutcomeRepository is a mock implementation of the outcome repository
type MockOutcomeRepository struct {
	mock.Mock
}

func (m *MockOutcomeRepository) GetByID(ctx context.Context, outcomeID string) (*domain.Outcome, error) {
	args := m.Called(ctx, outcomeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Outcome), args.Error(1)
}

// MockWalletService is a mock implementation of the wallet service
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error) {
	args := m.Called(ctx, userID, amount, movement)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockWalletService) Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error) {
	args := m.Called(ctx, userID, amount, movement)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

// MockEventBus is a mock implementation of the event bus
type MockEventBus struct {
	mock.Mock
}

func (m *MockEventBus) Publish(topic string, event interface{}) error {
	args := m.Called(topic, event)
	return args.Error(0)
}

// MockBetFilters for test support
type MockBetFilters struct {
	Status *domain.BetStatus
}

// Test helper to create a sport bet with a specific country code
func createTestBet(countryCode string, amount decimal.Decimal, status domain.BetStatus) *domain.SportBet {
	return &domain.SportBet{
		ID:          fmt.Sprintf("bet-%s-001", countryCode),
		UserID:      "user-123",
		CountryCode: countryCode,
		EventID:     "match-456",
		MarketID:    "market-789",
		OutcomeID:   "outcome-101",
		Amount:      amount,
		Odds:        decimal.NewFromFloat(2.5),
		Currency:    "KES",
		Status:      status,
		PlacedAt:    testNow.Add(-1 * time.Minute),
		UpdatedAt:   testNow,
	}
}

// Test ProcessRefund with Kenya (KE) country code
func TestProcessRefundWithKenyaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("KE", decimal.NewFromInt(1000), domain.BetStatusPending)
	refundAmount := decimal.NewFromInt(500)

	// Expect Credit to be called with KE country code
	walletService.On("Credit", ctx, bet.UserID, refundAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "KE"
	})).Return(&Transaction{}, nil)

	err := service.processRefund(ctx, bet, refundAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Credit", ctx, bet.UserID, refundAmount, mock.Anything)
}

// Test ProcessRefund with Nigeria (NG) country code
func TestProcessRefundWithNigeriaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("NG", decimal.NewFromInt(2000), domain.BetStatusPending)
	refundAmount := decimal.NewFromInt(1000)

	// Expect Credit to be called with NG country code
	walletService.On("Credit", ctx, bet.UserID, refundAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "NG"
	})).Return(&Transaction{}, nil)

	err := service.processRefund(ctx, bet, refundAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Credit", ctx, bet.UserID, refundAmount, mock.Anything)
}

// Test ProcessRefund with Ghana (GH) country code
func TestProcessRefundWithGhanaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("GH", decimal.NewFromInt(500), domain.BetStatusPending)
	refundAmount := decimal.NewFromInt(250)

	// Expect Credit to be called with GH country code
	walletService.On("Credit", ctx, bet.UserID, refundAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "GH"
	})).Return(&Transaction{}, nil)

	err := service.processRefund(ctx, bet, refundAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Credit", ctx, bet.UserID, refundAmount, mock.Anything)
}

// Test ProcessAdditionalPayment with Kenya (KE) country code
func TestProcessAdditionalPaymentWithKenyaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("KE", decimal.NewFromInt(1000), domain.BetStatusPending)
	additionalAmount := decimal.NewFromInt(500)

	// Expect Debit to be called with KE country code
	walletService.On("Debit", ctx, bet.UserID, additionalAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "KE"
	})).Return(&Transaction{}, nil)

	err := service.processAdditionalPayment(ctx, bet, additionalAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Debit", ctx, bet.UserID, additionalAmount, mock.Anything)
}

// Test ProcessAdditionalPayment with Nigeria (NG) country code
func TestProcessAdditionalPaymentWithNigeriaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("NG", decimal.NewFromInt(2000), domain.BetStatusPending)
	additionalAmount := decimal.NewFromInt(1000)

	// Expect Debit to be called with NG country code
	walletService.On("Debit", ctx, bet.UserID, additionalAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "NG"
	})).Return(&Transaction{}, nil)

	err := service.processAdditionalPayment(ctx, bet, additionalAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Debit", ctx, bet.UserID, additionalAmount, mock.Anything)
}

// Test ProcessAdditionalPayment with Ghana (GH) country code
func TestProcessAdditionalPaymentWithGhanaCountryCode(t *testing.T) {
	ctx := context.Background()
	walletService := new(MockWalletService)

	service := &EditBetService{
		walletService: walletService,
	}

	bet := createTestBet("GH", decimal.NewFromInt(500), domain.BetStatusPending)
	additionalAmount := decimal.NewFromInt(250)

	// Expect Debit to be called with GH country code
	walletService.On("Debit", ctx, bet.UserID, additionalAmount, mock.MatchedBy(func(m Movement) bool {
		return m.CountryCode == "GH"
	})).Return(&Transaction{}, nil)

	err := service.processAdditionalPayment(ctx, bet, additionalAmount)
	assert.NoError(t, err)
	walletService.AssertCalled(t, "Debit", ctx, bet.UserID, additionalAmount, mock.Anything)
}

// Test that country code is passed through in movement details
func TestMovementContainsCorrectCountryCode(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		betAmount   decimal.Decimal
	}{
		{
			name:        "Kenya",
			countryCode: "KE",
			betAmount:   decimal.NewFromInt(1000),
		},
		{
			name:        "Nigeria",
			countryCode: "NG",
			betAmount:   decimal.NewFromInt(2000),
		},
		{
			name:        "Ghana",
			countryCode: "GH",
			betAmount:   decimal.NewFromInt(500),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			walletService := new(MockWalletService)

			service := &EditBetService{
				walletService: walletService,
			}

			bet := createTestBet(tt.countryCode, tt.betAmount, domain.BetStatusPending)
			refundAmount := tt.betAmount.Div(decimal.NewFromInt(2))

			var capturedMovement Movement

			walletService.On("Credit", ctx, bet.UserID, refundAmount, mock.MatchedBy(func(m Movement) bool {
				capturedMovement = m
				return m.CountryCode == tt.countryCode
			})).Return(&Transaction{}, nil)

			err := service.processRefund(ctx, bet, refundAmount)
			require.NoError(t, err)

			// Verify the captured movement has correct country code
			assert.Equal(t, tt.countryCode, capturedMovement.CountryCode)
			assert.Equal(t, bet.UserID, capturedMovement.UserID)
			assert.Equal(t, refundAmount, capturedMovement.Amount)
		})
	}
}

// Benchmark test for processing refunds across multiple countries
func BenchmarkProcessRefundMultiTenant(b *testing.B) {
	ctx := context.Background()
	walletService := new(MockWalletService)
	walletService.On("Credit", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&Transaction{}, nil)

	service := &EditBetService{
		walletService: walletService,
	}

	countries := []string{"KE", "NG", "GH"}
	refundAmount := decimal.NewFromInt(500)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countryCode := countries[i%len(countries)]
		bet := createTestBet(countryCode, decimal.NewFromInt(1000), domain.BetStatusPending)
		service.processRefund(ctx, bet, refundAmount)
	}
}
