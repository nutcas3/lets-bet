package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// GameBetRepository implements game bet repository using PostgreSQL
type GameBetRepository struct {
	db *sql.DB
}

func NewGameBetRepository(db *sql.DB) *GameBetRepository {
	return &GameBetRepository{db: db}
}

func (r *GameBetRepository) Create(ctx context.Context, bet *domain.GameBet) error {
	query := `
		INSERT INTO game_bets (
			id, game_id, user_id, amount, currency, cashed_out, cashout_at,
			payout, status, placed_at, cashed_out_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		bet.ID, bet.GameID, bet.UserID, bet.Amount, bet.Currency,
		bet.CashedOut, bet.CashoutAt, bet.Payout, bet.Status,
		bet.PlacedAt, bet.CashedOutAt,
	)

	if err != nil {
		log.Printf("Error creating game bet: %v", err)
		return err
	}

	return nil
}

func (r *GameBetRepository) GetActiveByGame(ctx context.Context, gameID uuid.UUID) ([]*domain.GameBet, error) {
	query := `
		SELECT id, game_id, user_id, amount, currency, cashed_out, cashout_at,
			   payout, status, placed_at, cashed_out_at
		FROM game_bets WHERE game_id = $1 AND status = $2
	`

	rows, err := r.db.QueryContext(ctx, query, gameID, domain.GameBetStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []*domain.GameBet

	for rows.Next() {
		var bet domain.GameBet
		var cashoutAt database.NullDecimal
		var cashedOutAt sql.NullTime

		err := rows.Scan(
			&bet.ID, &bet.GameID, &bet.UserID, &bet.Amount, &bet.Currency,
			&bet.CashedOut, &cashoutAt, &bet.Payout, &bet.Status,
			&bet.PlacedAt, &cashedOutAt,
		)

		if err != nil {
			return nil, err
		}

		if cashoutAt.Valid {
			bet.CashoutAt = &cashoutAt.Decimal
		}

		if cashedOutAt.Valid {
			bet.CashedOutAt = &cashedOutAt.Time
		}

		bets = append(bets, &bet)
	}

	return bets, nil
}

func (r *GameBetRepository) UpdateCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) error {
	query := `
		UPDATE game_bets 
		SET cashed_out = true, cashout_at = $1, payout = $2, status = $3, cashed_out_at = $4
		WHERE id = $5
	`

	now := time.Now()

	_, err := r.db.ExecContext(ctx, query, cashoutAt, payout, domain.GameBetStatusCashedOut, now, id)
	if err != nil {
		log.Printf("Error updating cashout: %v", err)
		return err
	}

	return nil
}

func (r *GameBetRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.GameBet, error) {
	query := `
		SELECT id, game_id, user_id, amount, currency, cashed_out, cashout_at,
			   payout, status, placed_at, cashed_out_at
		FROM game_bets WHERE id = $1
	`

	var bet domain.GameBet
	var cashoutAt database.NullDecimal
	var cashedOutAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&bet.ID, &bet.GameID, &bet.UserID, &bet.Amount, &bet.Currency,
		&bet.CashedOut, &cashoutAt, &bet.Payout, &bet.Status,
		&bet.PlacedAt, &cashedOutAt,
	)

	if err != nil {
		return nil, err
	}

	if cashoutAt.Valid {
		bet.CashoutAt = &cashoutAt.Decimal
	}

	if cashedOutAt.Valid {
		bet.CashedOutAt = &cashedOutAt.Time
	}

	return &bet, nil
}

func (r *GameBetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.GameBetStatus) error {
	query := `UPDATE game_bets SET status = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		log.Printf("Error updating bet status: %v", err)
		return err
	}

	return nil
}

// AtomicCashout implements atomic SQL update for double-cashout prevention
func (r *GameBetRepository) AtomicCashout(ctx context.Context, id uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal) (bool, error) {
	query := `
		UPDATE game_bets 
		SET cashed_out = true, cashout_at = $1, payout = $2, status = $3, cashed_out_at = $4
		WHERE id = $5 AND status = $6
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, cashoutAt, payout, domain.GameBetStatusCashedOut, now, id, domain.GameBetStatusActive)
	if err != nil {
		log.Printf("Error performing atomic cashout: %v", err)
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false, err
	}

	return rowsAffected > 0, nil
}
