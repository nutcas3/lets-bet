package odds_test

import (
    "context"
    "errors"
    "testing"
    "time"

    "github.com/betting-platform/internal/core/domain"
    "github.com/betting-platform/internal/core/usecase/odds"
    "github.com/shopspring/decimal"
)

// Fakes
// We implement the interfaces with the minimum needed for tests.
// No real cache, no real provider

type fakeProvider struct {
    matchOdds map[string]*domain.Match
}

func (f *fakeProvider) GetLiveOdds(_ context.Context) ([]*domain.Match, error) {
    return nil, nil
}
func (f *fakeProvider) GetMatchOdds(_ context.Context, matchID string) (*domain.Match, error) {
    m, ok := f.matchOdds[matchID]
    if !ok {
        return nil, errors.New("match not found")
    }
    return m, nil
}
func (f *fakeProvider) CalculateParlayOdds(o []decimal.Decimal) decimal.Decimal {
    return decimal.Zero
}

type fakeCache struct{}

// Get always misses  forces every lookup to go to the provider.
func (f *fakeCache) Get(_ string) (any, error)               { return nil, errors.New("miss") }
func (f *fakeCache) Set(_ string, _ any, _ time.Duration) error { return nil }
func (f *fakeCache) Delete(_ string) error                      { return nil }

type fakeEventBus struct{}

func (f *fakeEventBus) Publish(_ string, _ any) error { return nil }

// Helper to build a match with one market and one outcome

func matchWithOdds(matchID, marketName, outcomeName string, oddsVal float64) *domain.Match {
    return &domain.Match{
        ID: matchID,
        Markets: []domain.Market{
            {
                Name: marketName,
                Outcomes: []domain.Outcome{
                    {Name: outcomeName, Odds: decimal.NewFromFloat(oddsVal)},
                },
            },
        },
    }
}

// Tests

// TestCalculateBetOdds_SingleSelection is the core regression test.
// Before the fix: totalOdds starts at 0, so 0 * 2.5 = 0. This test fails.
// After the fix:  totalOdds starts at 1, so 1 * 2.5 = 2.5. This test passes.
func TestCalculateBetOdds_SingleSelection(t *testing.T) {
    t.Parallel()

    provider := &fakeProvider{
        matchOdds: map[string]*domain.Match{
            "match-1": matchWithOdds("match-1", "Match Winner", "Arsenal", 2.5),
        },
    }
    engine := odds.NewOddsEngine(provider, &fakeCache{}, &fakeEventBus{})

    bet := &domain.Bet{
        Stake: decimal.NewFromInt(100),
        Selections: []domain.Selection{
            {EventID: "match-1", MarketName: "Match Winner", OutcomeName: "Arsenal"},
        },
    }

    result, err := engine.CalculateBetOdds(bet)
    if err != nil {
        t.Fatalf("CalculateBetOdds returned error: %v", err)
    }

    want := decimal.NewFromFloat(2.5)
    if !result.TotalOdds.Equal(want) {
        t.Errorf("TotalOdds = %s, want %s", result.TotalOdds, want)
    }

    wantPayout := decimal.NewFromFloat(250) // 100 * 2.5
    if !result.PotentialPayout.Equal(wantPayout) {
        t.Errorf("PotentialPayout = %s, want %s", result.PotentialPayout, wantPayout)
    }
}

// TestCalculateBetOdds_MultiSelection proves the accumulator multiplication works.
// Arsenal 2.5 × Chelsea 1.8 = 4.5
func TestCalculateBetOdds_MultiSelection(t *testing.T) {
    t.Parallel()

    provider := &fakeProvider{
        matchOdds: map[string]*domain.Match{
            "match-1": matchWithOdds("match-1", "Match Winner", "Arsenal", 2.5),
            "match-2": matchWithOdds("match-2", "Match Winner", "Chelsea", 1.8),
        },
    }
    engine := odds.NewOddsEngine(provider, &fakeCache{}, &fakeEventBus{})

    bet := &domain.Bet{
        Stake: decimal.NewFromInt(100),
        Selections: []domain.Selection{
            {EventID: "match-1", MarketName: "Match Winner", OutcomeName: "Arsenal"},
            {EventID: "match-2", MarketName: "Match Winner", OutcomeName: "Chelsea"},
        },
    }

    result, err := engine.CalculateBetOdds(bet)
    if err != nil {
        t.Fatalf("CalculateBetOdds returned error: %v", err)
    }

    want := decimal.NewFromFloat(4.5) // 2.5 * 1.8
    if !result.TotalOdds.Equal(want) {
        t.Errorf("TotalOdds = %s, want %s", result.TotalOdds, want)
    }
}

// TestCalculateBetOdds_NoValidSelections confirms the error path works.
func TestCalculateBetOdds_NoValidSelections(t *testing.T) {
    t.Parallel()

    provider := &fakeProvider{matchOdds: map[string]*domain.Match{}}
    engine := odds.NewOddsEngine(provider, &fakeCache{}, &fakeEventBus{})

    bet := &domain.Bet{
        Stake:      decimal.NewFromInt(100),
        Selections: []domain.Selection{
            {EventID: "nonexistent", MarketName: "Match Winner", OutcomeName: "Arsenal"},
        },
    }

    _, err := engine.CalculateBetOdds(bet)
    if !errors.Is(err, odds.ErrNoValidSelections) {
        t.Errorf("expected ErrNoValidSelections, got %v", err)
    }
}