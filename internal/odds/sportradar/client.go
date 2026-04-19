// Package sportradar provides Sportradar odds feed integration
package sportradar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
)

// SportradarConfig provides configuration for Sportradar client
type SportradarConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Environment string        `json:"environment"` // "trial", "production"
	Timeout     time.Duration `json:"timeout"`
	RateLimit   int           `json:"rate_limit"` // requests per minute
}

// DefaultSportradarConfig returns default configuration
func DefaultSportradarConfig() *SportradarConfig {
	return &SportradarConfig{
		Environment: "trial",
		BaseURL:     "https://api.sportradar.com",
		Timeout:     30 * time.Second,
		RateLimit:   60, // 60 requests per minute
	}
}

// SportradarClient provides Sportradar odds feed integration
type SportradarClient struct {
	config      *SportradarConfig
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewSportradarClient creates a new Sportradar client
func NewSportradarClient(config *SportradarConfig) *SportradarClient {
	if config == nil {
		config = DefaultSportradarConfig()
	}

	return &SportradarClient{
		config:      config,
		httpClient:  &http.Client{Timeout: config.Timeout},
		rateLimiter: NewRateLimiter(config.RateLimit, time.Minute),
	}
}

// Sport represents a sport in Sportradar
type Sport struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Tournament represents a tournament in Sportradar
type Tournament struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Sport    Sport  `json:"sport"`
	Category string `json:"category"`
}

// Match represents a match/event in Sportradar
type Match struct {
	ID          string     `json:"id"`
	Sport       Sport      `json:"sport"`
	Tournament  Tournament `json:"tournament"`
	HomeTeam    Team       `json:"home_team"`
	AwayTeam    Team       `json:"away_team"`
	Status      string     `json:"status"`
	ScheduledAt time.Time  `json:"scheduled_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Score       Score      `json:"score"`
	Odds        []Odds     `json:"odds"`
}

// Team represents a team in Sportradar
type Team struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

// Score represents match score
type Score struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// Odds represents betting odds for a match
type Odds struct {
	ID          string          `json:"id"`
	Market      string          `json:"market"` // "moneyline", "spread", "total"
	Outcome     string          `json:"outcome"`
	Price       decimal.Decimal `json:"price"`
	UpdatedAt   time.Time       `json:"updated_at"`
	IsAvailable bool            `json:"is_available"`
}

// SportradarOddsResponse represents Sportradar API response
type SportradarOddsResponse struct {
	Success bool    `json:"success"`
	Data    []Match `json:"data"`
	Error   string  `json:"error,omitempty"`
}

// GetSports retrieves available sports
func (s *SportradarClient) GetSports(ctx context.Context) ([]Sport, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sports", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var sports []Sport
	if err := json.NewDecoder(resp.Body).Decode(&sports); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return sports, nil
}

// GetTournaments retrieves tournaments for a sport
func (s *SportradarClient) GetTournaments(ctx context.Context, sportID string) ([]Tournament, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sports/%s/tournaments", s.config.BaseURL, sportID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var tournaments []Tournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tournaments, nil
}

// GetMatches retrieves matches for a tournament
func (s *SportradarClient) GetMatches(ctx context.Context, tournamentID string) ([]Match, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/tournaments/%s/matches", s.config.BaseURL, tournamentID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var response SportradarOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Sportradar API error: %s", response.Error)
	}

	return response.Data, nil
}

// GetLiveMatches retrieves live matches
func (s *SportradarClient) GetLiveMatches(ctx context.Context) ([]Match, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/live", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var response SportradarOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Sportradar API error: %s", response.Error)
	}

	return response.Data, nil
}

// GetMatchOdds retrieves odds for a specific match
func (s *SportradarClient) GetMatchOdds(ctx context.Context, matchID string) ([]Odds, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/%s/odds", s.config.BaseURL, matchID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var odds []Odds
	if err := json.NewDecoder(resp.Body).Decode(&odds); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return odds, nil
}

// GetUpcomingMatches retrieves upcoming matches
func (s *SportradarClient) GetUpcomingMatches(ctx context.Context, hours int) ([]Match, error) {
	// Wait for rate limiter
	if err := s.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/upcoming?hours=%d", s.config.BaseURL, hours)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", s.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Sportradar API error: %d", resp.StatusCode)
	}

	var response SportradarOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Sportradar API error: %s", response.Error)
	}

	return response.Data, nil
}

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
