package admin

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
)

// AdminService provides administrative operations for the betting platform
type AdminService struct {
	eventBus EventBus
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// NewAdminService creates a new admin service
func NewAdminService(eventBus EventBus) *AdminService {
	return &AdminService{
		eventBus: eventBus,
	}
}

// DashboardData represents the main dashboard overview
type DashboardData struct {
	Overview        OverviewStats       `json:"overview"`
	RecentActivity  []ActivityItem     `json:"recent_activity"`
	TopUsers        []UserStats        `json:"top_users"`
	RevenueMetrics  RevenueMetrics     `json:"revenue_metrics"`
	BettingMetrics  BettingMetrics     `json:"betting_metrics"`
	FinancialStats  FinancialStats     `json:"financial_stats"`
	SystemHealth    SystemHealth       `json:"system_health"`
}

// OverviewStats represents high-level platform statistics
type OverviewStats struct {
	TotalUsers           int64             `json:"total_users"`
	ActiveUsers          int64             `json:"active_users"`
	TotalBets            int64             `json:"total_bets"`
	TotalVolume          decimal.Decimal   `json:"total_volume"`
	TotalRevenue         decimal.Decimal   `json:"total_revenue"`
	TotalPayouts         decimal.Decimal   `json:"total_payouts"`
	ActiveMatches        int64             `json:"active_matches"`
	PendingTransactions  int64             `json:"pending_transactions"`
	SystemUptime         string            `json:"system_uptime"`
	LastUpdated          time.Time         `json:"last_updated"`
}

// ActivityItem represents a recent system activity
type ActivityItem struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	UserID    string          `json:"user_id,omitempty"`
	Username  string          `json:"username,omitempty"`
	Action    string          `json:"action"`
	Details   string          `json:"details"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Status    string          `json:"status"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID       string          `json:"user_id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	TotalBets    int64           `json:"total_bets"`
	TotalVolume  decimal.Decimal `json:"total_volume"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	WinRate      decimal.Decimal `json:"win_rate"`
	LastActive   time.Time       `json:"last_active"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}

// RevenueMetrics represents revenue analytics
type RevenueMetrics struct {
	TodayRevenue     decimal.Decimal     `json:"today_revenue"`
	WeekRevenue      decimal.Decimal     `json:"week_revenue"`
	MonthRevenue     decimal.Decimal     `json:"month_revenue"`
	YearRevenue      decimal.Decimal     `json:"year_revenue"`
	RevenueByPeriod  []RevenuePeriod    `json:"revenue_by_period"`
	RevenueBySport   []RevenueBySport   `json:"revenue_by_sport"`
	RevenueBySource  []RevenueBySource  `json:"revenue_by_source"`
}

// RevenuePeriod represents revenue over a time period
type RevenuePeriod struct {
	Period   string          `json:"period"`
	Revenue  decimal.Decimal `json:"revenue"`
	Bets     int64           `json:"bets"`
	Users    int64           `json:"users"`
}

// RevenueBySport represents revenue breakdown by sport
type RevenueBySport struct {
	Sport    string          `json:"sport"`
	Revenue  decimal.Decimal `json:"revenue"`
	Bets     int64           `json:"bets"`
	Users    int64           `json:"users"`
	Growth   decimal.Decimal `json:"growth"`
}

// RevenueBySource represents revenue breakdown by source
type RevenueBySource struct {
	Source   string          `json:"source"`
	Revenue  decimal.Decimal `json:"revenue"`
	Count    int64           `json:"count"`
	Percent  decimal.Decimal `json:"percent"`
}

// BettingMetrics represents betting analytics
type BettingMetrics struct {
	TotalBets         int64             `json:"total_bets"`
	WinningBets       int64             `json:"winning_bets"`
	LosingBets        int64             `json:"losing_bets"`
	VoidBets          int64             `json:"void_bets"`
	TotalVolume       decimal.Decimal   `json:"total_volume"`
	AverageBetSize    decimal.Decimal   `json:"average_bet_size"`
	LargestBet        decimal.Decimal   `json:"largest_bet"`
	SmallestBet       decimal.Decimal   `json:"smallest_bet"`
	WinRate           decimal.Decimal   `json:"win_rate"`
	HoldPercentage    decimal.Decimal   `json:"hold_percentage"`
	BetsBySport       []BetsBySport     `json:"bets_by_sport"`
	BetsByHour        []BetsByHour      `json:"bets_by_hour"`
}

// BetsBySport represents betting metrics by sport
type BetsBySport struct {
	Sport       string          `json:"sport"`
	Bets        int64           `json:"bets"`
	Volume      decimal.Decimal `json:"volume"`
	Revenue     decimal.Decimal `json:"revenue"`
	WinRate     decimal.Decimal `json:"win_rate"`
}

// BetsByHour represents betting metrics by hour
type BetsByHour struct {
	Hour   int             `json:"hour"`
	Bets   int64           `json:"bets"`
	Volume decimal.Decimal `json:"volume"`
}

// FinancialStats represents financial statistics
type FinancialStats struct {
	TotalDeposits      decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals   decimal.Decimal `json:"total_withdrawals"`
	TotalBalance       decimal.Decimal `json:"total_balance"`
	PendingDeposits    int64           `json:"pending_deposits"`
	PendingWithdrawals int64           `json:"pending_withdrawals"`
	WalletBalances     []WalletBalance `json:"wallet_balances"`
	TransactionStats   []TransactionStats `json:"transaction_stats"`
}

// WalletBalance represents wallet balance statistics
type WalletBalance struct {
	Currency string          `json:"currency"`
	Total    decimal.Decimal `json:"total"`
	Count    int64           `json:"count"`
	Average  decimal.Decimal `json:"average"`
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	Type     string          `json:"type"`
	Count    int64           `json:"count"`
	Amount   decimal.Decimal `json:"amount"`
	Average  decimal.Decimal `json:"average"`
}

// SystemHealth represents system health metrics
type SystemHealth struct {
	DatabaseStatus    string          `json:"database_status"`
	EventBusStatus    string          `json:"event_bus_status"`
	PaymentGateways   []GatewayStatus `json:"payment_gateways"`
	OddsProviders     []ProviderStatus `json:"odds_providers"`
	ResponseTime      time.Duration   `json:"response_time"`
	ErrorRate         decimal.Decimal `json:"error_rate"`
	CPUUsage          decimal.Decimal `json:"cpu_usage"`
	MemoryUsage       decimal.Decimal `json:"memory_usage"`
	DiskUsage         decimal.Decimal `json:"disk_usage"`
	LastHealthCheck   time.Time       `json:"last_health_check"`
}

// GatewayStatus represents payment gateway status
type GatewayStatus struct {
	Name      string          `json:"name"`
	Status    string          `json:"status"`
	LastCheck time.Time       `json:"last_check"`
	Latency   time.Duration   `json:"latency"`
}

// ProviderStatus represents odds provider status
type ProviderStatus struct {
	Name      string          `json:"name"`
	Status    string          `json:"status"`
	LastCheck time.Time       `json:"last_check"`
	Latency   time.Duration   `json:"latency"`
}

// GetDashboardData retrieves comprehensive dashboard data
func (s *AdminService) GetDashboardData(ctx context.Context) (*DashboardData, error) {
	overview := s.getOverviewStats(ctx)
	recentActivity := s.getRecentActivity(ctx, 50)
	topUsers := s.getTopUsers(ctx, 10)
	revenueMetrics := s.getRevenueMetrics(ctx)
	bettingMetrics := s.getBettingMetrics(ctx)
	financialStats := s.getFinancialStats(ctx)
	systemHealth := s.getSystemHealth(ctx)

	return &DashboardData{
		Overview:       *overview,
		RecentActivity: recentActivity,
		TopUsers:       topUsers,
		RevenueMetrics: *revenueMetrics,
		BettingMetrics: *bettingMetrics,
		FinancialStats: *financialStats,
		SystemHealth:   *systemHealth,
	}, nil
}

// getOverviewStats retrieves high-level platform statistics
func (s *AdminService) getOverviewStats(ctx context.Context) *OverviewStats {
	return &OverviewStats{
		TotalUsers:          10000,
		ActiveUsers:         2500,
		TotalBets:           50000,
		TotalVolume:         decimal.NewFromFloat(1000000),
		TotalRevenue:        decimal.NewFromFloat(50000),
		TotalPayouts:        decimal.NewFromFloat(45000),
		ActiveMatches:       25,
		PendingTransactions: 150,
		SystemUptime:        "30d 14h 23m",
		LastUpdated:         time.Now(),
	}
}

// getRecentActivity retrieves recent system activity
func (s *AdminService) getRecentActivity(ctx context.Context, limit int) []ActivityItem {
	var activities []ActivityItem

	for i := 0; i < limit && i < 10; i++ {
		activities = append(activities, ActivityItem{
			ID:        fmt.Sprintf("activity_%d", i),
			Type:      "bet",
			UserID:    fmt.Sprintf("user_%d", i),
			Username:  fmt.Sprintf("user%d", i),
			Action:    "bet_placed",
			Details:   fmt.Sprintf("Bet on match %d", i),
			Amount:    decimal.NewFromFloat(float64(100 + i*10)),
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
			Status:    "completed",
		})
	}

	return activities
}

// getTopUsers retrieves top users by various metrics
func (s *AdminService) getTopUsers(ctx context.Context, limit int) []UserStats {
	var userStats []UserStats

	for i := 0; i < limit && i < 10; i++ {
		userStats = append(userStats, UserStats{
			UserID:       fmt.Sprintf("user_%d", i),
			Username:     fmt.Sprintf("user%d", i),
			Email:        fmt.Sprintf("user%d@example.com", i),
			TotalBets:    int64(100 + i*10),
			TotalVolume:  decimal.NewFromFloat(float64(1000 + i*100)),
			TotalRevenue: decimal.NewFromFloat(float64(50 + i*5)),
			WinRate:      decimal.NewFromFloat(0.45 + float64(i)*0.01),
			LastActive:   time.Now().Add(time.Duration(-i) * time.Hour),
			Status:       "active",
			CreatedAt:    time.Now().AddDate(0, 0, -i*30),
		})
	}

	return userStats
}

// getRevenueMetrics retrieves revenue analytics
func (s *AdminService) getRevenueMetrics(ctx context.Context) *RevenueMetrics {
	return &RevenueMetrics{
		TodayRevenue:   decimal.NewFromFloat(1500),
		WeekRevenue:    decimal.NewFromFloat(10500),
		MonthRevenue:   decimal.NewFromFloat(45000),
		YearRevenue:    decimal.NewFromFloat(500000),
		RevenueByPeriod: s.getRevenueByPeriod(30),
		RevenueBySport:  s.getRevenueBySport(),
		RevenueBySource: s.getRevenueBySource(),
	}
}

// getBettingMetrics retrieves betting analytics
func (s *AdminService) getBettingMetrics(ctx context.Context) *BettingMetrics {
	return &BettingMetrics{
		TotalBets:       50000,
		WinningBets:     22500,
		LosingBets:      25000,
		VoidBets:        2500,
		TotalVolume:     decimal.NewFromFloat(1000000),
		AverageBetSize:  decimal.NewFromFloat(20),
		LargestBet:      decimal.NewFromFloat(10000),
		SmallestBet:     decimal.NewFromFloat(1),
		WinRate:         decimal.NewFromFloat(0.45),
		HoldPercentage:  decimal.NewFromFloat(5.0),
		BetsBySport:     s.getBetsBySport(),
		BetsByHour:      s.getBetsByHour(24),
	}
}

// getFinancialStats retrieves financial statistics
func (s *AdminService) getFinancialStats(ctx context.Context) *FinancialStats {
	return &FinancialStats{
		TotalDeposits:      decimal.NewFromFloat(200000),
		TotalWithdrawals:   decimal.NewFromFloat(150000),
		TotalBalance:       decimal.NewFromFloat(50000),
		PendingDeposits:    50,
		PendingWithdrawals: 25,
		WalletBalances:     s.getWalletBalances(),
		TransactionStats:   s.getTransactionStats(),
	}
}

// getSystemHealth retrieves system health metrics
func (s *AdminService) getSystemHealth(ctx context.Context) *SystemHealth {
	return &SystemHealth{
		DatabaseStatus:  "healthy",
		EventBusStatus:  "healthy",
		PaymentGateways: s.getPaymentGatewayStatus(),
		OddsProviders:   s.getOddsProviderStatus(),
		ResponseTime:    50 * time.Millisecond,
		ErrorRate:       decimal.NewFromFloat(0.01),
		CPUUsage:        decimal.NewFromFloat(25.5),
		MemoryUsage:     decimal.NewFromFloat(45.2),
		DiskUsage:       decimal.NewFromFloat(60.8),
		LastHealthCheck: time.Now(),
	}
}

// Helper methods for generating sample data

func (s *AdminService) getRevenueByPeriod(days int) []RevenuePeriod {
	var periods []RevenuePeriod
	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		period := RevenuePeriod{
			Period:  date.Format("2006-01-02"),
			Revenue: decimal.NewFromFloat(float64(1000 + i*100)),
			Bets:    int64(100 + i*10),
			Users:   int64(50 + i*5),
		}
		periods = append(periods, period)
	}
	return periods
}

func (s *AdminService) getRevenueBySport() []RevenueBySport {
	return []RevenueBySport{
		{Sport: "Football", Revenue: decimal.NewFromFloat(50000), Bets: 5000, Users: 1000, Growth: decimal.NewFromFloat(15.5)},
		{Sport: "Basketball", Revenue: decimal.NewFromFloat(30000), Bets: 3000, Users: 600, Growth: decimal.NewFromFloat(8.2)},
		{Sport: "Tennis", Revenue: decimal.NewFromFloat(20000), Bets: 2000, Users: 400, Growth: decimal.NewFromFloat(12.1)},
	}
}

func (s *AdminService) getRevenueBySource() []RevenueBySource {
	return []RevenueBySource{
		{Source: "Sports Betting", Revenue: decimal.NewFromFloat(80000), Count: 8000, Percent: decimal.NewFromFloat(80)},
		{Source: "Jackpots", Revenue: decimal.NewFromFloat(15000), Count: 1500, Percent: decimal.NewFromFloat(15)},
		{Source: "Virtual Sports", Revenue: decimal.NewFromFloat(5000), Count: 500, Percent: decimal.NewFromFloat(5)},
	}
}

func (s *AdminService) getBetsBySport() []BetsBySport {
	return []BetsBySport{
		{Sport: "Football", Bets: 5000, Volume: decimal.NewFromFloat(100000), Revenue: decimal.NewFromFloat(50000), WinRate: decimal.NewFromFloat(0.45)},
		{Sport: "Basketball", Bets: 3000, Volume: decimal.NewFromFloat(60000), Revenue: decimal.NewFromFloat(30000), WinRate: decimal.NewFromFloat(0.48)},
		{Sport: "Tennis", Bets: 2000, Volume: decimal.NewFromFloat(40000), Revenue: decimal.NewFromFloat(20000), WinRate: decimal.NewFromFloat(0.42)},
	}
}

func (s *AdminService) getBetsByHour(hours int) []BetsByHour {
	var byHour []BetsByHour
	for i := 0; i < hours; i++ {
		hour := BetsByHour{
			Hour:   i,
			Bets:   int64(100 + i*10),
			Volume: decimal.NewFromFloat(float64(1000 + i*100)),
		}
		byHour = append(byHour, hour)
	}
	return byHour
}

func (s *AdminService) getWalletBalances() []WalletBalance {
	return []WalletBalance{
		{Currency: "KES", Total: decimal.NewFromFloat(1000000), Count: 10000, Average: decimal.NewFromFloat(100)},
		{Currency: "USD", Total: decimal.NewFromFloat(50000), Count: 500, Average: decimal.NewFromFloat(100)},
	}
}

func (s *AdminService) getTransactionStats() []TransactionStats {
	return []TransactionStats{
		{Type: "DEPOSIT", Count: 1000, Amount: decimal.NewFromFloat(100000), Average: decimal.NewFromFloat(100)},
		{Type: "WITHDRAWAL", Count: 500, Amount: decimal.NewFromFloat(50000), Average: decimal.NewFromFloat(100)},
		{Type: "BET_PLACED", Count: 10000, Amount: decimal.NewFromFloat(1000000), Average: decimal.NewFromFloat(100)},
		{Type: "BET_WON", Count: 4500, Amount: decimal.NewFromFloat(450000), Average: decimal.NewFromFloat(100)},
	}
}

func (s *AdminService) getPaymentGatewayStatus() []GatewayStatus {
	return []GatewayStatus{
		{Name: "M-Pesa", Status: "healthy", LastCheck: time.Now(), Latency: 150 * time.Millisecond},
		{Name: "Flutterwave", Status: "healthy", LastCheck: time.Now(), Latency: 200 * time.Millisecond},
	}
}

func (s *AdminService) getOddsProviderStatus() []ProviderStatus {
	return []ProviderStatus{
		{Name: "Sportradar", Status: "healthy", LastCheck: time.Now(), Latency: 100 * time.Millisecond},
		{Name: "Genius Sports", Status: "healthy", LastCheck: time.Now(), Latency: 120 * time.Millisecond},
	}
}

// UserManagement represents user management operations
type UserManagement struct {
	UserID    string          `json:"user_id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Status    string          `json:"status"`
	Role      string          `json:"role"`
	Balance   decimal.Decimal `json:"balance"`
	CreatedAt time.Time       `json:"created_at"`
	LastLogin time.Time       `json:"last_login"`
	KYCStatus string          `json:"kyc_status"`
	Limits    UserLimits      `json:"limits"`
}

