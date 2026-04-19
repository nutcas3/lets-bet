package admin

import (
	"time"

	"github.com/shopspring/decimal"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// AdminService handles administrative operations
type AdminService struct {
	eventBus EventBus
}

// NewAdminService creates a new admin service
func NewAdminService(eventBus EventBus) *AdminService {
	return &AdminService{
		eventBus: eventBus,
	}
}

// DashboardData represents the main dashboard overview
type DashboardData struct {
	Overview       OverviewStats  `json:"overview"`
	RecentActivity []ActivityItem `json:"recent_activity"`
	TopUsers       []UserStats    `json:"top_users"`
	RevenueMetrics RevenueMetrics `json:"revenue_metrics"`
	BettingMetrics BettingMetrics `json:"betting_metrics"`
	FinancialStats FinancialStats `json:"financial_stats"`
	SystemHealth   SystemHealth   `json:"system_health"`
}

// OverviewStats represents high-level platform statistics
type OverviewStats struct {
	TotalUsers          int64           `json:"total_users"`
	ActiveUsers         int64           `json:"active_users"`
	TotalBets           int64           `json:"total_bets"`
	TotalVolume         decimal.Decimal `json:"total_volume"`
	TotalRevenue        decimal.Decimal `json:"total_revenue"`
	TotalPayouts        decimal.Decimal `json:"total_payouts"`
	ActiveMatches       int64           `json:"active_matches"`
	PendingTransactions int64           `json:"pending_transactions"`
	SystemUptime        string          `json:"system_uptime"`
}

// ActivityItem represents a recent activity item
type ActivityItem struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Description string          `json:"description"`
	UserID      string          `json:"user_id"`
	Timestamp   time.Time       `json:"timestamp"`
	Amount      decimal.Decimal `json:"amount,omitempty"`
	Status      string          `json:"status"`
}

// UserStats represents user statistics
type UserStats struct {
	UserID       string          `json:"user_id"`
	Username     string          `json:"username"`
	TotalBets    int64           `json:"total_bets"`
	TotalVolume  decimal.Decimal `json:"total_volume"`
	TotalRevenue decimal.Decimal `json:"total_revenue"`
	WinRate      decimal.Decimal `json:"win_rate"`
	LastActive   time.Time       `json:"last_active"`
	Status       string          `json:"status"`
}

// RevenueMetrics represents revenue metrics
type RevenueMetrics struct {
	DailyRevenue   decimal.Decimal `json:"daily_revenue"`
	WeeklyRevenue  decimal.Decimal `json:"weekly_revenue"`
	MonthlyRevenue decimal.Decimal `json:"monthly_revenue"`
	YearlyRevenue  decimal.Decimal `json:"yearly_revenue"`
	GrowthRate     decimal.Decimal `json:"growth_rate"`
}

// BettingMetrics represents betting metrics
type BettingMetrics struct {
	DailyBets   int64           `json:"daily_bets"`
	WeeklyBets  int64           `json:"weekly_bets"`
	MonthlyBets int64           `json:"monthly_bets"`
	YearlyBets  int64           `json:"yearly_bets"`
	AverageBet  decimal.Decimal `json:"average_bet"`
}

// FinancialStats represents financial statistics
type FinancialStats struct {
	TotalDeposits      decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals   decimal.Decimal `json:"total_withdrawals"`
	NetCashFlow        decimal.Decimal `json:"net_cash_flow"`
	PendingDeposits    int64           `json:"pending_deposits"`
	PendingWithdrawals int64           `json:"pending_withdrawals"`
}

