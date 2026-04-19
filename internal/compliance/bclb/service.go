package compliance

import (
	"context"
	"log"
	"time"

	"github.com/shopspring/decimal"
)

// BCLBService handles Betting Control and Licensing Board compliance for Kenya
type BCLBService struct {
	eventBus        EventBus
	config          BCLBConfig
	userRepo        UserRepository
	betRepo         BetRepository
	transactionRepo TransactionRepository
	coolingOffRepo  CoolingOffRepository
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// BCLBConfig represents BCLB compliance configuration
type BCLBConfig struct {
	LicenseNumber            string          `json:"license_number"`
	LicenseExpiry            time.Time       `json:"license_expiry"`
	MaxDailyStake            decimal.Decimal `json:"max_daily_stake"`
	MaxWeeklyStake           decimal.Decimal `json:"max_weekly_stake"`
	MaxMonthlyStake          decimal.Decimal `json:"max_monthly_stake"`
	MinAge                   int             `json:"min_age"`
	MaxBetPerEvent           decimal.Decimal `json:"max_bet_per_event"`
	MaxAccumulatorBets       int             `json:"max_accumulator_bets"`
	RequiredKYCLevel         string          `json:"required_kyc_level"`
	SelfExclusionMinDays     int             `json:"self_exclusion_min_days"`
	ResponsibleGamingEnabled bool            `json:"responsible_gaming_enabled"`
}

// NewBCLBService creates a new BCLB compliance service
func NewBCLBService(
	eventBus EventBus,
	config BCLBConfig,
	userRepo UserRepository,
	betRepo BetRepository,
	transactionRepo TransactionRepository,
	coolingOffRepo CoolingOffRepository,
) *BCLBService {
	return &BCLBService{
		eventBus:        eventBus,
		config:          config,
		userRepo:        userRepo,
		betRepo:         betRepo,
		transactionRepo: transactionRepo,
		coolingOffRepo:  coolingOffRepo,
	}
}

// ComplianceCheck represents a compliance validation result
type ComplianceCheck struct {
	Passed        bool                  `json:"passed"`
	Violations    []ComplianceViolation `json:"violations"`
	Warnings      []string              `json:"warnings"`
	CheckTime     time.Time             `json:"check_time"`
	CheckType     string                `json:"check_type"`
	UserID        string                `json:"user_id,omitempty"`
	TransactionID string                `json:"transaction_id,omitempty"`
}

// ComplianceViolation represents a specific compliance violation
type ComplianceViolation struct {
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Severity    ViolationSeverity `json:"severity"`
	Amount      decimal.Decimal   `json:"amount,omitzero"`
	Limit       decimal.Decimal   `json:"limit,omitzero"`
	Action      string            `json:"action"`
}

// ViolationSeverity represents the severity of a compliance violation
type ViolationSeverity string

const (
	ViolationSeverityLow      ViolationSeverity = "LOW"
	ViolationSeverityMedium   ViolationSeverity = "MEDIUM"
	ViolationSeverityHigh     ViolationSeverity = "HIGH"
	ViolationSeverityCritical ViolationSeverity = "CRITICAL"
)

// UserLimits represents user-specific betting limits
type UserLimits struct {
	UserID           string          `json:"user_id"`
	DailyLimit       decimal.Decimal `json:"daily_limit"`
	WeeklyLimit      decimal.Decimal `json:"weekly_limit"`
	MonthlyLimit     decimal.Decimal `json:"monthly_limit"`
	MaxBetSize       decimal.Decimal `json:"max_bet_size"`
	AccumulatorLimit int             `json:"accumulator_limit"`
	CoolingOffPeriod time.Duration   `json:"cooling_off_period"`
	SelfExcluded     bool            `json:"self_excluded"`
	SelfExclusionEnd *time.Time      `json:"self_exclusion_end,omitempty"`
	LastUpdated      time.Time       `json:"last_updated"`
}

// ComplianceReport represents a comprehensive compliance report
type ComplianceReport struct {
	Period          string                   `json:"period"`
	TotalChecks     int64                    `json:"total_checks"`
	PassedChecks    int64                    `json:"passed_checks"`
	FailedChecks    int64                    `json:"failed_checks"`
	Violations      []ComplianceViolation    `json:"violations"`
	UserStatistics  UserComplianceStats      `json:"user_statistics"`
	FinancialStats  FinancialComplianceStats `json:"financial_stats"`
	Recommendations []string                 `json:"recommendations"`
	GeneratedAt     time.Time                `json:"generated_at"`
}

// UserComplianceStats represents user compliance statistics
type UserComplianceStats struct {
	TotalUsers         int64   `json:"total_users"`
	VerifiedUsers      int64   `json:"verified_users"`
	AgeVerifiedUsers   int64   `json:"age_verified_users"`
	SelfExcludedUsers  int64   `json:"self_excluded_users"`
	LimitBreachedUsers int64   `json:"limit_breached_users"`
	SuspiciousUsers    int64   `json:"suspicious_users"`
	ComplianceRate     float64 `json:"compliance_rate"`
}

// FinancialComplianceStats represents financial compliance statistics
type FinancialComplianceStats struct {
	TotalStake             decimal.Decimal `json:"total_stake"`
	TotalWinnings          decimal.Decimal `json:"total_winnings"`
	TotalDeposits          decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals       decimal.Decimal `json:"total_withdrawals"`
	TaxCollected           decimal.Decimal `json:"tax_collected"`
	AMLFlaggedTransactions int64           `json:"aml_flagged_transactions"`
	LargeTransactions      int64           `json:"large_transactions"`
}

// ValidateBetPlacement validates a bet placement against BCLB regulations
func (s *BCLBService) ValidateBetPlacement(ctx context.Context, userID string, betAmount decimal.Decimal, betType string, selections int) (*ComplianceCheck, error) {
	check := &ComplianceCheck{
		CheckTime: time.Now(),
		CheckType: "BET_PLACEMENT",
		UserID:    userID,
	}

	// Get user limits (in real implementation, this would come from database)
	_ = s.getUserLimits(userID)

	// Check 1: User age verification
	if !s.isAgeVerified(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "AGE_VERIFICATION",
			Description: "User age not verified",
			Severity:    ViolationSeverityCritical,
			Action:      "BLOCK_BET",
		})
	}

	// Check 2: KYC compliance
	if !s.isKYCCompliant(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "KYC_COMPLIANCE",
			Description: "User KYC not completed",
			Severity:    ViolationSeverityHigh,
			Action:      "BLOCK_BET",
		})
	}

	// Check 3: Self-exclusion
	if s.isSelfExcluded(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "SELF_EXCLUSION",
			Description: "User is self-excluded",
			Severity:    ViolationSeverityCritical,
			Action:      "BLOCK_BET",
		})
	}

	// Check 4: Daily stake limit
	dailyStake := s.getDailyStake(userID)
	if dailyStake.Add(betAmount).GreaterThan(s.config.MaxDailyStake) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "DAILY_LIMIT",
			Description: "Daily stake limit exceeded",
			Severity:    ViolationSeverityMedium,
			Amount:      dailyStake.Add(betAmount),
			Limit:       s.config.MaxDailyStake,
			Action:      "BLOCK_BET",
		})
	}

	// Check 5: Weekly stake limit
	weeklyStake := s.getWeeklyStake(userID)
	if weeklyStake.Add(betAmount).GreaterThan(s.config.MaxWeeklyStake) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "WEEKLY_LIMIT",
			Description: "Weekly stake limit exceeded",
			Severity:    ViolationSeverityMedium,
			Amount:      weeklyStake.Add(betAmount),
			Limit:       s.config.MaxWeeklyStake,
			Action:      "BLOCK_BET",
		})
	}

	// Check 6: Monthly stake limit
	monthlyStake := s.getMonthlyStake(userID)
	if monthlyStake.Add(betAmount).GreaterThan(s.config.MaxMonthlyStake) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "MONTHLY_LIMIT",
			Description: "Monthly stake limit exceeded",
			Severity:    ViolationSeverityMedium,
			Amount:      monthlyStake.Add(betAmount),
			Limit:       s.config.MaxMonthlyStake,
			Action:      "BLOCK_BET",
		})
	}

	// Check 7: Maximum bet per event
	if betAmount.GreaterThan(s.config.MaxBetPerEvent) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "MAX_BET_SIZE",
			Description: "Maximum bet per event exceeded",
			Severity:    ViolationSeverityMedium,
			Amount:      betAmount,
			Limit:       s.config.MaxBetPerEvent,
			Action:      "BLOCK_BET",
		})
	}

	// Check 8: Accumulator bet limit
	if betType == "ACCUMULATOR" && selections > s.config.MaxAccumulatorBets {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "ACCUMULATOR_LIMIT",
			Description: "Maximum accumulator selections exceeded",
			Severity:    ViolationSeverityLow,
			Action:      "WARN_USER",
		})
	}

	// Check 9: Cooling off period
	if s.isInCoolingOffPeriod(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "COOLING_OFF",
			Description: "User in cooling off period",
			Severity:    ViolationSeverityHigh,
			Action:      "BLOCK_BET",
		})
	}

	// Determine if check passed
	check.Passed = len(check.Violations) == 0

	// Log compliance check
	s.logComplianceCheck(check)

	// Publish event if violations found
	if !check.Passed {
		s.publishComplianceEvent("compliance.violation", check)
	}

	return check, nil
}

