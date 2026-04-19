package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// BetRepository implements bet repository using PostgreSQL
type BetRepository struct {
	db *sql.DB
}

func NewBetRepository(db *sql.DB) *BetRepository {
	return &BetRepository{db: db}
}

func (r *BetRepository) Create(ctx context.Context, bet *domain.Bet) error {
	query := `
		INSERT INTO bets (
			id, user_id, country_code, bet_type, stake, currency,
			potential_win, total_odds, status, actual_win,
			settled_at, placed_at, ip_address, device_id,
			tax_amount, tax_paid
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err := r.db.ExecContext(ctx, query,
		bet.ID, bet.UserID, bet.CountryCode, bet.BetType, bet.Stake, bet.Currency,
		bet.PotentialWin, bet.TotalOdds, bet.Status, bet.ActualWin,
		bet.SettledAt, bet.PlacedAt, bet.IPAddress, bet.DeviceID,
		bet.TaxAmount, bet.TaxPaid,
	)

	if err != nil {
		log.Printf("Error creating bet: %v", err)
		return err
	}

	return nil
}

func (r *BetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bet, error) {
	query := `
		SELECT id, user_id, country_code, bet_type, stake, currency,
			   potential_win, total_odds, status, actual_win,
			   settled_at, placed_at, ip_address, device_id,
			   tax_amount, tax_paid
		FROM bets WHERE id = $1
	`

	var bet domain.Bet
	var settledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&bet.ID, &bet.UserID, &bet.CountryCode, &bet.BetType, &bet.Stake, &bet.Currency,
		&bet.PotentialWin, &bet.TotalOdds, &bet.Status, &bet.ActualWin,
		&settledAt, &bet.PlacedAt, &bet.IPAddress, &bet.DeviceID,
		&bet.TaxAmount, &bet.TaxPaid,
	)

	if err != nil {
		return nil, err
	}

	if settledAt.Valid {
		bet.SettledAt = &settledAt.Time
	}

	return &bet, nil
}

func (r *BetRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Bet, error) {
	query := `
		SELECT id, user_id, country_code, bet_type, stake, currency,
			   potential_win, total_odds, status, actual_win,
			   settled_at, placed_at, ip_address, device_id,
			   tax_amount, tax_paid
		FROM bets WHERE user_id = $1
		ORDER BY placed_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []*domain.Bet

	for rows.Next() {
		var bet domain.Bet
		var settledAt sql.NullTime

		err := rows.Scan(
			&bet.ID, &bet.UserID, &bet.CountryCode, &bet.BetType, &bet.Stake, &bet.Currency,
			&bet.PotentialWin, &bet.TotalOdds, &bet.Status, &bet.ActualWin,
			&settledAt, &bet.PlacedAt, &bet.IPAddress, &bet.DeviceID,
			&bet.TaxAmount, &bet.TaxPaid,
		)

		if err != nil {
			return nil, err
		}

		if settledAt.Valid {
			bet.SettledAt = &settledAt.Time
		}

		bets = append(bets, &bet)
	}

	return bets, nil
}

func (r *BetRepository) GetPendingBets(ctx context.Context) ([]*domain.Bet, error) {
	query := `
		SELECT id, user_id, country_code, bet_type, stake, currency,
			   potential_win, total_odds, status, actual_win,
			   settled_at, placed_at, ip_address, device_id,
			   tax_amount, tax_paid
		FROM bets WHERE status = $1
		ORDER BY placed_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, domain.BetStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []*domain.Bet

	for rows.Next() {
		var bet domain.Bet
		var settledAt sql.NullTime

		err := rows.Scan(
			&bet.ID, &bet.UserID, &bet.CountryCode, &bet.BetType, &bet.Stake, &bet.Currency,
			&bet.PotentialWin, &bet.TotalOdds, &bet.Status, &bet.ActualWin,
			&settledAt, &bet.PlacedAt, &bet.IPAddress, &bet.DeviceID,
			&bet.TaxAmount, &bet.TaxPaid,
		)

		if err != nil {
			return nil, err
		}

		if settledAt.Valid {
			bet.SettledAt = &settledAt.Time
		}

		bets = append(bets, &bet)
	}

	return bets, nil
}

func (r *BetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BetStatus, actualWin decimal.Decimal) error {
	query := `
		UPDATE bets SET 
			status = $2, actual_win = $3, settled_at = $4
		WHERE id = $1
	`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, id, status, actualWin, now)
	if err != nil {
		log.Printf("Error updating bet status: %v", err)
		return err
	}

	return nil
}

func (r *BetRepository) UpdateTaxPaid(ctx context.Context, id uuid.UUID, taxAmount decimal.Decimal) error {
	query := `
		UPDATE bets SET 
			tax_amount = $2, tax_paid = true
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id, taxAmount)
	if err != nil {
		log.Printf("Error updating tax paid: %v", err)
		return err
	}

	return nil
}
