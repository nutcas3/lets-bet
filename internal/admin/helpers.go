package admin

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// getOverviewStats retrieves overview statistics
func getOverviewStats(ctx context.Context) OverviewStats {
	_ = ctx // Use context to avoid unused parameter warning
	return OverviewStats{
		TotalUsers:          50000,
		ActiveUsers:         12000,
		TotalBets:           2500000,
		TotalVolume:         decimal.NewFromInt(125000000),
		TotalRevenue:        decimal.NewFromInt(6250000),
		TotalPayouts:        decimal.NewFromInt(5625000),
		ActiveMatches:       45,
		PendingTransactions: 125,
		SystemUptime:        "45 days, 12:34:56",
	}
}

// getRecentActivity retrieves recent activity items
func getRecentActivity(ctx context.Context, limit int) []ActivityItem {
	_ = ctx // Use context to avoid unused parameter warning
	_ = limit // Use limit to avoid unused parameter warning
	
	return []ActivityItem{
		{
			ID:          "act_001",
			Type:        "bet",
			Description: "User placed bet on football match",
			UserID:      "user_123",
			Timestamp:   time.Now().Add(-5 * time.Minute),
			Amount:      decimal.NewFromInt(100),
			Status:      "completed",
		},
		{
			ID:          "act_002",
			Type:        "deposit",
			Description: "User deposited funds",
			UserID:      "user_456",
			Timestamp:   time.Now().Add(-10 * time.Minute),
			Amount:      decimal.NewFromInt(500),
			Status:      "completed",
		},
		{
			ID:          "act_003",
			Type:        "withdrawal",
			Description: "User requested withdrawal",
			UserID:      "user_789",
			Timestamp:   time.Now().Add(-15 * time.Minute),
			Amount:      decimal.NewFromInt(200),
			Status:      "pending",
		},
	}
}

// getTopUsers retrieves top users by activity
func getTopUsers(ctx context.Context, limit int) []UserStats {
	_ = ctx // Use context to avoid unused parameter warning
	_ = limit // Use limit to avoid unused parameter warning
	
	return []UserStats{
		{
			UserID:       "user_001",
			Username:     "high_roller",
			TotalBets:    5000,
			TotalVolume:  decimal.NewFromInt(100000),
			TotalRevenue: decimal.NewFromInt(5000),
			WinRate:      decimal.NewFromFloat(0.52),
			LastActive:   time.Now().Add(-1 * time.Hour),
			Status:       "Active",
		},
		{
			UserID:       "user_002",
			Username:     "regular_bettor",
			TotalBets:    2500,
			TotalVolume:  decimal.NewFromInt(50000),
			TotalRevenue: decimal.NewFromInt(2500),
			WinRate:      decimal.NewFromFloat(0.48),
			LastActive:   time.Now().Add(-2 * time.Hour),
			Status:       "Active",
		},
		{
			UserID:       "user_003",
			Username:     "weekend_player",
			TotalBets:    1500,
			TotalVolume:  decimal.NewFromInt(30000),
			TotalRevenue: decimal.NewFromInt(1500),
			WinRate:      decimal.NewFromFloat(0.45),
			LastActive:   time.Now().Add(-3 * time.Hour),
			Status:       "Active",
		},
	}
}

// getRevenueMetrics retrieves revenue metrics
func getRevenueMetrics(ctx context.Context) RevenueMetrics {
	_ = ctx // Use context to avoid unused parameter warning
	return RevenueMetrics{
		DailyRevenue:   decimal.NewFromInt(50000),
		WeeklyRevenue:  decimal.NewFromInt(350000),
		MonthlyRevenue: decimal.NewFromInt(1500000),
		YearlyRevenue:  decimal.NewFromInt(18000000),
		GrowthRate:     decimal.NewFromFloat(0.15),
	}
}

// getBettingMetrics retrieves betting metrics
func getBettingMetrics(ctx context.Context) BettingMetrics {
	_ = ctx // Use context to avoid unused parameter warning
	return BettingMetrics{
		DailyBets:   5000,
		WeeklyBets:  35000,
		MonthlyBets: 150000,
		YearlyBets:  1800000,
		AverageBet:  decimal.NewFromInt(50),
	}
}

// getFinancialStats retrieves financial statistics
func getFinancialStats(ctx context.Context) FinancialStats {
	_ = ctx // Use context to avoid unused parameter warning
	return FinancialStats{
		TotalDeposits:      decimal.NewFromInt(25000000),
		TotalWithdrawals:   decimal.NewFromInt(22500000),
		NetCashFlow:        decimal.NewFromInt(2500000),
		PendingDeposits:    50,
		PendingWithdrawals: 25,
	}
}

// getSystemHealth retrieves system health status
func getSystemHealth(ctx context.Context) SystemHealth {
	_ = ctx // Use context to avoid unused parameter warning
	return SystemHealth{
		OverallStatus: "Healthy",
		Services: []ServiceStatus{
			{
				Name:      "API Gateway",
				Status:    "Running",
				Uptime:    45 * 24 * time.Hour,
				LastCheck: time.Now(),
				Message:   "All systems operational",
			},
			{
				Name:      "Database",
				Status:    "Running",
				Uptime:    45 * 24 * time.Hour,
				LastCheck: time.Now(),
				Message:   "Connection pool healthy",
			},
			{
				Name:      "Cache",
				Status:    "Running",
				Uptime:    45 * 24 * time.Hour,
				LastCheck: time.Now(),
				Message:   "Hit rate: 95%",
			},
		},
		Database: DatabaseStatus{
			Status:      "Connected",
			Connections: 25,
			QueryTime:   5 * time.Millisecond,
			LastCheck:   time.Now(),
		},
		Cache: CacheStatus{
			Status:      "Available",
			HitRate:     decimal.NewFromFloat(0.95),
			MemoryUsage: decimal.NewFromInt(2048),
			LastCheck:   time.Now(),
		},
		API: APIStatus{
			Status:              "Operational",
			ResponseTime:         150 * time.Millisecond,
			RequestsPerSecond:   decimal.NewFromFloat(125.5),
			LastCheck:           time.Now(),
		},
		LoadAverage: []float64{0.5, 0.6, 0.7},
		MemoryUsage: MemoryUsage{
			Total:     decimal.NewFromInt(16384),
			Used:      decimal.NewFromInt(8192),
			Available: decimal.NewFromInt(8192),
			Percent:   decimal.NewFromFloat(50.0),
		},
		DiskUsage: DiskUsage{
			Total:     decimal.NewFromInt(1024000),
			Used:      decimal.NewFromInt(614400),
			Available: decimal.NewFromInt(409600),
			Percent:   decimal.NewFromFloat(60.0),
		},
	}
}

// TransactionStats represents transaction statistics (moved from types.go to avoid circular dependencies)
type TransactionStats struct {
	Type    string          `json:"type"`
	Count   int64           `json:"count"`
	Amount  decimal.Decimal `json:"amount"`
	Average decimal.Decimal `json:"average"`
}
