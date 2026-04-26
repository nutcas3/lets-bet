package postgres

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// GetMetrics returns jackpot statistics
func (r *JackpotRepository) GetMetrics(ctx context.Context) (*JackpotMetrics, error) {
	// Get total jackpots
	var totalJackpots int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jackpots").Scan(&totalJackpots)
	if err != nil {
		return nil, err
	}

	// Get active jackpots
	var activeJackpots int64
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jackpots WHERE status = 'ACTIVE'").Scan(&activeJackpots)
	if err != nil {
		return nil, err
	}

	// Get total tickets
	var totalTickets int64
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jackpot_tickets").Scan(&totalTickets)
	if err != nil {
		return nil, err
	}

	// Get active tickets
	var activeTickets int64
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM jackpot_tickets WHERE status = 'ACTIVE'").Scan(&activeTickets)
	if err != nil {
		return nil, err
	}

	// Get total contributions
	var totalContributions decimal.Decimal
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(amount), 0) FROM jackpot_tickets").Scan(&totalContributions)
	if err != nil {
		return nil, err
	}

	// Get total payouts
	var totalPayouts decimal.Decimal
	err = r.db.QueryRowContext(ctx, "SELECT COALESCE(SUM(prize_amount), 0) FROM jackpot_tickets WHERE status = 'WON'").Scan(&totalPayouts)
	if err != nil {
		return nil, err
	}

	// Calculate average ticket value
	var averageTicketValue decimal.Decimal
	if totalTickets > 0 {
		averageTicketValue = totalContributions.Div(decimal.NewFromInt(totalTickets))
	}

	return &JackpotMetrics{
		TotalJackpots:      totalJackpots,
		ActiveJackpots:     activeJackpots,
		TotalTickets:       totalTickets,
		ActiveTickets:      activeTickets,
		TotalContributions: totalContributions,
		TotalPayouts:       totalPayouts,
		AverageTicketValue: averageTicketValue,
		LastDrawTime:       time.Now(),
		NextDrawTime:       time.Now().Add(24 * time.Hour),
	}, nil
}
