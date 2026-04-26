package live

import (
	"context"
	"fmt"
	"log"
)

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
