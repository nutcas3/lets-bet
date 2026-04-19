package security

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// _assessSelfExclusion assesses self-exclusion records
func _assessSelfExclusion(context.Context) []SelfExclusionRecord {
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

// _assessDepositLimits assesses deposit limit records
func _assessDepositLimits(context.Context) []DepositLimit {
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

// _assessBettingLimits assesses betting limit records
func _assessBettingLimits(context.Context) []BettingLimit {
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
			UserID:     "user_113",
			Type:       "Daily",
			Amount:     decimal.NewFromInt(1000),
			Period:     "Daily",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-3, 0, 0),
			ModifiedAt: time.Now().AddDate(-3, 0, 0),
		},
	}
}

// _assessTimeLimits assesses time limit records
func _assessTimeLimits(context.Context) []TimeLimit {
	return []TimeLimit{
		{
			ID:         generateID(),
			UserID:     "user_124",
			Type:       "Session Duration",
			Duration:   2 * time.Hour,
			Period:     "Per Session",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-1, 0, 0),
			ModifiedAt: time.Now().AddDate(-1, 0, 0),
		},
		{
			ID:         generateID(),
			UserID:     "user_125",
			Type:       "Daily",
			Duration:   4 * time.Hour,
			Period:     "Daily",
			Status:     "Active",
			CreatedAt:  time.Now().AddDate(-2, 0, 0),
			ModifiedAt: time.Now().AddDate(-2, 0, 0),
		},
	}
}

// _assessCoolingOffPeriods assesses cooling off period records
func _assessCoolingOffPeriods(context.Context) []CoolingOffRecord {
	return []CoolingOffRecord{
		{
			ID:        generateID(),
			UserID:    "user_126",
			Duration:  24 * time.Hour,
			StartDate: time.Now().Add(-2 * time.Hour),
			EndDate:   time.Now().Add(22 * time.Hour),
			Status:    "Active",
			Reason:    "User requested cooling off",
		},
		{
			ID:        generateID(),
			UserID:    "user_127",
			Duration:  7 * 24 * time.Hour,
			StartDate: time.Now().Add(-24 * time.Hour),
			EndDate:   time.Now().Add(6 * 24 * time.Hour),
			Status:    "Active",
			Reason:    "Loss limit reached",
		},
	}
}

// _assessInterventions assesses intervention records
func _assessInterventions(context.Context) []Intervention {
	return []Intervention{
		{
			ID:      generateID(),
			UserID:  "user_128",
			Type:    "Automated",
			Trigger: "Daily limit reached",
			Action:  "Session terminated",
			Outcome: "User contacted support",
			Date:    time.Now().Add(-24 * time.Hour),
			Agent:   "System",
		},
		{
			ID:      generateID(),
			UserID:  "user_129",
			Type:    "Manual",
			Trigger: "Suspicious betting pattern",
			Action:  "Account review",
			Outcome: "Limits adjusted",
			Date:    time.Now().Add(-48 * time.Hour),
			Agent:   "Support Agent",
		},
	}
}

// _assessEducation assesses education materials
func _assessEducation(context.Context) []EducationMaterial {
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
			Content:   "Video tutorial on how to set responsible betting limits...",
			Language:  "English",
			Views:     800,
			CreatedAt: time.Now().AddDate(-3, 0, 0),
			UpdatedAt: time.Now().AddDate(-3, 0, 0),
		},
	}
}

// _assessViolations assesses responsible gaming violations
func _assessViolations(context.Context) []RGViolation {
	return []RGViolation{
		{
			ID:          generateID(),
			Type:        "Underage Betting",
			Description: "Attempted betting by underage user",
			Severity:    SeverityCritical,
			UserID:      "user_130",
			Date:        time.Now().Add(-72 * time.Hour),
			Status:      FindingStatusResolved,
			Resolved:    time.Now().Add(-71 * time.Hour),
		},
		{
			ID:          generateID(),
			Type:        "Excessive Betting",
			Description: "User betting more than reasonable limits",
			Severity:    SeverityMedium,
			UserID:      "user_131",
			Date:        time.Now().Add(-48 * time.Hour),
			Status:      FindingStatusOpen,
		},
		{
			ID:          generateID(),
			Type:        "Problem Gambling Indicators",
			Description: "User showing signs of problem gambling",
			Severity:    SeverityHigh,
			UserID:      "user_132",
			Date:        time.Now().Add(-24 * time.Hour),
			Status:      FindingStatusInProgress,
		},
	}
}