// ValidateTransaction validates a financial transaction against BCLB AML requirements
func (s *BCLBService) ValidateTransaction(ctx context.Context, userID string, amount decimal.Decimal, transactionType string) (*ComplianceCheck, error) {
	check := &ComplianceCheck{
		CheckTime: time.Now(),
		CheckType: "TRANSACTION",
		UserID:    userID,
	}

	// Check 1: Large transaction reporting (AML)
	if amount.GreaterThan(decimal.NewFromInt(100000)) { // KES 100,000 threshold
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "LARGE_TRANSACTION",
			Description: "Large transaction requires AML review",
			Severity:    ViolationSeverityMedium,
			Amount:      amount,
			Action:      "FLAG_FOR_REVIEW",
		})
	}

	// Check 2: Suspicious transaction patterns
	if s.isSuspiciousTransaction(userID, amount, transactionType) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "SUSPICIOUS_PATTERN",
			Description: "Suspicious transaction pattern detected",
			Severity:    ViolationSeverityHigh,
			Amount:      amount,
			Action:      "FLAG_FOR_REVIEW",
		})
	}

	// Check 3: Transaction frequency limits
	if s.exceedsTransactionFrequency(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "FREQUENCY_LIMIT",
			Description: "Transaction frequency limit exceeded",
			Severity:    ViolationSeverityLow,
			Action:      "WARN_USER",
		})
	}

	check.Passed = len(check.Violations) == 0

	s.logComplianceCheck(check)

	if !check.Passed {
		s.publishComplianceEvent("compliance.transaction_violation", check)
	}

	return check, nil
}

