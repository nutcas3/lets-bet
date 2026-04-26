package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/shopspring/decimal"
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
// UPDATE bets SET status = 'cashed_out', cashout_at = ?, payout = ?
// WHERE id = ? AND status = 'active'
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

// CreateBetWithWalletUpdate implements atomic wallet update + bet creation in transaction
// This method performs both wallet debit and bet creation in a single database transaction
func (r *GameBetRepository) CreateBetWithWalletUpdate(ctx context.Context, bet *domain.GameBet, userID uuid.UUID, amount decimal.Decimal) (uuid.UUID, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return uuid.Nil, err
	}

	// Ensure rollback if function fails
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Update wallet balance atomically
	// This ensures the user has sufficient funds and prevents double-spending
	walletUpdateQuery := `
		UPDATE wallets 
		SET balance = balance - $1, updated_at = $2
		WHERE user_id = $3 AND balance >= $1
		RETURNING balance
	`

	var newBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, walletUpdateQuery, amount, time.Now(), userID).Scan(&newBalance)
	if err != nil {
		log.Printf("Error updating wallet balance: %v", err)
		return uuid.Nil, fmt.Errorf("insufficient balance or wallet error: %w", err)
	}

	// Step 2: Create the bet record
	betCreateQuery := `
		INSERT INTO game_bets (
			id, game_id, user_id, amount, currency, cashed_out, cashout_at,
			payout, status, placed_at, cashed_out_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	var returnedID uuid.UUID
	err = tx.QueryRowContext(ctx, betCreateQuery,
		bet.ID, bet.GameID, bet.UserID, bet.Amount, bet.Currency,
		bet.CashedOut, bet.CashoutAt, bet.Payout, bet.Status,
		bet.PlacedAt, bet.CashedOutAt,
	).Scan(&returnedID)

	if err != nil {
		log.Printf("Error creating game bet: %v", err)
		return uuid.Nil, fmt.Errorf("bet creation failed: %w", err)
	}

	// Step 3: Create transaction record for audit trail
	transactionQuery := `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency, 
			balance_before, balance_after, reference_id, reference_type,
			status, description, created_at, country_code
		) 
		SELECT $1, w.id, $2, $3, $4, $5, 
			   w.balance + $4, w.balance, $6, $7,
			   $8, $9, $10, $11
		FROM wallets w WHERE w.user_id = $2
	`

	transactionID := uuid.New()
	_, err = tx.ExecContext(ctx, transactionQuery,
		transactionID, userID, "BET_PLACED", amount, bet.Currency,
		bet.ID, "BET", "COMPLETED", "Bet placed", time.Now(), "US",
	)
	if err != nil {
		log.Printf("Error creating transaction record: %v", err)
		return uuid.Nil, fmt.Errorf("transaction recording failed: %w", err)
	}

	// Commit transaction if all operations succeeded
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return uuid.Nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// Transaction was successful - no rollback needed
	err = nil

	log.Printf("Successfully created bet %s and updated wallet for user %s, new balance: %s",
		returnedID.String(), userID.String(), newBalance.String())

	return returnedID, nil
}

// AtomicAutoCashoutWithCredit implements atomic auto-cashout with wallet credit in a single transaction
// This prevents the bug where a bet is marked as cashed out but the wallet credit fails
func (r *GameBetRepository) AtomicAutoCashoutWithCredit(ctx context.Context, id uuid.UUID, userID uuid.UUID, cashoutAt decimal.Decimal, payout decimal.Decimal, country string) (bool, error) {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction for auto cashout: %v", err)
		return false, err
	}

	// Ensure rollback if function fails
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Step 1: Atomic bet update - only update if bet is still active
	betUpdateQuery := `
		UPDATE game_bets
		SET cashed_out = true, cashout_at = $1, payout = $2, status = $3, cashed_out_at = $4
		WHERE id = $5 AND status = $6
	`

	now := time.Now()
	result, err := tx.ExecContext(ctx, betUpdateQuery, cashoutAt, payout, domain.GameBetStatusCashedOut, now, id, domain.GameBetStatusActive)
	if err != nil {
		log.Printf("Error updating auto cashout bet: %v", err)
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return false, err
	}

	// If no rows were affected, the bet was already processed (not active anymore)
	if rowsAffected == 0 {
		return false, nil
	}

	// Step 2: Credit the wallet balance
	walletUpdateQuery := `
		UPDATE wallets
		SET balance = balance + $1, bonus_balance = bonus_balance + $2, updated_at = $3
		WHERE user_id = $4
		RETURNING balance
	`

	// For now, credit the full payout to main balance (can be split between main/bonus based on business logic)
	var newBalance decimal.Decimal
	err = tx.QueryRowContext(ctx, walletUpdateQuery, payout, decimal.Zero, now, userID).Scan(&newBalance)
	if err != nil {
		log.Printf("Error crediting wallet for auto cashout bet %s: %v", id, err)
		return false, fmt.Errorf("wallet credit failed: %w", err)
	}

	// Step 3: Create transaction record for audit trail
	transactionQuery := `
		INSERT INTO transactions (
			id, wallet_id, user_id, type, amount, currency,
			balance_before, balance_after, reference_id, reference_type,
			status, description, created_at, country_code
		)
		SELECT $1, w.id, $2, $3, $4, 'KES',
			   w.balance - $4, w.balance, $5, $6,
			   $7, $8, $9, $10
		FROM wallets w WHERE w.user_id = $2
	`

	transactionID := uuid.New()
	_, err = tx.ExecContext(ctx, transactionQuery,
		transactionID, userID, domain.TransactionTypeBetWon, payout,
		id, "BET", "COMPLETED", "Auto cashout payout", now, country,
	)
	if err != nil {
		log.Printf("Error creating transaction record for auto cashout: %v", err)
		return false, fmt.Errorf("transaction recording failed: %w", err)
	}

	// Commit transaction if all operations succeeded
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing auto cashout transaction: %v", err)
		return false, fmt.Errorf("transaction commit failed: %w", err)
	}

	// Transaction was successful - no rollback needed
	err = nil

	log.Printf("Successfully processed auto cashout for bet %s, credited %s to user %s, new balance: %s",
		id.String(), payout.String(), userID.String(), newBalance.String())

	return true, nil
}
