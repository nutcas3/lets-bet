// Package genius provides Genius Sports odds feed integration
package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
)

// GeniusConfig provides configuration for Genius Sports client
type GeniusConfig struct {
	APIKey      string        `json:"api_key"`
	BaseURL     string        `json:"base_url"`
	Environment string        `json:"environment"` // "trial", "production"
	Timeout     time.Duration `json:"timeout"`
	RateLimit   int           `json:"rate_limit"` // requests per minute
}

// DefaultGeniusConfig returns default configuration
func DefaultGeniusConfig() *GeniusConfig {
	return &GeniusConfig{
		Environment: "trial",
		BaseURL:     "https://api.geniussports.com",
		Timeout:     30 * time.Second,
		RateLimit:   60, // 60 requests per minute
	}
}

// GeniusClient provides Genius Sports odds feed integration
type GeniusClient struct {
	config      *GeniusConfig
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewGeniusClient creates a new Genius Sports client
func NewGeniusClient(config *GeniusConfig) *GeniusClient {
	if config == nil {
		config = DefaultGeniusConfig()
	}

	return &GeniusClient{
		config:      config,
		httpClient:  &http.Client{Timeout: config.Timeout},
		rateLimiter: NewRateLimiter(config.RateLimit, time.Minute),
	}
}

// GeniusSport represents a sport in Genius Sports
type GeniusSport struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GeniusTournament represents a tournament in Genius Sports
type GeniusTournament struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Sport    GeniusSport `json:"sport"`
	Category string      `json:"category"`
	Season   string      `json:"season"`
}

// GeniusMatch represents a match/event in Genius Sports
type GeniusMatch struct {
	ID          string           `json:"id"`
	Sport       GeniusSport      `json:"sport"`
	Tournament  GeniusTournament `json:"tournament"`
	HomeTeam    GeniusTeam       `json:"home_team"`
	AwayTeam    GeniusTeam       `json:"away_team"`
	Status      string           `json:"status"`
	ScheduledAt time.Time        `json:"scheduled_at"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
	Score       GeniusScore      `json:"score"`
	Odds        []GeniusOdds     `json:"odds"`
	Meta        map[string]any   `json:"meta"`
}

// GeniusTeam represents a team in Genius Sports
type GeniusTeam struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Country      string `json:"country"`
	Abbreviation string `json:"abbreviation"`
}

// GeniusScore represents match score
type GeniusScore struct {
	Home    int                 `json:"home"`
	Away    int                 `json:"away"`
	Periods []GeniusScorePeriod `json:"periods,omitempty"`
}

// GeniusScorePeriod represents score for a specific period
type GeniusScorePeriod struct {
	Number int    `json:"number"`
	Home   int    `json:"home"`
	Away   int    `json:"away"`
	Type   string `json:"type"` // "quarter", "half", "set", etc.
}

// GeniusOdds represents betting odds for a match
type GeniusOdds struct {
	ID          string          `json:"id"`
	Market      string          `json:"market"`
	Outcome     string          `json:"outcome"`
	Price       decimal.Decimal `json:"price"`
	UpdatedAt   time.Time       `json:"updated_at"`
	IsAvailable bool            `json:"is_available"`
	Liquidity   decimal.Decimal `json:"liquidity"`
	Volume      decimal.Decimal `json:"volume"`
}

// GeniusOddsResponse represents Genius Sports API response
type GeniusOddsResponse struct {
	Success bool          `json:"success"`
	Data    []GeniusMatch `json:"data"`
	Error   string        `json:"error,omitempty"`
	Meta    GeniusMeta    `json:"meta"`
}

// GeniusMeta represents response metadata
type GeniusMeta struct {
	Total       int       `json:"total"`
	Page        int       `json:"page"`
	PerPage     int       `json:"per_page"`
	TotalPages  int       `json:"total_pages"`
	LastUpdated time.Time `json:"last_updated"`
}

// RateLimiter provides rate limiting for API requests
type RateLimiter struct {
	tokens     int
	maxTokens  int
	interval   time.Duration
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		interval:   interval,
		lastRefill: time.Now(),
	}
}

// Wait waits until a token is available
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		if r.tryConsume() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.interval / time.Duration(r.maxTokens)):
			// Wait for a short period before retrying
		}
	}
}

// tryConsume tries to consume a token
func (r *RateLimiter) tryConsume() bool {
	now := time.Now()
	// Refill tokens based on time elapsed
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed / r.interval)
	if tokensToAdd > 0 {
		r.tokens = min(r.maxTokens, r.tokens+tokensToAdd)
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}

	return false
}

// GetSports retrieves available sports
func (g *GeniusClient) GetSports(ctx context.Context) ([]GeniusSport, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sports", g.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var sports []GeniusSport
	if err := json.NewDecoder(resp.Body).Decode(&sports); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return sports, nil
}

// GetTournaments retrieves tournaments for a sport
func (g *GeniusClient) GetTournaments(ctx context.Context, sportID string) ([]GeniusTournament, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sports/%s/tournaments", g.config.BaseURL, sportID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var tournaments []GeniusTournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tournaments, nil
}

// GetMatches retrieves matches for a tournament
func (g *GeniusClient) GetMatches(ctx context.Context, tournamentID string, page, perPage int) (*GeniusOddsResponse, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/tournaments/%s/matches?page=%d&per_page=%d",
		g.config.BaseURL, tournamentID, page, perPage)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var response GeniusOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Genius Sports API error: %s", response.Error)
	}

	return &response, nil
}

// GetLiveMatches retrieves live matches
func (g *GeniusClient) GetLiveMatches(ctx context.Context) (*GeniusOddsResponse, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/live", g.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var response GeniusOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Genius Sports API error: %s", response.Error)
	}

	return &response, nil
}

// GetMatchOdds retrieves odds for a specific match
func (g *GeniusClient) GetMatchOdds(ctx context.Context, matchID string) ([]GeniusOdds, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/%s/odds", g.config.BaseURL, matchID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var odds []GeniusOdds
	if err := json.NewDecoder(resp.Body).Decode(&odds); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return odds, nil
}

// GetUpcomingMatches retrieves upcoming matches
func (g *GeniusClient) GetUpcomingMatches(ctx context.Context, hours int, page, perPage int) (*GeniusOddsResponse, error) {
	// Wait for rate limiter
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches/upcoming?hours=%d&page=%d&per_page=%d",
		g.config.BaseURL, hours, page, perPage)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Genius Sports API error: %d", resp.StatusCode)
	}

	var response GeniusOddsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("Genius Sports API error: %s", response.Error)
	}

	return &response, nil
}

// ConvertToDomainMatch converts Genius Sports match to domain match
func (g *GeniusClient) ConvertToDomainMatch(match GeniusMatch) *domain.Match {
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

// ConvertToDomainOdds converts Genius Sports odds to domain markets
func (g *GeniusClient) ConvertToDomainOdds(odds []GeniusOdds, matchID string) []domain.Market {
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
