// Package live provides live sports betting functionality
package live

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/betting-platform/internal/odds/genius"
	"github.com/betting-platform/internal/odds/sportradar"
)

// LiveBetRequest represents a request to place a live bet
type LiveBetRequest struct {
	MatchID   string          `json:"match_id"`
	MarketID  string          `json:"market_id"`
	OutcomeID string          `json:"outcome_id"`
	Amount    decimal.Decimal `json:"amount"`
	Odds      decimal.Decimal `json:"odds"`
	UserID    string          `json:"user_id"`
}

// LiveBetResponse represents the response to a live bet request
type LiveBetResponse struct {
	Success      bool            `json:"success"`
	BetID        string          `json:"bet_id,omitempty"`
	Message      string          `json:"message"`
	Balance      decimal.Decimal `json:"balance"`
	Odds         decimal.Decimal `json:"odds"`
	Amount       decimal.Decimal `json:"amount"`
	PotentialWin decimal.Decimal `json:"potential_win"`
}

// OddsUpdateRequest represents a request to update odds
type OddsUpdateRequest struct {
	MatchID   string             `json:"match_id"`
	MarketID  string             `json:"market_id"`
	OutcomeID string             `json:"outcome_id"`
	NewOdds   decimal.Decimal    `json:"new_odds"`
	Markets   []MarketOddsUpdate `json:"markets"`
	MatchOdds *MatchOddsUpdate   `json:"match_odds,omitempty"`
}

// MarketOddsUpdate represents odds update for a market
type MarketOddsUpdate struct {
	MarketID string              `json:"market_id"`
	Outcomes []OutcomeOddsUpdate `json:"outcomes"`
}

// OutcomeOddsUpdate represents odds update for an outcome
type OutcomeOddsUpdate struct {
	OutcomeID string          `json:"outcome_id"`
	Odds      decimal.Decimal `json:"odds"`
}

// MatchOddsUpdate represents match-level odds update
type MatchOddsUpdate struct {
	HomeWin decimal.Decimal `json:"home_win"`
	Draw    decimal.Decimal `json:"draw"`
	AwayWin decimal.Decimal `json:"away_win"`
}

// LiveBettingMetrics represents live betting metrics
type LiveBettingMetrics struct {
	ActiveMatches      int64           `json:"active_matches"`
	TotalBets          int64           `json:"total_bets"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	AverageBetSize     decimal.Decimal `json:"average_bet_size"`
	LastUpdated        time.Time       `json:"last_updated"`
	TotalMatches       int64           `json:"total_matches"`
	SuspendedMatches   int64           `json:"suspended_matches"`
	OddsUpdateInterval time.Duration   `json:"odds_update_interval"`
	LastActivity       time.Time       `json:"last_activity"`
}

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

// GetLiveMatch retrieves a live match by ID
func (s *LiveBettingService) GetLiveMatch(ctx context.Context, matchID string) (*LiveMatch, error) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	liveMatch, exists := s.liveMatches[matchID]
	if !exists {
		return nil, fmt.Errorf("live match not found: %s", matchID)
	}

	return liveMatch, nil
}

// GetLiveMatches retrieves all live matches
func (s *LiveBettingService) GetLiveMatches(ctx context.Context) ([]*LiveMatch, error) {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	var matches []*LiveMatch
	for _, match := range s.liveMatches {
		matches = append(matches, match)
	}

	return matches, nil
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

// UpdateLiveOdds updates odds for a live match
func (s *LiveBettingService) UpdateLiveOdds(ctx context.Context, req *OddsUpdateRequest) error {
	s.liveMatchesMutex.RLock()
	liveMatch, exists := s.liveMatches[req.MatchID]
	s.liveMatchesMutex.RUnlock()

	if !exists {
		return fmt.Errorf("live match not found: %s", req.MatchID)
	}

	// Update odds in live match
	for _, market := range liveMatch.LiveMarkets {
		if market.Market.ID == req.MarketID {
			for _, outcome := range market.LiveOdds {
				if outcome.Outcome.ID == req.OutcomeID {
					// Record odds update
					update := &OddsUpdate{
						MatchID:    req.MatchID,
						MarketID:   req.MarketID,
						OutcomeID:  req.OutcomeID,
						OldOdds:    outcome.CurrentOdds,
						NewOdds:    req.NewOdds,
						UpdateTime: time.Now(),
						UpdateType: "manual",
					}
					s.recordOddsUpdate(ctx, update)

					// Update outcome odds
					outcome.PreviousOdds = outcome.CurrentOdds
					outcome.CurrentOdds = req.NewOdds
					outcome.OddsChangeTime = time.Now()
					outcome.OddsChangeCount++
					market.LastOddsUpdate = time.Now()
					market.OddsUpdateCount++

					// Publish event
					if err := s.eventBus.Publish("live.odds.updated", req); err != nil {
						log.Printf("failed to publish odds updated event: %v", err)
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("market or outcome not found")
}

// SuspendMatch suspends a live match
func (s *LiveBettingService) SuspendMatch(ctx context.Context, matchID, reason string) error {
	s.liveMatchesMutex.Lock()
	defer s.liveMatchesMutex.Unlock()

	liveMatch, exists := s.liveMatches[matchID]
	if !exists {
		return fmt.Errorf("live match not found: %s", matchID)
	}

	liveMatch.IsSuspended = true
	liveMatch.SuspensionReason = reason

	// Suspend all markets
	for _, market := range liveMatch.LiveMarkets {
		market.IsSuspended = true
		market.SuspensionReason = reason
	}

	// Publish event
	if err := s.eventBus.Publish("live.match.suspended", map[string]string{
		"matchID": matchID,
		"reason":  reason,
	}); err != nil {
		log.Printf("failed to publish match suspended event: %v", err)
	}

	return nil
}

// runOddsUpdates runs the odds update loop
func (s *LiveBettingService) runOddsUpdates(ctx context.Context) {
	ticker := time.NewTicker(s.oddsUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.updateOdds(ctx); err != nil {
				log.Printf("failed to update odds: %v", err)
			}
		}
	}
}

// runSettlement runs the settlement loop
func (s *LiveBettingService) runSettlement(ctx context.Context) {
	ticker := time.NewTicker(s.settlementInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.processSettlements(ctx); err != nil {
				log.Printf("failed to process settlements: %v", err)
			}
		}
	}
}

// handleMatchStarted handles match started events
func (s *LiveBettingService) handleMatchStarted(event any) {
	// Implementation for handling match started events
	log.Printf("Match started event received: %v", event)
}

// handleMatchEnded handles match ended events
func (s *LiveBettingService) handleMatchEnded(event any) {
	// Implementation for handling match ended events
	log.Printf("Match ended event received: %v", event)
}

// updateOdds updates odds for all live matches
func (s *LiveBettingService) updateOdds(ctx context.Context) error {
	s.liveMatchesMutex.RLock()
	defer s.liveMatchesMutex.RUnlock()

	for matchID := range s.liveMatches {
		// Update odds logic here
		log.Printf("Updating odds for match: %s", matchID)
	}

	return nil
}

// processSettlements processes pending settlements
func (s *LiveBettingService) processSettlements(ctx context.Context) error {
	// Settlement processing logic here
	log.Println("Processing settlements")
	return nil
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
