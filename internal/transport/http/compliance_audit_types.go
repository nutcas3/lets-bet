package http

import (
	"time"
)

// ComplianceAudit represents a compliance audit
type ComplianceAudit struct {
	AuditID     string          `json:"audit_id"`
	Type        string          `json:"type"`
	Status      string          `json:"status"`
	Scope       []string        `json:"scope"`
	StartedAt   time.Time       `json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	ConductedBy string          `json:"conducted_by"`
	Findings    []*AuditFinding `json:"findings"`
	Score       *AuditScore     `json:"score"`
	Report      *AuditReport    `json:"report,omitempty"`
}

// AuditFinding represents an audit finding
type AuditFinding struct {
	ID             string     `json:"id"`
	Category       string     `json:"category"`
	Severity       string     `json:"severity"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Impact         string     `json:"impact"`
	Recommendation string     `json:"recommendation"`
	Status         string     `json:"status"`
	FoundAt        time.Time  `json:"found_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy     string     `json:"resolved_by,omitempty"`
}

// AuditScore represents audit score
type AuditScore struct {
	Overall     int               `json:"overall"`
	MaxScore    int               `json:"max_score"`
	Grade       string            `json:"grade"`
	Categories  map[string]int    `json:"categories"`
	Breakdown   []*ScoreBreakdown `json:"breakdown"`
	LastUpdated time.Time         `json:"last_updated"`
}

// ScoreBreakdown represents score breakdown
type ScoreBreakdown struct {
	Category    string `json:"category"`
	Score       int    `json:"score"`
	MaxScore    int    `json:"max_score"`
	Weight      int    `json:"weight"`
	Description string `json:"description"`
}

// AuditReport represents audit report
type AuditReport struct {
	ReportID    string    `json:"report_id"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	Content     string    `json:"content"`
	GeneratedAt time.Time `json:"generated_at"`
	GeneratedBy string    `json:"generated_by"`
	Format      string    `json:"format"`
}
