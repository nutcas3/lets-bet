package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/betting-platform/internal/core/domain"
)

// GetByUserID retrieves bets by user ID with optional filters
func (r *SportBetRepository) GetByUserID(ctx context.Context, userID string, filters *BetFilters) ([]*domain.SportBet, error) {
	query := `
		SELECT id, user_id, event_id, market_id, outcome_id, amount, odds,
			   currency, status, payout, net_payout, settled_at,
			   settlement_reason, placed_at, updated_at
		FROM sport_bets
		WHERE user_id = $1
	`

	args := []any{userID}
	argIndex := 2

	if filters != nil {
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(*filters.Status))
			argIndex++
		}
		if filters.From != nil {
			query += fmt.Sprintf(" AND placed_at >= $%d", argIndex)
			args = append(args, *filters.From)
			argIndex++
		}
		if filters.To != nil {
			query += fmt.Sprintf(" AND placed_at <= $%d", argIndex)
			args = append(args, *filters.To)
			argIndex++
		}
	}

	query += " ORDER BY placed_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBets(rows)
}

// GetByEventID retrieves bets by event ID with optional filters
func (r *SportBetRepository) GetByEventID(ctx context.Context, eventID string, filters *BetFilters) ([]*domain.SportBet, error) {
	query := `
		SELECT id, user_id, event_id, market_id, outcome_id, amount, odds,
			   currency, status, payout, net_payout, settled_at,
			   settlement_reason, placed_at, updated_at
		FROM sport_bets
		WHERE event_id = $1
	`

	args := []any{eventID}
	argIndex := 2

	if filters != nil {
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(*filters.Status))
			argIndex++
		}
		if filters.From != nil {
			query += fmt.Sprintf(" AND placed_at >= $%d", argIndex)
			args = append(args, *filters.From)
			argIndex++
		}
		if filters.To != nil {
			query += fmt.Sprintf(" AND placed_at <= $%d", argIndex)
			args = append(args, *filters.To)
			argIndex++
		}
	}

	query += " ORDER BY placed_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBets(rows)
}

// GetByStatus retrieves bets by status with optional filters
func (r *SportBetRepository) GetByStatus(ctx context.Context, status domain.BetStatus, filters *BetFilters) ([]*domain.SportBet, error) {
	query := `
		SELECT id, user_id, event_id, market_id, outcome_id, amount, odds,
			   currency, status, payout, net_payout, settled_at,
			   settlement_reason, placed_at, updated_at
		FROM sport_bets
		WHERE status = $1
	`

	args := []any{string(status)}
	argIndex := 2

	if filters != nil {
		if filters.From != nil {
			query += fmt.Sprintf(" AND placed_at >= $%d", argIndex)
			args = append(args, *filters.From)
			argIndex++
		}
		if filters.To != nil {
			query += fmt.Sprintf(" AND placed_at <= $%d", argIndex)
			args = append(args, *filters.To)
			argIndex++
		}
	}

	query += " ORDER BY placed_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBets(rows)
}

// GetPendingBets retrieves all pending bets
func (r *SportBetRepository) GetPendingBets(ctx context.Context, filters *BetFilters) ([]*domain.SportBet, error) {
	return r.GetByStatus(ctx, domain.BetStatusPending, filters)
}

// GetSettledBets retrieves all settled bets
func (r *SportBetRepository) GetSettledBets(ctx context.Context, filters *BetFilters) ([]*domain.SportBet, error) {
	query := `
		SELECT id, user_id, event_id, market_id, outcome_id, amount, odds,
			   currency, status, payout, net_payout, settled_at,
			   settlement_reason, placed_at, updated_at
		FROM sport_bets
		WHERE status IN ('WON', 'LOST', 'VOID')
	`

	args := []any{}
	argIndex := 1

	if filters != nil {
		if filters.From != nil {
			query += fmt.Sprintf(" AND placed_at >= $%d", argIndex)
			args = append(args, *filters.From)
			argIndex++
		}
		if filters.To != nil {
			query += fmt.Sprintf(" AND placed_at <= $%d", argIndex)
			args = append(args, *filters.To)
			argIndex++
		}
	}

	query += " ORDER BY placed_at DESC"

	if filters != nil && filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filters.Limit)
		if filters.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", filters.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBets(rows)
}

// scanBets scans multiple bets from rows
func (r *SportBetRepository) scanBets(rows *sql.Rows) ([]*domain.SportBet, error) {
	var bets []*domain.SportBet

	for rows.Next() {
		var bet domain.SportBet
		var currency, status, settlementReason string
		var settledAt sql.NullTime

		err := rows.Scan(
			&bet.ID, &bet.UserID, &bet.EventID, &bet.MarketID, &bet.OutcomeID,
			&bet.Amount, &bet.Odds, &currency, &status, &bet.Payout,
			&bet.NetPayout, &settledAt, &settlementReason, &bet.PlacedAt, &bet.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		bet.Status = domain.BetStatus(status)
		bet.SettledAt = &settledAt.Time
		bets = append(bets, &bet)
	}

	return bets, nil
}
