package security

import (
	"fmt"
	"time"

	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/shopspring/decimal"
)

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data any) error
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

// SecurityCategory represents different security categories
type SecurityCategory string

const (
	CategoryAuthentication SecurityCategory = "AUTHENTICATION"
	CategoryAuthorization  SecurityCategory = "AUTHORIZATION"
	CategoryDataProtection SecurityCategory = "DATA_PROTECTION"
	CategoryNetwork        SecurityCategory = "NETWORK"
	CategoryApplication    SecurityCategory = "APPLICATION"
	CategoryInfrastructure SecurityCategory = "INFRASTRUCTURE"
	CategoryCompliance     SecurityCategory = "COMPLIANCE"
)

// FindingStatus represents the status of a security finding
type FindingStatus string

const (
	FindingStatusOpen       FindingStatus = "OPEN"
	FindingStatusInProgress FindingStatus = "IN_PROGRESS"
	FindingStatusResolved   FindingStatus = "RESOLVED"
	FindingStatusAccepted   FindingStatus = "ACCEPTED"
)

// TestType represents different types of penetration tests
type TestType string

const (
	TestTypeBlackBox TestType = "BLACK_BOX"
	TestTypeWhiteBox TestType = "WHITE_BOX"
	TestTypeGrayBox  TestType = "GRAY_BOX"
	TestTypeWebApp   TestType = "WEB_APP"
	TestTypeMobile   TestType = "MOBILE"
	TestTypeNetwork  TestType = "NETWORK"
	TestTypeSocial   TestType = "SOCIAL"
)

// TestStatus represents the status of a penetration test
type TestStatus string

const (
	TestStatusPlanned    TestStatus = "PLANNED"
	TestStatusInProgress TestStatus = "IN_PROGRESS"
	TestStatusCompleted  TestStatus = "COMPLETED"
	TestStatusFailed     TestStatus = "FAILED"
	TestStatusCancelled  TestStatus = "CANCELLED"
)

// AuditStatus represents the status of a security audit
type AuditStatus string

const (
	AuditStatusPending    AuditStatus = "PENDING"
	AuditStatusInProgress AuditStatus = "IN_PROGRESS"
	AuditStatusCompleted  AuditStatus = "COMPLETED"
	AuditStatusFailed     AuditStatus = "FAILED"
)

// PenTestFinding represents a finding from penetration testing
type PenTestFinding struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Severity    SeverityLevel    `json:"severity"`
	Category    SecurityCategory `json:"category"`
	Endpoint    string           `json:"endpoint"`
	Payload     string           `json:"payload"`
	Evidence    string           `json:"evidence"`
	Impact      string           `json:"impact"`
	Remediation string           `json:"remediation"`
	CVSSScore   float64          `json:"cvss_score"`
	Discovered  time.Time        `json:"discovered"`
	Status      FindingStatus    `json:"status"`
}

// PenetrationTest represents a penetration test report
type PenetrationTest struct {
	ID              string           `json:"id"`
	TestType        TestType         `json:"test_type"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	Status          TestStatus       `json:"status"`
	Testers         []string         `json:"testers"`
	Scope           []string         `json:"scope"`
	Findings        []PenTestFinding `json:"findings"`
	RiskScore       int              `json:"risk_score"`
	Recommendations []string         `json:"recommendations"`
	ReportURL       string           `json:"report_url"`
	NextTestDate    time.Time        `json:"next_test_date"`
}

// GDPRCompliance represents GDPR compliance status
type GDPRCompliance struct {
	ID              string           `json:"id"`
	ComplianceScore float64          `json:"compliance_score"`
	LastAssessment  time.Time        `json:"last_assessment"`
	NextAssessment  time.Time        `json:"next_assessment"`
	DataProcessing  []DataProcessing `json:"data_processing"`
	DataSubjects    []DataSubject    `json:"data_subjects"`
	Rights          []GDPRRight      `json:"rights"`
	BreachHistory   []DataBreach     `json:"breach_history"`
	ConsentRecords  []ConsentRecord  `json:"consent_records"`
	Violations      []GDPRViolation  `json:"violations"`
	Recommendations []string         `json:"recommendations"`
	DPOContact      string           `json:"dpo_contact"`
}

// DataProcessing represents data processing activities
type DataProcessing struct {
	ID          string    `json:"id"`
	Purpose     string    `json:"purpose"`
	Categories  []string  `json:"categories"`
	DataTypes   []string  `json:"data_types"`
	LegalBasis  string    `json:"legal_basis"`
	Retention   string    `json:"retention"`
	Recipients  []string  `json:"recipients"`
	Transfers   []string  `json:"transfers"`
	Security    []string  `json:"security_measures"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataSubject represents data subject information
type DataSubject struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Categories  []string  `json:"categories"`
	Count       int64     `json:"count"`
	LastUpdated time.Time `json:"last_updated"`
}

// GDPRRight represents GDPR rights implementation
type GDPRRight struct {
	Right       string    `json:"right"`
	Implemented bool      `json:"implemented"`
	ProcessTime string    `json:"process_time"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataBreach represents a data breach record
type DataBreach struct {
	ID         string    `json:"id"`
	Date       time.Time `json:"date"`
	Type       string    `json:"type"`
	Affected   int64     `json:"affected"`
	Categories []string  `json:"categories"`
	Cause      string    `json:"cause"`
	Impact     string    `json:"impact"`
	Notified   bool      `json:"notified"`
	Reported   bool      `json:"reported"`
	Resolved   bool      `json:"resolved"`
}

// ConsentRecord represents consent record
type ConsentRecord struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Purpose   string    `json:"purpose"`
	Granted   bool      `json:"granted"`
	Date      time.Time `json:"date"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Withdrawn time.Time `json:"withdrawn"`
}

// GDPRViolation represents GDPR compliance violations
type GDPRViolation struct {
	ID          string          `json:"id"`
	Article     string          `json:"article"`
	Description string          `json:"description"`
	Severity    SeverityLevel   `json:"severity"`
	Fine        decimal.Decimal `json:"fine"`
	Status      FindingStatus   `json:"status"`
	Discovered  time.Time       `json:"discovered"`
	Resolved    time.Time       `json:"resolved"`
}

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
	Resolved    time.Time     `json:"resolved"`
}

var securityGenerator *id.SnowflakeGenerator

func init() {
	var err error
	securityGenerator, err = id.ServiceTypeGenerator("security")
	if err != nil {
		panic(fmt.Sprintf("Failed to create security ID generator: %v", err))
	}
}

// generateID generates a unique time-based deterministic ID
func generateID() string {
	return securityGenerator.GenerateID()
}
