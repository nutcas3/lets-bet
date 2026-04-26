package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/betting-platform/internal/core/domain"
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
