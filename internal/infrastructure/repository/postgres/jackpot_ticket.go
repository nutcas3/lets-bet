package postgres

import (
	"context"
	"fmt"
)

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
