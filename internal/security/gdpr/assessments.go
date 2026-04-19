package security

import (
	"context"
	"time"
)

// assessDataProcessing assesses data processing activities
func assessDataProcessing(context.Context) []DataProcessing {
	return []DataProcessing{
		{
			ID:          generateID(),
			Type:        "User Registration",
			Purpose:     "Account creation and management",
			LegalBasis:  "Consent",
			Categories:  []string{"Personal Data", "Contact Information"},
			Recipients:  []string{"Internal Systems"},
			Retention:   "7 years",
			Location:    "EU",
			Status:      "Active",
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Type:        "Betting Transactions",
			Purpose:     "Processing bets and payments",
			LegalBasis:  "Contractual Necessity",
			Categories:  []string{"Financial Data", "Betting History"},
			Recipients:  []string{"Payment Processors", "Internal Systems"},
			Retention:   "10 years",
			Location:    "EU",
			Status:      "Active",
			LastUpdated: time.Now(),
		},
	}
}

// assessDataSubjects assesses data subject information
func assessDataSubjects(context.Context) []DataSubject {
	return []DataSubject{
		{
			ID:           generateID(),
			Type:         "Customers",
			Category:     "Individual",
			Count:        50000,
			Location:     "EU",
			ConsentGiven: true,
			LastActivity: time.Now().AddDate(0, -1, 0),
		},
		{
			ID:           generateID(),
			Type:         "Employees",
			Category:     "Individual",
			Count:        100,
			Location:     "EU",
			ConsentGiven: true,
			LastActivity: time.Now().AddDate(0, -7, 0),
		},
	}
}

// assessGDPRRights assesses GDPR rights implementation
func assessGDPRRights(context.Context) []GDPRRight {
	return []GDPRRight{
		{
			ID:          generateID(),
			Type:        "Right to Access",
			Status:      "Implemented",
			Description: "Users can request access to their personal data",
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Type:        "Right to Rectification",
			Status:      "Implemented",
			Description: "Users can correct inaccurate personal data",
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Type:        "Right to Erasure",
			Status:      "Implemented",
			Description: "Users can request deletion of their personal data",
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
		{
			ID:          generateID(),
			Type:        "Right to Data Portability",
			Status:      "Partially Implemented",
			Description: "Users can export their data in machine-readable format",
			ProcessTime: "30 days",
			LastUpdated: time.Now(),
		},
	}
}

// assessBreachHistory assesses data breach history
func assessBreachHistory(context.Context) []DataBreach {
	return []DataBreach{
		{
			ID:           generateID(),
			Type:         "Phishing Attack",
			Severity:     SeverityMedium,
			Description:  "Employee credentials compromised through phishing",
			Affected:     5,
			Reported:     time.Now().AddDate(0, -6, 0),
			Resolved:     time.Now().AddDate(0, -5, 0),
			Notification: "Affected users notified within 72 hours",
			Status:       FindingStatusResolved,
		},
		{
			ID:           generateID(),
			Type:         "System Vulnerability",
			Severity:     SeverityLow,
			Description:  "Minor security vulnerability patched",
			Affected:     0,
			Reported:     time.Now().AddDate(0, -3, 0),
			Resolved:     time.Now().AddDate(0, -3, 0),
			Notification: "No notification required",
			Status:       FindingStatusResolved,
		},
	}
}

// assessConsentRecords assesses consent records
func assessConsentRecords(context.Context) []ConsentRecord {
	return []ConsentRecord{
		{
			ID:        generateID(),
			UserID:    "user_123",
			Type:      "Marketing",
			Purpose:   "Email marketing communications",
			Status:    "Active",
			GivenAt:   time.Now().AddDate(-1, 0, 0),
			ExpiresAt: time.Now().AddDate(1, 0, 0),
			IPAddress: "192.168.1.100",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
		{
			ID:        generateID(),
			UserID:    "user_456",
			Type:      "Analytics",
			Purpose:   "Website usage analytics",
			Status:    "Active",
			GivenAt:   time.Now().AddDate(-2, 0, 0),
			ExpiresAt: time.Now().AddDate(2, 0, 0),
			IPAddress: "192.168.1.101",
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		},
	}
}

// assessGDPRViolations assesses GDPR compliance violations
func assessGDPRViolations(context.Context) []GDPRViolation {
	return []GDPRViolation{
		{
			ID:          generateID(),
			Type:        "Missing Consent",
			Description: "Processing data without explicit consent",
			Severity:    SeverityHigh,
			Article:     "Article 6",
			Date:        time.Now().AddDate(0, -2, 0),
			Status:      FindingStatusOpen,
		},
		{
			ID:          generateID(),
			Type:        "Delayed Response",
			Description: "DSAR response exceeded 30-day deadline",
			Severity:    SeverityMedium,
			Article:     "Article 12",
			Date:        time.Now().AddDate(0, -1, 0),
			Status:      FindingStatusInProgress,
		},
		{
			ID:          generateID(),
			Type:        "Insufficient Documentation",
			Description: "Lack of proper documentation for data processing",
			Severity:    SeverityLow,
			Article:     "Article 30",
			Date:        time.Now().AddDate(0, -6, 0),
			Status:      FindingStatusResolved,
			Resolved:    time.Now().AddDate(0, -5, 0),
		},
	}
}
