package compliance

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/shopspring/decimal"
)

// ValidateTransaction validates a transaction against BCLB regulations
func (s *BCLBService) ValidateTransaction(ctx context.Context, userID string, amount decimal.Decimal, transactionType string) (*ComplianceCheck, error) {
	check := &ComplianceCheck{
		CheckTime:     time.Now(),
		CheckType:     "TRANSACTION",
		UserID:        userID,
		TransactionID: generateTransactionID(),
	}

	// Check for suspicious transaction patterns
	if s.isSuspiciousTransaction(userID, amount, transactionType) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "SUSPICIOUS_TRANSACTION",
			Description: "Transaction pattern flagged as suspicious",
			Severity:    ViolationSeverityHigh,
			Amount:      amount,
			Action:      "FLAG_FOR_REVIEW",
		})
	}

	// Check transaction frequency
	if s.exceedsTransactionFrequency(userID) {
		check.Violations = append(check.Violations, ComplianceViolation{
			Type:        "TRANSACTION_FREQUENCY_EXCEEDED",
			Description: "Transaction frequency exceeds allowed limits",
			Severity:    ViolationSeverityMedium,
			Action:      "FLAG_FOR_REVIEW",
		})
	}

	check.Passed = len(check.Violations) == 0

	// Log compliance check
	s.logComplianceCheck(check)

	// Publish compliance event
	s.publishComplianceEvent("compliance.transaction_validated", check)

	return check, nil
}

// logComplianceCheck logs compliance check results
func (s *BCLBService) logComplianceCheck(check *ComplianceCheck) {
	log.Printf("Compliance check: Type=%s, UserID=%s, Passed=%t, Violations=%d",
		check.CheckType, check.UserID, check.Passed, len(check.Violations))
}

// publishComplianceEvent publishes compliance events
func (s *BCLBService) publishComplianceEvent(topic string, data any) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing compliance event %s: %v", topic, err)
		}
	}
}

var complianceGenerator *id.SnowflakeGenerator

func init() {
	var err error
	complianceGenerator, err = id.ServiceTypeGenerator("compliance")
	if err != nil {
		panic(fmt.Sprintf("Failed to create compliance ID generator: %v", err))
	}
}

// generateTransactionID generates a unique time-based deterministic transaction ID
func generateTransactionID() string {
	return fmt.Sprintf("txn_%s", complianceGenerator.GenerateID())
}
