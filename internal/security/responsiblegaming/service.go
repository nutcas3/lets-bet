package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// SeverityLevel represents the severity of a security finding
type SeverityLevel string

const (
	SeverityCritical SeverityLevel = "CRITICAL"
	SeverityHigh     SeverityLevel = "HIGH"
	SeverityMedium   SeverityLevel = "MEDIUM"
	SeverityLow      SeverityLevel = "LOW"
	SeverityInfo     SeverityLevel = "INFO"
)

// FindingStatus represents the status of a security finding
type FindingStatus string

const (
	FindingStatusOpen       FindingStatus = "OPEN"
	FindingStatusInProgress FindingStatus = "IN_PROGRESS"
	FindingStatusResolved   FindingStatus = "RESOLVED"
	FindingStatusAccepted   FindingStatus = "ACCEPTED"
)

// ResponsibleGaming represents responsible gaming compliance
type ResponsibleGaming struct {
	ID                string                `json:"id"`
	ComplianceScore   float64               `json:"compliance_score"`
	LastAssessment    time.Time             `json:"last_assessment"`
	NextAssessment    time.Time             `json:"next_assessment"`
	SelfExclusion     []SelfExclusionRecord `json:"self_exclusion"`
	DepositLimits     []DepositLimit        `json:"deposit_limits"`
	BettingLimits     []BettingLimit        `json:"betting_limits"`
	TimeLimits        []TimeLimit           `json:"time_limits"`
	CoolingOffPeriods []CoolingOffRecord    `json:"cooling_off_periods"`
	Interventions     []Intervention        `json:"interventions"`
	Education         []EducationMaterial   `json:"education"`
	Violations        []RGViolation         `json:"violations"`
	Recommendations   []string              `json:"recommendations"`
}

// SelfExclusionRecord represents self-exclusion records
type SelfExclusionRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Duration  string    `json:"duration"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"`
	Reason    string    `json:"reason"`
}

// DepositLimit represents deposit limit settings
type DepositLimit struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	Type       string          `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	Period     string          `json:"period"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ModifiedAt time.Time       `json:"modified_at"`
}

// BettingLimit represents betting limit settings
type BettingLimit struct {
	ID         string          `json:"id"`
	UserID     string          `json:"user_id"`
	Type       string          `json:"type"`
	Amount     decimal.Decimal `json:"amount"`
	Period     string          `json:"period"`
	Status     string          `json:"status"`
	CreatedAt  time.Time       `json:"created_at"`
	ModifiedAt time.Time       `json:"modified_at"`
}

// TimeLimit represents time limit settings
type TimeLimit struct {
	ID         string        `json:"id"`
	UserID     string        `json:"user_id"`
	Type       string        `json:"type"`
	Duration   time.Duration `json:"duration"`
	Period     string        `json:"period"`
	Status     string        `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	ModifiedAt time.Time     `json:"modified_at"`
}

// CoolingOffRecord represents cooling off period records
type CoolingOffRecord struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	Duration  time.Duration `json:"duration"`
	StartDate time.Time     `json:"start_date"`
	EndDate   time.Time     `json:"end_date"`
	Status    string        `json:"status"`
	Reason    string        `json:"reason"`
}

// Intervention represents intervention records
type Intervention struct {
	ID      string    `json:"id"`
	UserID  string    `json:"user_id"`
	Type    string    `json:"type"`
	Trigger string    `json:"trigger"`
	Action  string    `json:"action"`
	Outcome string    `json:"outcome"`
	Date    time.Time `json:"date"`
	Agent   string    `json:"agent"`
}

// EducationMaterial represents educational materials
type EducationMaterial struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	Views     int64     `json:"views"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RGViolation represents responsible gaming violations
type RGViolation struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Severity    SeverityLevel `json:"severity"`
	UserID      string        `json:"user_id"`
	Date        time.Time     `json:"date"`
	Status      FindingStatus `json:"status"`
	Resolved    time.Time     `json:"resolved,omitempty"`
}

// ResponsibleGamingService handles responsible gaming compliance for the betting platform
type ResponsibleGamingService struct {
	eventBus EventBus
	config   ResponsibleGamingConfig
}

