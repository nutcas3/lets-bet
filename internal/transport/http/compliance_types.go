package http

import (
	"time"

	bclb "github.com/betting-platform/internal/compliance/bclb"
	"github.com/shopspring/decimal"
)

// ComplianceHandler handles compliance-related HTTP requests
type ComplianceHandler struct {
	bclbService *bclb.BCLBService
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(bclbService *bclb.BCLBService) *ComplianceHandler {
	return &ComplianceHandler{
		bclbService: bclbService,
	}
}

// BetValidationRequest represents a bet validation request
type BetValidationRequest struct {
	UserID     string          `json:"user_id"`
	BetAmount  decimal.Decimal `json:"bet_amount"`
	BetType    string          `json:"bet_type"`
	Selections int             `json:"selections"`
	MarketID   string          `json:"market_id,omitempty"`
	EventID    string          `json:"event_id,omitempty"`
	Odds       decimal.Decimal `json:"odds"`
}

// BetValidationResponse represents a bet validation response
type BetValidationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// BetValidation represents bet validation results
type BetValidation struct {
	Allowed      bool       `json:"allowed"`
	Reason       string     `json:"reason,omitempty"`
	Limits       *BetLimits `json:"limits,omitempty"`
	RiskScore    int        `json:"risk_score"`
	Warnings     []string   `json:"warnings,omitempty"`
	Requirements []string   `json:"requirements,omitempty"`
	ValidUntil   time.Time  `json:"valid_until"`
}

// BetLimits represents betting limits
type BetLimits struct {
	MinBetAmount     decimal.Decimal `json:"min_bet_amount"`
	MaxBetAmount     decimal.Decimal `json:"max_bet_amount"`
	MaxDailyAmount   decimal.Decimal `json:"max_daily_amount"`
	MaxWeeklyAmount  decimal.Decimal `json:"max_weekly_amount"`
	MaxMonthlyAmount decimal.Decimal `json:"max_monthly_amount"`
	CurrentDaily     decimal.Decimal `json:"current_daily"`
	CurrentWeekly    decimal.Decimal `json:"current_weekly"`
	CurrentMonthly   decimal.Decimal `json:"current_monthly"`
}

// UserComplianceStatus represents user compliance status
type UserComplianceStatus struct {
	UserID            string             `json:"user_id"`
	Status            string             `json:"status"`
	VerificationLevel string             `json:"verification_level"`
	RiskProfile       string             `json:"risk_profile"`
	Limits            *UserLimits        `json:"limits"`
	Restrictions      []*UserRestriction `json:"restrictions"`
	LastUpdated       time.Time          `json:"last_updated"`
	NextReview        time.Time          `json:"next_review"`
}

// UserLimits represents user-specific limits
type UserLimits struct {
	DailyLimit   decimal.Decimal `json:"daily_limit"`
	WeeklyLimit  decimal.Decimal `json:"weekly_limit"`
	MonthlyLimit decimal.Decimal `json:"monthly_limit"`
	BetLimit     decimal.Decimal `json:"bet_limit"`
	LossLimit    decimal.Decimal `json:"loss_limit"`
	TimeLimit    int             `json:"time_limit"` // minutes
	CurrentUsage *CurrentUsage   `json:"current_usage"`
}

// CurrentUsage represents current usage against limits
type CurrentUsage struct {
	DailyAmount   decimal.Decimal `json:"daily_amount"`
	WeeklyAmount  decimal.Decimal `json:"weekly_amount"`
	MonthlyAmount decimal.Decimal `json:"monthly_amount"`
	BetCount      int64           `json:"bet_count"`
	SessionTime   int             `json:"session_time"` // minutes
	LastBet       time.Time       `json:"last_bet"`
}

// UserRestriction represents user restrictions
type UserRestriction struct {
	Type        string    `json:"type"`
	Reason      string    `json:"reason"`
	StartedAt   time.Time `json:"started_at"`
	EndsAt      time.Time `json:"ends_at"`
	IsActive    bool      `json:"is_active"`
	AppliedBy   string    `json:"applied_by"`
	Description string    `json:"description"`
}

// ComplianceReport represents a compliance report
type ComplianceReport struct {
	ReportID         string            `json:"report_id"`
	Type             string            `json:"type"`
	Period           *TimePeriod       `json:"period"`
	GeneratedAt      time.Time         `json:"generated_at"`
	GeneratedBy      string            `json:"generated_by"`
	Summary          *ReportSummary    `json:"summary"`
	UserStats        *UserStatistics   `json:"user_stats"`
	TransactionStats *TransactionStats `json:"transaction_stats"`
	Violations       []*Violation      `json:"violations"`
	Recommendations  []*Recommendation `json:"recommendations"`
}

// TimePeriod represents a time period
type TimePeriod struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// ReportSummary represents report summary
type ReportSummary struct {
	TotalUsers        int64           `json:"total_users"`
	ActiveUsers       int64           `json:"active_users"`
	HighRiskUsers     int64           `json:"high_risk_users"`
	TotalTransactions int64           `json:"total_transactions"`
	TotalVolume       decimal.Decimal `json:"total_volume"`
	AlertCount        int64           `json:"alert_count"`
	ViolationCount    int64           `json:"violation_count"`
	ComplianceScore   decimal.Decimal `json:"compliance_score"`
}

// UserStatistics represents user statistics
type UserStatistics struct {
	NewUsers              int64           `json:"new_users"`
	VerifiedUsers         int64           `json:"verified_users"`
	HighRiskUsers         int64           `json:"high_risk_users"`
	SelfExcludedUsers     int64           `json:"self_excluded_users"`
	UsersWithLimits       int64           `json:"users_with_limits"`
	UsersWithRestrictions int64           `json:"users_with_restrictions"`
	AverageRiskScore      decimal.Decimal `json:"average_risk_score"`
}

// TransactionStats represents transaction statistics
type TransactionStats struct {
	TotalTransactions int64            `json:"total_transactions"`
	TotalAmount       decimal.Decimal  `json:"total_amount"`
	TotalPayout       decimal.Decimal  `json:"total_payout"`
	TotalProfit       decimal.Decimal  `json:"total_profit"`
	AverageBetSize    decimal.Decimal  `json:"average_bet_size"`
	LargestBet        decimal.Decimal  `json:"largest_bet"`
	SmallestBet       decimal.Decimal  `json:"smallest_bet"`
	ByType            map[string]int64 `json:"by_type"`
	ByStatus          map[string]int64 `json:"by_status"`
}

// Violation represents a compliance violation
type Violation struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	UserID      string     `json:"user_id"`
	Description string     `json:"description"`
	DetectedAt  time.Time  `json:"detected_at"`
	Status      string     `json:"status"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy  string     `json:"resolved_by,omitempty"`
	Actions     []string   `json:"actions"`
}

// Recommendation represents a compliance recommendation
type Recommendation struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Priority    string     `json:"priority"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Impact      string     `json:"impact"`
	Effort      string     `json:"effort"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	AssignedTo  string     `json:"assigned_to,omitempty"`
}

// ComplianceAlert represents a compliance alert
type ComplianceAlert struct {
	ID             string         `json:"id"`
	Type           string         `json:"type"`
	Severity       string         `json:"severity"`
	UserID         string         `json:"user_id,omitempty"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TriggeredAt    time.Time      `json:"triggered_at"`
	Status         string         `json:"status"`
	AcknowledgedBy string         `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
	ResolvedBy     string         `json:"resolved_by,omitempty"`
	ResolvedAt     *time.Time     `json:"resolved_at,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// ComplianceRule represents a compliance rule
type ComplianceRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Category    string         `json:"category"`
	Enabled     bool           `json:"enabled"`
	Severity    string         `json:"severity"`
	Conditions  map[string]any `json:"conditions"`
	Actions     []string       `json:"actions"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   string         `json:"created_by"`
	UpdatedBy   string         `json:"updated_by"`
}

// ComplianceCheck represents a compliance check
type ComplianceCheck struct {
	CheckID   string         `json:"check_id"`
	UserID    string         `json:"user_id"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Result    *CheckResult   `json:"result"`
	CheckedAt time.Time      `json:"checked_at"`
	CheckedBy string         `json:"checked_by"`
	Notes     string         `json:"notes,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// CheckResult represents check result
type CheckResult struct {
	Passed          bool     `json:"passed"`
	Score           int      `json:"score"`
	MaxScore        int      `json:"max_score"`
	Reason          string   `json:"reason,omitempty"`
	Warnings        []string `json:"warnings,omitempty"`
	Errors          []string `json:"errors,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// ComplianceMetrics represents compliance metrics
type ComplianceMetrics struct {
	Period           *TimePeriod      `json:"period"`
	TotalChecks      int64            `json:"total_checks"`
	PassedChecks     int64            `json:"passed_checks"`
	FailedChecks     int64            `json:"failed_checks"`
	AlertCount       int64            `json:"alert_count"`
	ViolationCount   int64            `json:"violation_count"`
	ResolutionTime   time.Duration    `json:"resolution_time"`
	ComplianceRate   decimal.Decimal  `json:"compliance_rate"`
	RiskDistribution map[string]int64 `json:"risk_distribution"`
	TopViolations    []*Violation     `json:"top_violations"`
	Trends           []*MetricTrend   `json:"trends"`
}

// MetricTrend represents metric trends
type MetricTrend struct {
	Date       time.Time       `json:"date"`
	Metric     string          `json:"metric"`
	Value      int64           `json:"value"`
	Change     decimal.Decimal `json:"change"`
	ChangeType string          `json:"change_type"`
}

// ComplianceSettings represents compliance settings
type ComplianceSettings struct {
	Enabled               bool                     `json:"enabled"`
	AutoVerification      bool                     `json:"auto_verification"`
	RiskAssessmentEnabled bool                     `json:"risk_assessment_enabled"`
	LimitEnforcement      bool                     `json:"limit_enforcement"`
	AlertThresholds       map[string]int           `json:"alert_thresholds"`
	VerificationLevels    []*VerificationLevel     `json:"verification_levels"`
	DefaultLimits         *DefaultLimits           `json:"default_limits"`
	RiskProfiles          []*ComplianceRiskProfile `json:"risk_profiles"`
	ComplianceRules       []*ComplianceRule        `json:"compliance_rules"`
	ReportingSchedule     string                   `json:"reporting_schedule"`
	NotificationSettings  *NotificationSettings    `json:"notification_settings"`
}

// VerificationLevel represents verification levels
type VerificationLevel struct {
	Level           int             `json:"level"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Requirements    []string        `json:"requirements"`
	Benefits        []string        `json:"benefits"`
	LimitMultiplier decimal.Decimal `json:"limit_multiplier"`
}

// DefaultLimits represents default limits
type DefaultLimits struct {
	MinBetAmount decimal.Decimal `json:"min_bet_amount"`
	MaxBetAmount decimal.Decimal `json:"max_bet_amount"`
	DailyLimit   decimal.Decimal `json:"daily_limit"`
	WeeklyLimit  decimal.Decimal `json:"weekly_limit"`
	MonthlyLimit decimal.Decimal `json:"monthly_limit"`
	LossLimit    decimal.Decimal `json:"loss_limit"`
	SessionLimit int             `json:"session_limit"`
}

// ComplianceRiskProfile represents risk profiles
type ComplianceRiskProfile struct {
	Profile         string          `json:"profile"`
	Description     string          `json:"description"`
	ScoreRange      [2]int          `json:"score_range"`
	LimitMultiplier decimal.Decimal `json:"limit_multiplier"`
	Restrictions    []string        `json:"restrictions"`
	MonitoringLevel string          `json:"monitoring_level"`
}

// NotificationSettings represents notification settings
type NotificationSettings struct {
	EmailEnabled   bool     `json:"email_enabled"`
	SMSEnabled     bool     `json:"sms_enabled"`
	PushEnabled    bool     `json:"push_enabled"`
	Recipients     []string `json:"recipients"`
	AlertTypes     []string `json:"alert_types"`
	DigestSchedule string   `json:"digest_schedule"`
}
