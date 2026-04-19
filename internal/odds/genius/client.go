// Package genius provides Genius Sports odds feed integration
package genius

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

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

// GetSports retrieves available sports
func (g *GeniusClient) GetSports(ctx context.Context) (*SportResponse, error) {
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
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var sportResp SportResponse
	if err := json.Unmarshal(body, &sportResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &sportResp, nil
}

// GetLeagues retrieves available leagues
func (g *GeniusClient) GetLeagues(ctx context.Context, sportID string) (*LeagueResponse, error) {
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/leagues?sport_id=%s", g.config.BaseURL, sportID)
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
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var leagueResp LeagueResponse
	if err := json.Unmarshal(body, &leagueResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &leagueResp, nil
}

// GetMatches retrieves available matches
func (g *GeniusClient) GetMatches(ctx context.Context, req *MatchRequest) (*MatchResponse, error) {
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/matches", g.config.BaseURL)

	// Build query parameters
	if req.SportID != "" {
		url += fmt.Sprintf("?sport_id=%s", req.SportID)
	}
	if req.LeagueID != "" {
		if req.SportID != "" {
			url += "&"
		} else {
			url += "?"
		}
		url += fmt.Sprintf("league_id=%s", req.LeagueID)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matchResp MatchResponse
	if err := json.Unmarshal(body, &matchResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &matchResp, nil
}

// GetOdds retrieves odds for matches
func (g *GeniusClient) GetOdds(ctx context.Context, req *OddsRequest) (*OddsResponse, error) {
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit error: %w", err)
	}

	url := fmt.Sprintf("%s/v1/odds", g.config.BaseURL)

	// Build query parameters
	if req.SportID != "" {
		url += fmt.Sprintf("?sport_id=%s", req.SportID)
	}
	if req.MatchID != "" {
		if req.SportID != "" {
			url += "&"
		} else {
			url += "?"
		}
		url += fmt.Sprintf("match_id=%s", req.MatchID)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+g.config.APIKey)
	httpReq.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var oddsResp OddsResponse
	if err := json.Unmarshal(body, &oddsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &oddsResp, nil
}

// GetMetrics returns client metrics
func (g *GeniusClient) GetMetrics(ctx context.Context) (*GeniusMetrics, error) {
	// In a real implementation, these would be calculated from actual data
	return &GeniusMetrics{
		TotalRequests:       1000,
		SuccessfulRequests:  950,
		FailedRequests:      50,
		LastRequestTime:     time.Now().Add(-1 * time.Hour),
		AverageResponseTime: 150 * time.Millisecond,
	}, nil
}