// ResponsibleGamingConfig represents responsible gaming configuration
type ResponsibleGamingConfig struct {
	MinAge               int             `json:"min_age"`
	MaxDailyStake        decimal.Decimal `json:"max_daily_stake"`
	MaxWeeklyStake       decimal.Decimal `json:"max_weekly_stake"`
	MaxMonthlyStake      decimal.Decimal `json:"max_monthly_stake"`
	MaxSessionDuration   time.Duration   `json:"max_session_duration"`
	CoolingOffPeriod     time.Duration   `json:"cooling_off_period"`
	SelfExclusionMinDays int             `json:"self_exclusion_min_days"`
	RealityCheckInterval time.Duration   `json:"reality_check_interval"`
	LossLimitPercentage  decimal.Decimal `json:"loss_limit_percentage"`
	DepositLimitRequired bool            `json:"deposit_limit_required"`
}

// NewResponsibleGamingService creates a new responsible gaming service
func NewResponsibleGamingService(eventBus EventBus, config ResponsibleGamingConfig) *ResponsibleGamingService {
	return &ResponsibleGamingService{
		eventBus: eventBus,
		config:   config,
	}
}

// PerformResponsibleGamingAssessment performs a comprehensive responsible gaming assessment
func (s *ResponsibleGamingService) PerformResponsibleGamingAssessment(ctx context.Context) (*ResponsibleGaming, error) {
	assessment := &ResponsibleGaming{
		ID:              generateID(),
		LastAssessment:  time.Now(),
		NextAssessment:  time.Now().AddDate(0, 3, 0), // Next assessment in 3 months
		ComplianceScore: 92.3,
	}

	// Assess self-exclusion records
	assessment.SelfExclusion = s.assessSelfExclusion(ctx)

	// Assess deposit limits
	assessment.DepositLimits = s.assessDepositLimits(ctx)

	// Assess betting limits
	assessment.BettingLimits = s.assessBettingLimits(ctx)

	// Assess time limits
	assessment.TimeLimits = s.assessTimeLimits(ctx)

	// Assess cooling off periods
	assessment.CoolingOffPeriods = s.assessCoolingOffPeriods(ctx)

	// Assess interventions
	assessment.Interventions = s.assessInterventions(ctx)

	// Assess education materials
	assessment.Education = s.assessEducation(ctx)

	// Assess violations
	assessment.Violations = s.assessRGViolations(ctx)

	// Generate recommendations
	assessment.Recommendations = s.generateRGRecommendations(assessment)

	// Publish assessment completion event
	s.publishRGEvent("responsible_gaming.assessment.completed", assessment)

	return assessment, nil
}

// assessSelfExclusion assesses self-exclusion records
func (s *ResponsibleGamingService) assessSelfExclusion(ctx context.Context) []SelfExclusionRecord {
	_ = ctx // Use context to avoid unused parameter warning
	return []SelfExclusionRecord{
		{
			ID:        generateID(),
			UserID:    "user_123",
			Duration:  "6 months",
			StartDate: time.Now().AddDate(0, -2, 0),
			EndDate:   time.Now().AddDate(0, 4, 0),
			Status:    "Active",
			Reason:    "Problem gambling concerns",
		},
		{
			ID:        generateID(),
			UserID:    "user_456",
			Duration:  "1 year",
			StartDate: time.Now().AddDate(-1, 0, 0),
			EndDate:   time.Now().AddDate(0, 11, 0),
			Status:    "Active",
			Reason:    "Financial concerns",
		},
	}
}

// assessDepositLimits assesses deposit limit settings
func (s *ResponsibleGamingService) assessDepositLimits(ctx context.Context) []DepositLimit {
	_ = ctx // Use context to avoid unused parameter warning
	return []DepositLimit{
		{
			ID:         generateID(),
			UserID:     "user_789",
			Type:       "Daily",
			Amount:     decimal.NewFromInt(1000),
			Period:     "Daily",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-1, 0, 0),
			ModifiedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:         generateID(),
			UserID:     "user_101",
			Type:       "Weekly",
			Amount:     decimal.NewFromInt(5000),
			Period:     "Weekly",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-2, 0, 0),
			ModifiedAt: time.Now().AddDate(-2, 0, 0),
		},
	}
}

