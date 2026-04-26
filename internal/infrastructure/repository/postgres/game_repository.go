package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// GameRepository implements game repository using PostgreSQL
type GameRepository struct {
	db *sql.DB
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) Create(ctx context.Context, game *domain.Game) error {
	query := `
		INSERT INTO games (
			id, game_type, round_number, server_seed, server_seed_hash, client_seed,
			crash_point, status, started_at, crashed_at, country_code,
			min_bet, max_bet, max_multiplier
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.db.ExecContext(ctx, query,
		game.ID, game.GameType, game.RoundNumber,
		game.ServerSeed, game.ServerSeedHash, game.ClientSeed,
		game.CrashPoint, game.Status, game.StartedAt, game.CrashedAt,
		game.CountryCode, game.MinBet, game.MaxBet, game.MaxMultiplier,
	)

	if err != nil {
		log.Printf("Error creating game: %v", err)
		return err
	}

	return nil
}

func (r *GameRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.GameStatus) error {
	query := `UPDATE games SET status = $1, crashed_at = $2 WHERE id = $3`

	var crashedAt *time.Time
	if status == domain.GameStatusCrashed {
		now := time.Now()
		crashedAt = &now
	}

	_, err := r.db.ExecContext(ctx, query, status, crashedAt, id)
	if err != nil {
		log.Printf("Error updating game status: %v", err)
		return err
	}

	return nil
}

func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	query := `
		SELECT id, game_type, round_number, server_seed, server_seed_hash, client_seed,
			   crash_point, status, started_at, crashed_at, country_code,
			   min_bet, max_bet, max_multiplier
		FROM games WHERE id = $1
	`

	var game domain.Game
	var crashedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&game.ID, &game.GameType, &game.RoundNumber,
		&game.ServerSeed, &game.ServerSeedHash, &game.ClientSeed,
		&game.CrashPoint, &game.Status, &game.StartedAt, &crashedAt,
		&game.CountryCode, &game.MinBet, &game.MaxBet, &game.MaxMultiplier,
	)

	if err != nil {
		return nil, err
	}

	if crashedAt.Valid {
		game.CrashedAt = &crashedAt.Time
	}

	return &game, nil
}

func (r *GameRepository) GetActive(ctx context.Context) ([]*domain.Game, error) {
	query := `
		SELECT id, game_type, round_number, server_seed, server_seed_hash, client_seed,
			   crash_point, status, started_at, crashed_at, country_code,
			   min_bet, max_bet, max_multiplier
		FROM games WHERE status IN ($1, $2) ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, domain.GameStatusWaiting, domain.GameStatusRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []*domain.Game

	for rows.Next() {
		var game domain.Game
		var crashedAt sql.NullTime

		err := rows.Scan(
			&game.ID, &game.GameType, &game.RoundNumber,
			&game.ServerSeed, &game.ServerSeedHash, &game.ClientSeed,
			&game.CrashPoint, &game.Status, &game.StartedAt, &crashedAt,
			&game.CountryCode, &game.MinBet, &game.MaxBet, &game.MaxMultiplier,
		)

		if err != nil {
			return nil, err
		}

		if crashedAt.Valid {
			game.CrashedAt = &crashedAt.Time
		}

		games = append(games, &game)
	}

	return games, nil
}
