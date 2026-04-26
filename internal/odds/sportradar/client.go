package sportradar

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// DefaultSportradarConfig returns default configuration
func DefaultSportradarConfig() *SportradarConfig {
	return &SportradarConfig{
		Environment: "trial",
		BaseURL:     "https://api.sportradar.com",
		Timeout:     30 * time.Second,
		RateLimit:   60, // 60 requests per minute
	}
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
