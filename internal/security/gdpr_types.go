package security

import (
	"time"

	"github.com/shopspring/decimal"
)

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
