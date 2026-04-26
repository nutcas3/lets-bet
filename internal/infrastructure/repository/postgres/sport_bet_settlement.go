package postgres

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// MarkAsWon marks a bet as won with payout
func (r *SportBetRepository) MarkAsWon(ctx context.Context, betID string, payout decimal.Decimal) error {
	query := `
		UPDATE sport_bets SET
			status = 'WON',
			payout = $2,
			net_payout = $2,
			settled_at = $3,
			updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, betID, payout, time.Now())
	return err
}

// MarkAsLost marks a bet as lost
func (r *SportBetRepository) MarkAsLost(ctx context.Context, betID string) error {
	query := `
		UPDATE sport_bets SET
			status = 'LOST',
			payout = 0,
			net_payout = 0,
			settled_at = $2,
			updated_at = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, betID, time.Now())
	return err
}

// MarkAsVoid marks a bet as void with reason
func (r *SportBetRepository) MarkAsVoid(ctx context.Context, betID string, reason string) error {
	query := `
		UPDATE sport_bets SET
			status = 'VOID',
			payout = 0,
			net_payout = 0,
			settlement_reason = $2,
			settled_at = $3,
			updated_at = $3
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, betID, reason, time.Now())
	return err
}