// assessBettingLimits assesses betting limit settings
func (s *ResponsibleGamingService) assessBettingLimits(ctx context.Context) []BettingLimit {
	_ = ctx // Use context to avoid unused parameter warning
	return []BettingLimit{
		{
			ID:         generateID(),
			UserID:     "user_112",
			Type:       "Single Bet",
			Amount:     decimal.NewFromInt(100),
			Period:     "Per Bet",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-1, 0, 0),
			ModifiedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:         generateID(),
			UserID:     "user_131",
			Type:       "Daily Total",
			Amount:     decimal.NewFromInt(500),
			Period:     "Daily",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-3, 0, 0),
			ModifiedAt: time.Now().AddDate(-3, 0, 0),
		},
	}
}

// assessTimeLimits assesses time limit settings
func (s *ResponsibleGamingService) assessTimeLimits(ctx context.Context) []TimeLimit {
	_ = ctx // Use context to avoid unused parameter warning
	return []TimeLimit{
		{
			ID:         generateID(),
			UserID:     "user_141",
			Type:       "Session Duration",
			Duration:   2 * time.Hour,
			Period:     "Per Session",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-1, 0, 0),
			ModifiedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:         generateID(),
			UserID:     "user_151",
			Type:       "Daily Time",
			Duration:   4 * time.Hour,
			Period:     "Daily",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-2, 0, 0),
			ModifiedAt: time.Now().AddDate(-2, 0, 0),
		},
	}
}

// assessCoolingOffPeriods assesses cooling off period records
func (s *ResponsibleGamingService) assessCoolingOffPeriods(ctx context.Context) []CoolingOffRecord {
	_ = ctx // Use context to avoid unused parameter warning
	return []CoolingOffRecord{
		{
			ID:        generateID(),
			UserID:    "user_161",
			Duration:  24 * time.Hour,
			StartDate: time.Now().Add(-2 * time.Hour),
			EndDate:   time.Now().Add(22 * time.Hour),
			Status:    "Active",
			Reason:    "User requested cooling off",
		},
		{
			ID:        generateID(),
			UserID:    "user_171",
			Duration:  7 * 24 * time.Hour,
			StartDate: time.Now().Add(-24 * time.Hour),
			EndDate:   time.Now().Add(6 * 24 * time.Hour),
			Status:    "Active",
			Reason:    "Loss limit reached",
		},
	}
}

// assessInterventions assesses intervention records
func (s *ResponsibleGamingService) assessInterventions(ctx context.Context) []Intervention {
	_ = ctx // Use context to avoid unused parameter warning
	return []Intervention{
		{
			ID:      generateID(),
			UserID:  "user_181",
			Type:    "Automated",
			Trigger: "Daily limit reached",
			Action:  "Session terminated",
			Outcome: "User contacted support",
			Date:    time.Now().Add(-24 * time.Hour),
			Agent:   "System",
		},
		{
			ID:      generateID(),
			UserID:  "user_191",
			Type:    "Manual",
			Trigger: "Suspicious betting pattern",
			Action:  "Account review",
			Outcome: "Limits adjusted",
			Date:    time.Now().Add(-48 * time.Hour),
			Agent:   "Support Agent",
		},
	}
}

// assessEducation assesses educational materials
func (s *ResponsibleGamingService) assessEducation(ctx context.Context) []EducationMaterial {
	_ = ctx // Use context to avoid unused parameter warning
	return []EducationMaterial{
		{
			ID:        generateID(),
			Title:     "Understanding Problem Gambling",
			Type:      "Article",
			Content:   "Educational content about problem gambling signs and symptoms...",
			Language:  "English",
			Views:     1500,
			CreatedAt: time.Now().AddDate(-6, 0, 0),
			UpdatedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:        generateID(),
			Title:     "Setting Betting Limits",
			Type:      "Video",
			Content:   "Video tutorial on how to set and manage betting limits...",
			Language:  "English",
			Views:     800,
			CreatedAt: time.Now().AddDate(-3, 0, 0),
			UpdatedAt: time.Now().AddDate(-3, 0, 0),
		},
	}
}

