package http

import (
	"time"

	"github.com/shopspring/decimal"
)

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