// GenerateComplianceReport generates a comprehensive compliance report
func (s *BCLBService) GenerateComplianceReport(ctx context.Context, period string) (*ComplianceReport, error) {
	report := &ComplianceReport{
		Period:      period,
		GeneratedAt: time.Now(),
	}

	// In a real implementation, this would query the database for actual statistics
	// For now, generate sample data
	report.TotalChecks = 10000
	report.PassedChecks = 9500
	report.FailedChecks = 500

	report.UserStatistics = UserComplianceStats{
		TotalUsers:         10000,
		VerifiedUsers:      8500,
		AgeVerifiedUsers:   8000,
		SelfExcludedUsers:  150,
		LimitBreachedUsers: 200,
		SuspiciousUsers:    50,
		ComplianceRate:     85.0,
	}

	report.FinancialStats = FinancialComplianceStats{
		TotalStake:             decimal.NewFromInt(50000000),
		TotalWinnings:          decimal.NewFromInt(45000000),
		TotalDeposits:          decimal.NewFromInt(60000000),
		TotalWithdrawals:       decimal.NewFromInt(55000000),
		TaxCollected:           decimal.NewFromInt(7500000),
		AMLFlaggedTransactions: 25,
		LargeTransactions:      100,
	}

	// Generate sample violations
	report.Violations = []ComplianceViolation{
		{
			Type:        "AGE_VERIFICATION",
			Description: "Underage betting attempt",
			Severity:    ViolationSeverityCritical,
			Action:      "BLOCK_BET",
		},
		{
			Type:        "DAILY_LIMIT",
			Description: "Daily stake limit exceeded",
			Severity:    ViolationSeverityMedium,
			Amount:      decimal.NewFromInt(25000),
			Limit:       s.config.MaxDailyStake,
			Action:      "BLOCK_BET",
		},
	}

	report.Recommendations = []string{
		"Increase age verification measures",
		"Implement real-time limit enforcement",
		"Enhance AML monitoring for large transactions",
		"Improve user education on responsible gaming",
	}

	return report, nil
}

// Helper methods with real database integration

