package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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

// CreateTicket creates a new jackpot ticket
func (r *JackpotRepository) CreateTicket(ctx context.Context, ticket *JackpotTicket) error {
	query := `
		INSERT INTO jackpot_tickets (
			id, jackpot_id, user_id, numbers, amount, status,
			created_at, updated_at, drawn_at, won_at, prize_amount
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		ticket.ID, ticket.JackpotID, ticket.UserID, ticket.Numbers,
		ticket.Amount, string(ticket.Status), ticket.CreatedAt, ticket.UpdatedAt,
		ticket.DrawnAt, ticket.WonAt, ticket.PrizeAmount,
	)

	return err
}

// GetTicket retrieves a ticket by ID
func (r *JackpotRepository) GetTicket(ctx context.Context, id string) (*JackpotTicket, error) {
	query := `
		SELECT id, jackpot_id, user_id, numbers, amount, status,
			   created_at, updated_at, drawn_at, won_at, prize_amount
		FROM jackpot_tickets WHERE id = $1
	`

	var ticket JackpotTicket
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ticket.ID, &ticket.JackpotID, &ticket.UserID, &ticket.Numbers,
		&ticket.Amount, &ticket.Status, &ticket.CreatedAt, &ticket.UpdatedAt,
		&ticket.DrawnAt, &ticket.WonAt, &ticket.PrizeAmount,
	)

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

// GetTickets retrieves tickets with optional filters
func (r *JackpotRepository) GetTickets(ctx context.Context, filters *TicketFilters) ([]*JackpotTicket, error) {
	query := `
		SELECT id, jackpot_id, user_id, numbers, amount, status,
			   created_at, updated_at, drawn_at, won_at, prize_amount
		FROM jackpot_tickets
		WHERE 1=1
	`

	args := []any{}
	argIndex := 1

	if filters != nil {
		if filters.JackpotID != nil {
			query += fmt.Sprintf(" AND jackpot_id = $%d", argIndex)
			args = append(args, *filters.JackpotID)
			argIndex++
		}
		if filters.UserID != nil {
			query += fmt.Sprintf(" AND user_id = $%d", argIndex)
			args = append(args, *filters.UserID)
			argIndex++
		}
		if filters.Status != nil {
			query += fmt.Sprintf(" AND status = $%d", argIndex)
			args = append(args, string(*filters.Status))
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

	var tickets []*JackpotTicket
	for rows.Next() {
		var ticket JackpotTicket
		err := rows.Scan(
			&ticket.ID, &ticket.JackpotID, &ticket.UserID, &ticket.Numbers,
			&ticket.Amount, &ticket.Status, &ticket.CreatedAt, &ticket.UpdatedAt,
			&ticket.DrawnAt, &ticket.WonAt, &ticket.PrizeAmount,
		)

		if err != nil {
			return nil, err
		}

		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// UpdateTicket updates an existing ticket
func (r *JackpotRepository) UpdateTicket(ctx context.Context, ticket *JackpotTicket) error {
	query := `
		UPDATE jackpot_tickets SET
			jackpot_id = $2, user_id = $3, numbers = $4, amount = $5,
			status = $6, updated_at = $7, drawn_at = $8, won_at = $9, prize_amount = $10
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		ticket.ID, ticket.JackpotID, ticket.UserID, ticket.Numbers,
		ticket.Amount, string(ticket.Status), ticket.UpdatedAt,
		ticket.DrawnAt, ticket.WonAt, ticket.PrizeAmount,
	)

	return err
}

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
