package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/shopspring/decimal"
)

// JackpotRepository implements jackpot repository using PostgreSQL
type JackpotRepository struct {
	db *sql.DB
}

// NewJackpotRepository creates a new jackpot repository
func NewJackpotRepository(db *sql.DB) *JackpotRepository {
	return &JackpotRepository{db: db}
}

// CreateJackpot creates a new jackpot
func (r *JackpotRepository) CreateJackpot(ctx context.Context, jackpot *Jackpot) error {
	query := `
		INSERT INTO jackpots (
			id, name, type, current_amount, seed_amount, contribution_rate,
			min_bet, max_bet, status, created_at, updated_at, expires_at, next_draw_at,
			description, is_active, winning_numbers, winner_id, winner_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`

	_, err := r.db.ExecContext(ctx, query,
		jackpot.ID, jackpot.Name, string(jackpot.Type), jackpot.CurrentAmount,
		jackpot.SeedAmount, jackpot.ContributionRate, jackpot.MinBet, jackpot.MaxBet,
		string(jackpot.Status), jackpot.CreatedAt, jackpot.UpdatedAt, jackpot.ExpiresAt,
		jackpot.NextDrawAt, jackpot.Description, jackpot.IsActive,
		jackpot.WinningNumbers, jackpot.WinnerID, jackpot.WinnerAmount,
	)

	return err
}

// GetJackpot retrieves a jackpot by ID
func (r *JackpotRepository) GetJackpot(ctx context.Context, id string) (*Jackpot, error) {
	query := `
		SELECT id, name, type, current_amount, seed_amount, contribution_rate,
			   min_bet, max_bet, status, created_at, updated_at, expires_at, next_draw_at,
			   description, is_active, winning_numbers, winner_id, winner_amount
		FROM jackpots WHERE id = $1
	`

	var jackpot Jackpot
	var winningNumbers []int
	var winnerID *string
	var winnerAmount decimal.Decimal

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&jackpot.ID, &jackpot.Name, &jackpot.Type, &jackpot.CurrentAmount,
		&jackpot.SeedAmount, &jackpot.ContributionRate, &jackpot.MinBet, &jackpot.MaxBet,
		&jackpot.Status, &jackpot.CreatedAt, &jackpot.UpdatedAt, &jackpot.ExpiresAt,
		&jackpot.NextDrawAt, &jackpot.Description, &jackpot.IsActive,
		&winningNumbers, &winnerID, &winnerAmount,
	)

	if err != nil {
		return nil, err
	}

	jackpot.WinningNumbers = winningNumbers
	jackpot.WinnerID = winnerID
	jackpot.WinnerAmount = winnerAmount

	return &jackpot, nil
}

// GetJackpots retrieves jackpots with optional filters
func (r *JackpotRepository) GetJackpots(ctx context.Context, filters *JackpotFilters) ([]*Jackpot, error) {
	query := `
		SELECT id, name, type, current_amount, seed_amount, contribution_rate,
			   min_bet, max_bet, status, created_at, updated_at, expires_at, next_draw_at,
			   description, is_active, winning_numbers, winner_id, winner_amount
		FROM jackpots
		WHERE 1=1
	`

	args := []any{}
	argIndex := 1

	if filters != nil {
		if filters.Type != nil {
			query += fmt.Sprintf(" AND type = $%d", argIndex)
			args = append(args, string(*filters.Type))
			argIndex++
		}
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(*filters.Status))
			argIndex++
		}
		if filters.IsActive != nil {
			query += fmt.Sprintf(" AND is_active = $%d", argIndex)
			args = append(args, *filters.IsActive)
			argIndex++
		}
		if filters.From != nil {
			query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
			args = append(args, *filters.From)
			argIndex++
		}
		if filters.To != nil {
			query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
			args = append(args, *filters.To)
			argIndex++
		}
	}

	query += " ORDER BY created_at DESC"

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

	var jackpots []*Jackpot
	for rows.Next() {
		var jackpot Jackpot
		var winningNumbers []int
		var winnerID *string
		var winnerAmount decimal.Decimal

		err := rows.Scan(
			&jackpot.ID, &jackpot.Name, &jackpot.Type, &jackpot.CurrentAmount,
			&jackpot.SeedAmount, &jackpot.ContributionRate, &jackpot.MinBet, &jackpot.MaxBet,
			&jackpot.Status, &jackpot.CreatedAt, &jackpot.UpdatedAt, &jackpot.ExpiresAt,
			&jackpot.NextDrawAt, &jackpot.Description, &jackpot.IsActive,
			&winningNumbers, &winnerID, &winnerAmount,
		)

		if err != nil {
			return nil, err
		}

		jackpot.WinningNumbers = winningNumbers
		jackpot.WinnerID = winnerID
		jackpot.WinnerAmount = winnerAmount
		jackpots = append(jackpots, &jackpot)
	}

	return jackpots, nil
}

// UpdateJackpot updates an existing jackpot
func (r *JackpotRepository) UpdateJackpot(ctx context.Context, jackpot *Jackpot) error {
	query := `
		UPDATE jackpots SET
			name = $2, type = $3, current_amount = $4, seed_amount = $5,
			contribution_rate = $6, min_bet = $7, max_bet = $8, status = $9,
			updated_at = $10, expires_at = $11, next_draw_at = $12,
			description = $13, is_active = $14, winning_numbers = $15,
			winner_id = $16, winner_amount = $17
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		jackpot.ID, jackpot.Name, string(jackpot.Type), jackpot.CurrentAmount,
		jackpot.SeedAmount, jackpot.ContributionRate, jackpot.MinBet, jackpot.MaxBet,
		string(jackpot.Status), jackpot.UpdatedAt, jackpot.ExpiresAt,
		jackpot.NextDrawAt, jackpot.Description, jackpot.IsActive,
		jackpot.WinningNumbers, jackpot.WinnerID, jackpot.WinnerAmount,
	)

	return err
}

// DeleteJackpot deletes a jackpot
func (r *JackpotRepository) DeleteJackpot(ctx context.Context, id string) error {
	query := "DELETE FROM jackpots WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
