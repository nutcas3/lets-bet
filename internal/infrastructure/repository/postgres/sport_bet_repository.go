package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/shopspring/decimal"
)

// SportBetRepository implements sport bet repository using PostgreSQL
type SportBetRepository struct {
	db *sql.DB
}

// NewSportBetRepository creates a new sport bet repository
func NewSportBetRepository(db *sql.DB) *SportBetRepository {
	return &SportBetRepository{db: db}
}

// Create creates a new sport bet
func (r *SportBetRepository) Create(ctx context.Context, bet *domain.SportBet) error {
	query := `
		INSERT INTO sport_bets (
			id, user_id, event_id, market_id, outcome_id, amount, odds,
			currency, status, payout, net_payout, settled_at,
			settlement_reason, placed_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.ExecContext(ctx, query,
		bet.ID, bet.UserID, bet.EventID, bet.MarketID, bet.OutcomeID,
		bet.Amount, bet.Odds, "KES", string(bet.Status), bet.Payout,
		bet.NetPayout, bet.SettledAt, "", bet.PlacedAt, time.Now(),
	)

	return err
}

// GetByID retrieves a sport bet by ID
func (r *SportBetRepository) GetByID(ctx context.Context, id string) (*domain.SportBet, error) {
	query := `
		SELECT id, user_id, event_id, market_id, outcome_id, amount, odds,
			   currency, status, payout, net_payout, settled_at,
			   settlement_reason, placed_at, updated_at
		FROM sport_bets
		WHERE id = $1
	`

	var bet domain.SportBet
	var currency, status, settlementReason string
	var settledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&bet.ID, &bet.UserID, &bet.EventID, &bet.MarketID, &bet.OutcomeID,
		&bet.Amount, &bet.Odds, &currency, &status, &bet.Payout,
		&bet.NetPayout, &settledAt, &settlementReason, &bet.PlacedAt, &bet.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	bet.Status = domain.BetStatus(status)
	bet.SettledAt = &settledAt.Time

	return &bet, nil
}

// Update updates an existing sport bet
func (r *SportBetRepository) Update(ctx context.Context, bet *domain.SportBet) error {
	query := `
		UPDATE sport_bets SET
			user_id = $2, event_id = $3, market_id = $4, outcome_id = $5,
			amount = $6, odds = $7, status = $8, payout = $9,
			net_payout = $10, settled_at = $11, settlement_reason = $12,
			updated_at = $13
		WHERE id = $1
	`

	var settlementReason *string
	if bet.SettlementReason != "" {
		settlementReason = &bet.SettlementReason
	}

	var settledAt sql.NullTime
	if bet.SettledAt != nil && !bet.SettledAt.IsZero() {
		settledAt = sql.NullTime{Time: *bet.SettledAt, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		bet.ID, bet.UserID, bet.EventID, bet.MarketID, bet.OutcomeID,
		bet.Amount, bet.Odds, string(bet.Status), bet.Payout,
		bet.NetPayout, settledAt, settlementReason, time.Now(),
	)

	return err
}

// Delete deletes a sport bet
func (r *SportBetRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM sport_bets WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

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

	var bets []*domain.SportBet
	for rows.Next() {
		var bet domain.SportBet
		var currency, statusStr, settlementReason string
		var settledAt sql.NullTime

		err := rows.Scan(
			&bet.ID, &bet.UserID, &bet.EventID, &bet.MarketID, &bet.OutcomeID,
			&bet.Amount, &bet.Odds, &currency, &statusStr, &bet.Payout,
			&bet.NetPayout, &settledAt, &settlementReason, &bet.PlacedAt, &bet.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		bet.Status = domain.BetStatus(statusStr)
		if settledAt.Valid {
			bet.SettledAt = &settledAt.Time
		}
		bets = append(bets, &bet)
	}

	return bets, nil
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
