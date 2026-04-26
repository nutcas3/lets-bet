package sportradar

import (
	"github.com/betting-platform/internal/core/domain"
)

// ConvertToDomainMatch converts Sportradar match to domain match
func (s *SportradarClient) ConvertToDomainMatch(match Match) *domain.Match {
	domainMatch := &domain.Match{
		ID:        match.ID,
		Sport:     domain.Sport(match.Sport.Name),
		League:    match.Tournament.Name,
		HomeTeam:  match.HomeTeam.Name,
		AwayTeam:  match.AwayTeam.Name,
		StartTime: match.ScheduledAt,
		Status:    domain.MatchStatus(match.Status),
		Score: &domain.MatchScore{
			HomeScore: match.Score.Home,
			AwayScore: match.Score.Away,
		},
		Markets: make([]domain.Market, 0),
	}

	if match.StartedAt != nil {
		// Note: domain.Match doesn't have StartedAt field
	}

	if match.CompletedAt != nil {
		// Note: domain.Match doesn't have CompletedAt field
	}

	// Convert odds to markets
	for _, odds := range match.Odds {
		// Create market with outcome
		market := domain.Market{
			ID:      odds.ID,
			MatchID: match.ID,
			Type:    domain.MarketType(odds.Market),
			Name:    odds.Market,
			Outcomes: []domain.Outcome{
				{
					ID:       odds.ID + "_" + odds.Outcome,
					MarketID: odds.ID,
					Name:     odds.Outcome,
					Odds:     odds.Price,
					Price:    odds.Price,
					Status:   domain.OutcomeStatusPending,
				},
			},
			Status: domain.MarketStatusOpen,
		}
		if !odds.IsAvailable {
			market.Status = domain.MarketStatusSuspended
		}
		domainMatch.Markets = append(domainMatch.Markets, market)
	}

	return domainMatch
}

// ConvertToDomainOdds converts Sportradar odds to domain markets
func (s *SportradarClient) ConvertToDomainOdds(odds []Odds, matchID string) []domain.Market {
	domainMarkets := make([]domain.Market, len(odds))
	for i, odd := range odds {
		domainMarkets[i] = domain.Market{
			ID:      odd.ID,
			MatchID: matchID,
			Type:    domain.MarketType(odd.Market),
			Name:    odd.Market,
			Outcomes: []domain.Outcome{
				{
					ID:       odd.ID + "_" + odd.Outcome,
					MarketID: odd.ID,
					Name:     odd.Outcome,
					Odds:     odd.Price,
					Price:    odd.Price,
					Status:   domain.OutcomeStatusPending,
				},
			},
			Status: domain.MarketStatusOpen,
		}
		if !odd.IsAvailable {
			domainMarkets[i].Status = domain.MarketStatusSuspended
		}
	}
	return domainMarkets
}