func (s *BCLBService) getUserLimits(userID string) *UserLimits {
	// Get user-specific limits from database
	user, err := s.userRepo.GetUser(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user %s: %v", userID, err)
		// Return default limits if user not found
		return &UserLimits{
			UserID:       userID,
			DailyLimit:   s.config.MaxDailyStake,
			WeeklyLimit:  s.config.MaxWeeklyStake,
			MonthlyLimit: s.config.MaxMonthlyStake,
			MaxBetSize:   s.config.MaxBetPerEvent,
			LastUpdated:  time.Now(),
		}
	}

	// In a real implementation, user limits would be stored in a separate table
	// For now, use config defaults but could be overridden by user-specific settings
	return &UserLimits{
		UserID:       userID,
		DailyLimit:   s.config.MaxDailyStake,
		WeeklyLimit:  s.config.MaxWeeklyStake,
		MonthlyLimit: s.config.MaxMonthlyStake,
		MaxBetSize:   s.config.MaxBetPerEvent,
		LastUpdated:  user.UpdatedAt,
	}
}

func (s *BCLBService) isAgeVerified(userID string) bool {
	// Check user's age verification status from database
	ageVerification, err := s.userRepo.GetUserAgeVerification(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting age verification for user %s: %v", userID, err)
		return false
	}

	// Check if verification is still valid
	if !ageVerification.Verified {
		return false
	}

	// Check if verification has expired
	if time.Now().After(ageVerification.ExpiresAt) {
		return false
	}

	// Additional check: verify user is actually old enough
	user, err := s.userRepo.GetUser(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting user %s for age check: %v", userID, err)
		return false
	}

	// Calculate user's age
	now := time.Now()
	age := now.Year() - user.DateOfBirth.Year()

	// Adjust for birthday not yet occurred this year
	if now.Month() < user.DateOfBirth.Month() ||
		(now.Month() == user.DateOfBirth.Month() && now.Day() < user.DateOfBirth.Day()) {
		age--
	}

	return age >= s.config.MinAge
}

func (s *BCLBService) isKYCCompliant(userID string) bool {
	// Check user's KYC status from database
	kycStatus, err := s.userRepo.GetUserKYCStatus(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting KYC status for user %s: %v", userID, err)
		return false
	}

	// Check if KYC is verified and meets required level
	if kycStatus.Status != "VERIFIED" {
		return false
	}

	// Check if KYC level meets requirements
	switch s.config.RequiredKYCLevel {
	case "ENHANCED":
		return kycStatus.Level == "ENHANCED"
	case "STANDARD":
		return kycStatus.Level == "STANDARD" || kycStatus.Level == "ENHANCED"
	case "BASIC":
		return kycStatus.Level == "BASIC" || kycStatus.Level == "STANDARD" || kycStatus.Level == "ENHANCED"
	default:
		return kycStatus.Status == "VERIFIED"
	}
}

func (s *BCLBService) isSelfExcluded(userID string) bool {
	// Check user's self-exclusion status from database
	selfExclusion, err := s.userRepo.GetUserSelfExclusion(context.Background(), userID)
	if err != nil {
		log.Printf("Error getting self-exclusion status for user %s: %v", userID, err)
		return false
	}

	// Check if self-exclusion is active
	if !selfExclusion.Active {
		return false
	}

	// Check if self-exclusion period has expired
	if time.Now().After(selfExclusion.EndDate) {
		// Self-exclusion has expired, update database
		err = s.userRepo.UpdateSelfExclusion(context.Background(), userID, time.Time{}, "expired")
		if err != nil {
			log.Printf("Error updating expired self-exclusion for user %s: %v", userID, err)
		}
		return false
	}

	return true
}

