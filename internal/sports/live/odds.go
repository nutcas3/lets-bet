package live

import (
	"context"
	"fmt"
	"log"
	"time"
)

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
