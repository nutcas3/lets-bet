package games

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GameAnalytics represents game analytics data
type GameAnalytics struct {
	TimeRange          time.Time       `json:"time_range"`
	GamesPlayed        int64           `json:"games_played"`
	UniquePlayers      int64           `json:"unique_players"`
	TotalVolume        decimal.Decimal `json:"total_volume"`
	TotalRevenue       decimal.Decimal `json:"total_revenue"`
	AverageSessionTime time.Duration   `json:"average_session_time"`
	PlayerRetention    decimal.Decimal `json:"player_retention"`
	PopularTimes       []TimeSlot      `json:"popular_times"`
	OddsDistribution   []OddsRange     `json:"odds_distribution"`
	BetPatterns        []BetPattern    `json:"bet_patterns"`
}

// TimeSlot represents a time slot with activity data
type TimeSlot struct {
	Hour    int             `json:"hour"`
	Games   int64           `json:"games"`
	Players int64           `json:"players"`
	Volume  decimal.Decimal `json:"volume"`
}

// OddsRange represents an odds range with statistics
type OddsRange struct {
	MinOdds   decimal.Decimal `json:"min_odds"`
	MaxOdds   decimal.Decimal `json:"max_odds"`
	GameCount int64           `json:"game_count"`
	Frequency decimal.Decimal `json:"frequency"`
}

// BetPattern represents betting patterns
type BetPattern struct {
	Pattern    string          `json:"pattern"`
	Count      int64           `json:"count"`
	Percentage decimal.Decimal `json:"percentage"`
	AvgAmount  decimal.Decimal `json:"avg_amount"`
}

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Rank        int             `json:"rank"`
	UserID      uuid.UUID       `json:"user_id"`
	Username    string          `json:"username"`
	TotalProfit decimal.Decimal `json:"total_profit"`
	WinRate     decimal.Decimal `json:"win_rate"`
	GamesPlayed int64           `json:"games_played"`
	LastPlayed  time.Time       `json:"last_played"`
}

// GameLeaderboard represents the game leaderboard
type GameLeaderboard struct {
	Period       string              `json:"period"` // "daily", "weekly", "monthly", "all_time"
	UpdatedAt    time.Time           `json:"updated_at"`
	Entries      []*LeaderboardEntry `json:"entries"`
	TotalPlayers int64               `json:"total_players"`
}