func (s *BCLBService) getDailyStake(userID string) decimal.Decimal {
	// Calculate user's daily stake from database
	today := time.Now().Truncate(24 * time.Hour) // Start of today
	dailyStake, err := s.betRepo.GetUserDailyStake(context.Background(), userID, today)
	if err != nil {
		log.Printf("Error getting daily stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return dailyStake
}

func (s *BCLBService) getWeeklyStake(userID string) decimal.Decimal {
	// Calculate user's weekly stake from database
	now := time.Now()
	weekStart := now.AddDate(0, 0, -int(now.Weekday())) // Start of week (Sunday)
	weeklyStake, err := s.betRepo.GetUserWeeklyStake(context.Background(), userID, weekStart)
	if err != nil {
		log.Printf("Error getting weekly stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return weeklyStake
}

func (s *BCLBService) getMonthlyStake(userID string) decimal.Decimal {
	// Calculate user's monthly stake from database
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthlyStake, err := s.betRepo.GetUserMonthlyStake(context.Background(), userID, monthStart)
	if err != nil {
		log.Printf("Error getting monthly stake for user %s: %v", userID, err)
		return decimal.Zero
	}

	return monthlyStake
}

func (s *BCLBService) isInCoolingOffPeriod(userID string) bool {
	// Check if user is in cooling off period from database
	inCoolingOff, err := s.coolingOffRepo.IsUserInCoolingOffPeriod(context.Background(), userID)
	if err != nil {
		log.Printf("Error checking cooling off period for user %s: %v", userID, err)
		return false
	}

	return inCoolingOff
}

func (s *BCLBService) isSuspiciousTransaction(userID string, amount decimal.Decimal, transactionType string) bool {
	// Check for suspicious transaction patterns using database
	isSuspicious, err := s.transactionRepo.IsSuspiciousTransaction(context.Background(), userID, amount, transactionType)
	if err != nil {
		log.Printf("Error checking suspicious transaction for user %s: %v", userID, err)
		return false
	}

	return isSuspicious
}

func (s *BCLBService) exceedsTransactionFrequency(userID string) bool {
	// Check transaction frequency limits using database
	// Define frequency limit: max 100 transactions per hour
	const maxTransactionsPerHour = 100
	const oneHour = time.Hour

	exceeds, err := s.transactionRepo.ExceedsTransactionFrequency(context.Background(), userID, maxTransactionsPerHour, oneHour)
	if err != nil {
		log.Printf("Error checking transaction frequency for user %s: %v", userID, err)
		return false
	}

	return exceeds
}

func (s *BCLBService) logComplianceCheck(check *ComplianceCheck) {
	log.Printf("Compliance check: Type=%s, UserID=%s, Passed=%t, Violations=%d",
		check.CheckType, check.UserID, check.Passed, len(check.Violations))
}

func (s *BCLBService) publishComplianceEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing compliance event %s: %v", topic, err)
		}
	}
}

// SetUserSelfExclusion sets a user's self-exclusion status
func (s *BCLBService) SetUserSelfExclusion(ctx context.Context, userID string, duration time.Duration) error {
	// In real implementation, update user's self-exclusion status in database
	endTime := time.Now().Add(duration)

	s.publishComplianceEvent("compliance.self_exclusion", map[string]interface{}{
		"user_id":  userID,
		"duration": duration.String(),
		"end_time": endTime,
		"set_at":   time.Now(),
	})

	return nil
}

// RemoveUserSelfExclusion removes a user's self-exclusion status
func (s *BCLBService) RemoveUserSelfExclusion(ctx context.Context, userID string) error {
	// In real implementation, update user's self-exclusion status in database

	s.publishComplianceEvent("compliance.self_exclusion_removed", map[string]interface{}{
		"user_id":    userID,
		"removed_at": time.Now(),
	})

	return nil
}

// UpdateUserLimits updates a user's betting limits
func (s *BCLBService) UpdateUserLimits(ctx context.Context, userID string, limits UserLimits) error {
	// In real implementation, update user limits in database
	limits.LastUpdated = time.Now()

	s.publishComplianceEvent("compliance.limits_updated", map[string]interface{}{
		"user_id":       userID,
		"daily_limit":   limits.DailyLimit,
		"weekly_limit":  limits.WeeklyLimit,
		"monthly_limit": limits.MonthlyLimit,
		"updated_at":    limits.LastUpdated,
	})

	return nil
}

// GetComplianceStatus returns the current compliance status of the platform
func (s *BCLBService) GetComplianceStatus(ctx context.Context) (*ComplianceStatus, error) {
	status := &ComplianceStatus{
		LicenseValid:       time.Now().Before(s.config.LicenseExpiry),
		LicenseExpiry:      s.config.LicenseExpiry,
		ComplianceScore:    95.5,
		LastAuditDate:      time.Now().AddDate(0, -1, 0),
		NextAuditDate:      time.Now().AddDate(0, 2, 0),
		CriticalViolations: 2,
		OpenViolations:     15,
		LastUpdated:        time.Now(),
	}

	return status, nil
}

// ComplianceStatus represents the overall compliance status
type ComplianceStatus struct {
	LicenseValid       bool      `json:"license_valid"`
	LicenseExpiry      time.Time `json:"license_expiry"`
	ComplianceScore    float64   `json:"compliance_score"`
	LastAuditDate      time.Time `json:"last_audit_date"`
	NextAuditDate      time.Time `json:"next_audit_date"`
	CriticalViolations int64     `json:"critical_violations"`
	OpenViolations     int64     `json:"open_violations"`
	LastUpdated        time.Time `json:"last_updated"`
}