// SystemHealth represents system health status
type SystemHealth struct {
	OverallStatus string          `json:"overall_status"`
	Services      []ServiceStatus `json:"services"`
	Database      DatabaseStatus  `json:"database"`
	Cache         CacheStatus     `json:"cache"`
	API           APIStatus       `json:"api"`
	LoadAverage   []float64       `json:"load_average"`
	MemoryUsage   MemoryUsage     `json:"memory_usage"`
	DiskUsage     DiskUsage       `json:"disk_usage"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Uptime    time.Duration `json:"uptime"`
	LastCheck time.Time     `json:"last_check"`
	Message   string        `json:"message"`
}

// DatabaseStatus represents database status
type DatabaseStatus struct {
	Status      string        `json:"status"`
	Connections int           `json:"connections"`
	QueryTime   time.Duration `json:"query_time"`
	LastCheck   time.Time     `json:"last_check"`
}

// CacheStatus represents cache status
type CacheStatus struct {
	Status      string          `json:"status"`
	HitRate     decimal.Decimal `json:"hit_rate"`
	MemoryUsage decimal.Decimal `json:"memory_usage"`
	LastCheck   time.Time       `json:"last_check"`
}

// APIStatus represents API status
type APIStatus struct {
	Status            string          `json:"status"`
	ResponseTime      time.Duration   `json:"response_time"`
	RequestsPerSecond decimal.Decimal `json:"requests_per_second"`
	LastCheck         time.Time       `json:"last_check"`
}

// MemoryUsage represents memory usage statistics
type MemoryUsage struct {
	Total     decimal.Decimal `json:"total"`
	Used      decimal.Decimal `json:"used"`
	Available decimal.Decimal `json:"available"`
	Percent   decimal.Decimal `json:"percent"`
}

// DiskUsage represents disk usage statistics
type DiskUsage struct {
	Total     decimal.Decimal `json:"total"`
	Used      decimal.Decimal `json:"used"`
	Available decimal.Decimal `json:"available"`
	Percent   decimal.Decimal `json:"percent"`
}

// AdminConfig represents admin service configuration
type AdminConfig struct {
	DashboardRefreshInterval time.Duration `json:"dashboard_refresh_interval"`
	MaxActivityItems         int           `json:"max_activity_items"`
	MaxTopUsers              int           `json:"max_top_users"`
	EnableRealTimeUpdates    bool          `json:"enable_real_time_updates"`
}

// AdminMetrics represents admin service metrics
type AdminMetrics struct {
	TotalAdmins    int64           `json:"total_admins"`
	ActiveAdmins   int64           `json:"active_admins"`
	TotalActions   int64           `json:"total_actions"`
	ActionsPerHour decimal.Decimal `json:"actions_per_hour"`
	LastActivity   time.Time       `json:"last_activity"`
	SystemLoad     decimal.Decimal `json:"system_load"`
	ResponseTime   time.Duration   `json:"response_time"`
}

// UserManagementRequest represents a user management request
type UserManagementRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Search   string `json:"search"`
	Status   string `json:"status"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir"`
}

// UserActionRequest represents a user action request
type UserActionRequest struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
	Reason string `json:"reason,omitempty"`
}

// BettingMetricsRequest represents a betting metrics request
type BettingMetricsRequest struct {
	Period    string     `json:"period"`
	TimeRange string     `json:"time_range"`
	From      *time.Time `json:"from,omitempty"`
	FromDate  *time.Time `json:"from_date,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// FinancialReportRequest represents a financial report request
type FinancialReportRequest struct {
	Period     string     `json:"period"`
	Format     string     `json:"format"`
	ReportType string     `json:"report_type"`
	TimeRange  string     `json:"time_range"`
	From       *time.Time `json:"from,omitempty"`
	FromDate   *time.Time `json:"from_date,omitempty"`
	To         *time.Time `json:"to,omitempty"`
	ToDate     *time.Time `json:"to_date,omitempty"`
}

// AuditLogRequest represents an audit log request
type AuditLogRequest struct {
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
	UserID   string     `json:"user_id"`
	From     *time.Time `json:"from,omitempty"`
	FromDate *time.Time `json:"from_date,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	ToDate   *time.Time `json:"to_date,omitempty"`
	Action   string     `json:"action,omitempty"`
	User     string     `json:"user,omitempty"`
}
