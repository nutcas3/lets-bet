// Package live provides live sports betting functionality
package live

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/betting-platform/internal/odds/genius"
	"github.com/betting-platform/internal/odds/sportradar"
)

// NewLiveBettingService creates a new live betting service
func NewLiveBettingService(
	matchRepo postgres.MatchRepository,
	betRepo postgres.SportBetRepository,
	marketRepo postgres.BettingMarketRepository,
	outcomeRepo postgres.MarketOutcomeRepository,
	sportradarClient *sportradar.SportradarClient,
	geniusClient *genius.GeniusClient,
	eventBus EventBus,
) *LiveBettingService {
	betIDGenerator, err := id.ServiceTypeGenerator("betting")
	if err != nil {
		panic(fmt.Sprintf("Failed to create betting ID generator: %v", err))
	}

	service := &LiveBettingService{
		matchRepo:          matchRepo,
		betRepo:            betRepo,
		marketRepo:         marketRepo,
		outcomeRepo:        outcomeRepo,
		sportradarClient:   sportradarClient,
		geniusClient:       geniusClient,
		eventBus:           eventBus,
		betIDGenerator:     betIDGenerator,
		liveMatches:        make(map[string]*LiveMatch),
		oddsUpdates:        make(map[string]*OddsUpdate),
		settlementQueue:    make(chan *SettlementRequest, 1000),
		oddsUpdateInterval: 10 * time.Second,
		settlementInterval: 30 * time.Second,
		maxOddsDelay:       5 * time.Second,
	}

	return service
}

// StartLiveBetting starts the live betting service
func (s *LiveBettingService) StartLiveBetting(ctx context.Context) error {
	log.Println("Starting live betting service")

	// Start odds update goroutine
	go s.runOddsUpdates(ctx)

	// Start settlement goroutine
	go s.runSettlement(ctx)

	// Subscribe to external events
	if err := s.eventBus.Subscribe("match.started", s.handleMatchStarted); err != nil {
		return fmt.Errorf("failed to subscribe to match started events: %w", err)
	}

	if err := s.eventBus.Subscribe("match.ended", s.handleMatchEnded); err != nil {
		return fmt.Errorf("failed to subscribe to match ended events: %w", err)
	}

	return nil
}

// PlaceLiveBet places a bet on a live match
func (s *LiveBettingService) PlaceLiveBet(ctx context.Context, req *LiveBetRequest) (*LiveBetResponse, error) {
	s.liveMatchesMutex.RLock()
	liveMatch, exists := s.liveMatches[req.MatchID]
	s.liveMatchesMutex.RUnlock()

	if !exists {
		return &LiveBetResponse{
			Success: false,
			Message: "match not found",
		}, nil
	}

	if liveMatch.IsSuspended {
		return &LiveBetResponse{
			Success: false,
			Message: "match is suspended",
		}, nil
	}

	// Validate odds
	if err := s.validateLiveOdds(ctx, req.MatchID, req.Odds); err != nil {
		return &LiveBetResponse{
			Success: false,
			Message: fmt.Sprintf("invalid odds: %v", err),
		}, nil
	}

	// Create bet
	if _, err := uuid.Parse(req.UserID); err != nil {
		return &LiveBetResponse{
			Success: false,
			Message: "invalid user ID",
		}, nil
	}

	bet := &domain.SportBet{
		ID:        s.generateBetID(),
		UserID:    req.UserID,
		EventID:   req.MatchID,
		MarketID:  req.MarketID,
		OutcomeID: req.OutcomeID,
		Amount:    req.Amount,
		Odds:      req.Odds,
		Currency:  "KES", // Default currency
		Status:    domain.BetStatusPending,
		PlacedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save bet
	if err := s.betRepo.Create(ctx, bet); err != nil {
		return &LiveBetResponse{
			Success: false,
			Message: "failed to save bet",
		}, err
	}

	// Publish event
	if err := s.eventBus.Publish("live.bet.placed", bet); err != nil {
		log.Printf("failed to publish live bet placed event: %v", err)
	}

	return &LiveBetResponse{
		Success:      true,
		Message:      "bet placed successfully",
		BetID:        bet.ID,
		Amount:       req.Amount,
		Odds:         req.Odds,
		PotentialWin: req.Amount.Mul(req.Odds),
	}, nil
}

// GetMetrics returns service metrics
func (s *LiveBettingService) GetMetrics(ctx context.Context) (*LiveBettingMetrics, error) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	totalMatches := len(s.liveMatches)
	activeMatches := 0
	suspendedMatches := 0

	for _, match := range s.liveMatches {
		if match.IsSuspended {
			suspendedMatches++
		} else {
			activeMatches++
		}
	}

	return &LiveBettingMetrics{
		TotalMatches:       int64(totalMatches),
		ActiveMatches:      int64(activeMatches),
		SuspendedMatches:   int64(suspendedMatches),
		OddsUpdateInterval: s.oddsUpdateInterval,
		LastActivity:       time.Now(),
	}, nil
}