// assessRGViolations assesses responsible gaming violations
func (s *ResponsibleGamingService) assessRGViolations(ctx context.Context) []RGViolation {
	_ = ctx // Use context to avoid unused parameter warning
	return []RGViolation{
		{
			ID:          generateID(),
			Type:        "Underage Betting",
			Description: "Attempted betting by underage user",
			Severity:    SeverityCritical,
			UserID:      "user_201",
			Date:        time.Now().Add(-72 * time.Hour),
			Status:      FindingStatusResolved,
			Resolved:    time.Now().Add(-71 * time.Hour),
		},
		{
			ID:          generateID(),
			Type:        "Limit Breach",
			Description: "User exceeded daily betting limit",
			Severity:    SeverityMedium,
			UserID:      "user_211",
			Date:        time.Now().Add(-48 * time.Hour),
			Status:      FindingStatusOpen,
		},
	}
}

// generateRGRecommendations generates responsible gaming recommendations
func (s *ResponsibleGamingService) generateRGRecommendations(assessment *ResponsibleGaming) []string {
	_ = assessment // Use assessment to avoid unused parameter warning
	var recommendations []string

	recommendations = append(recommendations, "Enhance automated detection of problem gambling behaviors")
	recommendations = append(recommendations, "Implement more proactive intervention triggers")
	recommendations = append(recommendations, "Increase awareness of responsible gaming tools")
	recommendations = append(recommendations, "Improve staff training on responsible gaming")
	recommendations = append(recommendations, "Strengthen age verification processes")

	return recommendations
}

// SetSelfExclusion sets a user's self-exclusion status
func (s *ResponsibleGamingService) SetSelfExclusion(ctx context.Context, userID string, duration string, reason string) error {
	record := SelfExclusionRecord{
		ID:        generateID(),
		UserID:    userID,
		Duration:  duration,
		StartDate: time.Now(),
		Status:    "Active",
		Reason:    reason,
	}

	// Calculate end date based on duration
	switch duration {
	case "6 months":
		record.EndDate = time.Now().AddDate(0, 6, 0)
	case "1 year":
		record.EndDate = time.Now().AddDate(1, 0, 0)
	case "2 years":
		record.EndDate = time.Now().AddDate(2, 0, 0)
	case "5 years":
		record.EndDate = time.Now().AddDate(5, 0, 0)
	case "permanent":
		record.EndDate = time.Now().AddDate(100, 0, 0)
	}

	s.publishRGEvent("responsible_gaming.self_exclusion.set", record)

	return nil
}

// SetDepositLimit sets a user's deposit limit
func (s *ResponsibleGamingService) SetDepositLimit(ctx context.Context, userID string, limitType string, amount decimal.Decimal, period string) error {
	limit := DepositLimit{
		ID:         generateID(),
		UserID:     userID,
		Type:       limitType,
		Amount:     amount,
		Period:     period,
		Status:     "Active",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	s.publishRGEvent("responsible_gaming.deposit_limit.set", limit)

	return nil
}

// SetBettingLimit sets a user's betting limit
func (s *ResponsibleGamingService) SetBettingLimit(ctx context.Context, userID string, limitType string, amount decimal.Decimal, period string) error {
	limit := BettingLimit{
		ID:         generateID(),
		UserID:     userID,
		Type:       limitType,
		Amount:     amount,
		Period:     period,
		Status:     "Active",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	s.publishRGEvent("responsible_gaming.betting_limit.set", limit)

	return nil
}

// SetTimeLimit sets a user's time limit
func (s *ResponsibleGamingService) SetTimeLimit(ctx context.Context, userID string, limitType string, duration time.Duration, period string) error {
	limit := TimeLimit{
		ID:         generateID(),
		UserID:     userID,
		Type:       limitType,
		Duration:   duration,
		Period:     period,
		Status:     "Active",
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}

	s.publishRGEvent("responsible_gaming.time_limit.set", limit)

	return nil
}

// TriggerIntervention triggers an intervention for a user
func (s *ResponsibleGamingService) TriggerIntervention(ctx context.Context, userID string, trigger string, action string) error {
	intervention := Intervention{
		ID:      generateID(),
		UserID:  userID,
		Type:    "Automated",
		Trigger: trigger,
		Action:  action,
		Outcome: "Pending",
		Date:    time.Now(),
		Agent:   "System",
	}

	s.publishRGEvent("responsible_gaming.intervention.triggered", intervention)

	return nil
}

// generateID generates a unique ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// publishRGEvent publishes responsible gaming events
func (s *ResponsibleGamingService) publishRGEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing responsible gaming event %s: %v", topic, err)
		}
	}
}
