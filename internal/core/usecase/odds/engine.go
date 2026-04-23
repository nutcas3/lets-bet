package odds

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// OddsEngine manages odds calculation and updates for sports betting
type OddsEngine struct {
	oddsProvider OddsProvider
	cache        OddsCache
	eventBus     EventBus
}

// OddsProvider interface for external odds feeds (Sportradar, Genius Sports, etc.)
type OddsProvider interface {
	GetLiveOdds(ctx context.Context) ([]*domain.Match, error)
	GetMatchOdds(ctx context.Context, matchID string) (*domain.Match, error)
	CalculateParlayOdds(odds []decimal.Decimal) decimal.Decimal
}

// OddsCache interface for caching odds data
type OddsCache interface {
	Set(key string, odds any, ttl time.Duration) error
	Get(key string) (any, error)
	Delete(key string) error
}

// EventBus interface for publishing odds updates
type EventBus interface {
	Publish(topic string, message any) error
}

func NewOddsEngine(provider OddsProvider, cache OddsCache, eventBus EventBus) *OddsEngine {
	return &OddsEngine{
		oddsProvider: provider,
		cache:        cache,
		eventBus:     eventBus,
	}
}

// StartOddsSync begins continuous odds synchronization
func (e *OddsEngine) StartOddsSync(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("Starting odds sync from provider...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Odds sync stopped")
			return
		case <-ticker.C:
			e.syncOdds(ctx)
		}
	}
}

func (e *OddsEngine) syncOdds(ctx context.Context) {
	// Fetch live odds from provider
	matches, err := e.oddsProvider.GetLiveOdds(ctx)
	if err != nil {
		log.Printf("Error fetching odds: %v", err)
		return
	}

	// Cache odds
	for _, match := range matches {
		cacheKey := "match:" + match.ID
		_ = e.cache.Set(cacheKey, match, 30*time.Second)
	}

	// Publish odds updates
	if err := e.eventBus.Publish("odds.updated", matches); err != nil {
		log.Printf("Error publishing odds update: %v", err)
	}

	log.Printf("Synced %d matches", len(matches))
}

// CalculateBetOdds calculates total odds for a bet slip
func (e *OddsEngine) CalculateBetOdds(bet *domain.Bet) (*BetOddsResult, error) {
	//BEFORE
	//var totalOdds decimal.Decimal

	//AFTER
	totalOdds := decimal.NewFromInt(1)
	var validSelections []*domain.Selection

	for _, selection := range bet.Selections {
		// Get current odds for this selection
		match, err := e.getMatchOdds(selection.EventID)
		if err != nil {
			continue // Skip invalid selections
		}

		// Find odds for this market/outcome
		selectionOdds, found := e.findSelectionOdds(match, &selection)
		if !found {
			continue // Skip if odds not found
		}

		totalOdds = totalOdds.Mul(selectionOdds)
		validSelections = append(validSelections, &selection)
	}

	if len(validSelections) == 0 {
		return nil, ErrNoValidSelections
	}

	return &BetOddsResult{
		TotalOdds:       totalOdds,
		ValidSelections: validSelections,
		PotentialPayout: bet.Stake.Mul(totalOdds),
	}, nil
}

func (e *OddsEngine) getMatchOdds(matchID string) (*domain.Match, error) {
	// Try cache first
	cacheKey := "match:" + matchID
	if cached, err := e.cache.Get(cacheKey); err == nil {
		if match, ok := cached.(*domain.Match); ok {
			return match, nil
		}
	}

	// Fallback to provider
	return e.oddsProvider.GetMatchOdds(context.Background(), matchID)
}

func (e *OddsEngine) findSelectionOdds(match *domain.Match, selection *domain.Selection) (decimal.Decimal, bool) {
	for _, market := range match.Markets {
		if market.Name == selection.MarketName {
			for _, outcome := range market.Outcomes {
				if outcome.Name == selection.OutcomeName {
					return outcome.Odds, true
				}
			}
		}
	}
	return decimal.Zero, false
}

// BetOddsResult contains odds calculation results
type BetOddsResult struct {
	TotalOdds       decimal.Decimal     `json:"total_odds"`
	ValidSelections []*domain.Selection `json:"valid_selections"`
	PotentialPayout decimal.Decimal     `json:"potential_payout"`
}

// Error types
var (
	ErrNoValidSelections = errors.New("no valid selections found")
)