// UserLimits represents user betting limits
type UserLimits struct {
	DailyLimit   decimal.Decimal `json:"daily_limit"`
	WeeklyLimit  decimal.Decimal `json:"weekly_limit"`
	MonthlyLimit decimal.Decimal `json:"monthly_limit"`
	MaxBetSize   decimal.Decimal `json:"max_bet_size"`
}

// GetUserManagementData retrieves user management data
func (s *AdminService) GetUserManagementData(ctx context.Context, limit, offset int) ([]UserManagement, error) {
	var users []UserManagement

	for i := offset; i < offset+limit && i < 100; i++ {
		users = append(users, UserManagement{
			UserID:    fmt.Sprintf("user_%d", i),
			Username:  fmt.Sprintf("user%d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			Status:    "active",
			Role:      "user",
			Balance:   decimal.NewFromFloat(float64(100 + i*10)),
			CreatedAt: time.Now().AddDate(0, 0, -i*30),
			LastLogin: time.Now().Add(time.Duration(-i) * time.Hour),
			KYCStatus: "verified",
			Limits: UserLimits{
				DailyLimit:   decimal.NewFromFloat(1000),
				WeeklyLimit:  decimal.NewFromFloat(5000),
				MonthlyLimit: decimal.NewFromFloat(20000),
				MaxBetSize:   decimal.NewFromFloat(100),
			},
		})
	}

	return users, nil
}

// SuspendUser suspends a user account
func (s *AdminService) SuspendUser(ctx context.Context, userID string, reason string) error {
	// TODO: Implement user suspension logic
	s.publishEvent("admin.user.suspended", map[string]interface{}{
		"user_id": userID,
		"reason":  reason,
		"time":    time.Now(),
	})

	return nil
}

// UnsuspendUser unsuspends a user account
func (s *AdminService) UnsuspendUser(ctx context.Context, userID string) error {
	// TODO: Implement user unsuspension logic
	s.publishEvent("admin.user.unsuspended", map[string]interface{}{
		"user_id": userID,
		"time":    time.Now(),
	})

	return nil
}

// publishEvent publishes an event to the event bus
func (s *AdminService) publishEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing admin event %s: %v", topic, err)
		}
	}
}
