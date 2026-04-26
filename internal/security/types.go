package security

import (
	"fmt"
	"time"

	"github.com/betting-platform/internal/infrastructure/id"
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
