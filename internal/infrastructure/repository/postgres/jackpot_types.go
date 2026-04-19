package postgres

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// Jackpot types to avoid import cycle
type JackpotType string
type JackpotStatus string
type TicketStatus string

const (
	JackpotTypeDaily       JackpotType = "DAILY"
	JackpotTypeWeekly      JackpotType = "WEEKLY"
	JackpotTypeMonthly     JackpotType = "MONTHLY"
	JackpotTypeProgressive JackpotType = "PROGRESSIVE"
	JackpotTypeMystery     JackpotType = "MYSTERY"
)

const (
	JackpotStatusActive  JackpotStatus = "ACTIVE"
	JackpotStatusPaused  JackpotStatus = "PAUSED"
	JackpotStatusSettled JackpotStatus = "SETTLED"
	JackpotStatusExpired JackpotStatus = "EXPIRED"
)

const (
	TicketStatusActive  TicketStatus = "ACTIVE"
	TicketStatusDrawn   TicketStatus = "DRAWN"
	TicketStatusWon     TicketStatus = "WON"
	TicketStatusExpired TicketStatus = "EXPIRED"
)

// Jackpot represents a jackpot game
type Jackpot struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Type             JackpotType     `json:"type"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	SeedAmount       decimal.Decimal `json:"seed_amount"`
	ContributionRate decimal.Decimal `json:"contribution_rate"`
	MinBet           decimal.Decimal `json:"min_bet"`
	MaxBet           decimal.Decimal `json:"max_bet"`
	Status           JackpotStatus   `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	ExpiresAt        time.Time       `json:"expires_at"`
	NextDrawAt       time.Time       `json:"next_draw_at"`
	Description      string          `json:"description"`
	IsActive         bool            `json:"is_active"`
	WinningNumbers   []int           `json:"winning_numbers,omitempty"`
	WinnerID         *string         `json:"winner_id,omitempty"`
	WinnerAmount     decimal.Decimal `json:"winner_amount,omitempty"`
}

// JackpotTicket represents a jackpot ticket
type JackpotTicket struct {
	ID          string          `json:"id"`
	JackpotID   string          `json:"jackpot_id"`
	UserID      string          `json:"user_id"`
	Numbers     []int           `json:"numbers"`
	Amount      decimal.Decimal `json:"amount"`
	Status      TicketStatus    `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DrawnAt     *time.Time      `json:"drawn_at,omitempty"`
	WonAt       *time.Time      `json:"won_at,omitempty"`
	PrizeAmount decimal.Decimal `json:"prize_amount,omitempty"`
}

// JackpotContribution represents a contribution to the jackpot
type JackpotContribution struct {
	ID        string          `json:"id"`
	JackpotID string          `json:"jackpot_id"`
	UserID    string          `json:"user_id"`
	Amount    decimal.Decimal `json:"amount"`
	CreatedAt time.Time       `json:"created_at"`
}

// JackpotWinner represents a jackpot winner
type JackpotWinner struct {
	ID             string          `json:"id"`
	JackpotID      string          `json:"jackpot_id"`
	TicketID       string          `json:"ticket_id"`
	UserID         string          `json:"user_id"`
	WinningAmount  decimal.Decimal `json:"winning_amount"`
	WinningNumbers []int           `json:"winning_numbers"`
	WonAt          time.Time       `json:"won_at"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	Status         string          `json:"status"`
}

// JackpotDraw represents a jackpot draw event
type JackpotDraw struct {
	ID            string          `json:"id"`
	JackpotID     string          `json:"jackpot_id"`
	DrawNumbers   []int           `json:"draw_numbers"`
	DrawAt        time.Time       `json:"draw_at"`
	TotalTickets  int             `json:"total_tickets"`
	WinningTicket *string         `json:"winning_ticket,omitempty"`
	WinnerID      *string         `json:"winner_id,omitempty"`
	PrizeAmount   decimal.Decimal `json:"prize_amount"`
	Status        string          `json:"status"`
}

// JackpotMetrics represents jackpot statistics
type JackpotMetrics struct {
	TotalJackpots      int64           `json:"total_jackpots"`
	ActiveJackpots     int64           `json:"active_jackpots"`
	TotalTickets       int64           `json:"total_tickets"`
	ActiveTickets      int64           `json:"active_tickets"`
	TotalContributions decimal.Decimal `json:"total_contributions"`
	TotalPayouts       decimal.Decimal `json:"total_payouts"`
	AverageTicketValue decimal.Decimal `json:"average_ticket_value"`
	LastDrawTime       time.Time       `json:"last_draw_time"`
	NextDrawTime       time.Time       `json:"next_draw_time"`
}

// JackpotRepositoryInterface interface for jackpot operations
type JackpotRepositoryInterface interface {
	// Jackpot operations
	CreateJackpot(ctx context.Context, jackpot *Jackpot) error
	GetJackpot(ctx context.Context, id string) (*Jackpot, error)
	GetJackpots(ctx context.Context, filters *JackpotFilters) ([]*Jackpot, error)
	UpdateJackpot(ctx context.Context, jackpot *Jackpot) error
	DeleteJackpot(ctx context.Context, id string) error

	// Ticket operations
	CreateTicket(ctx context.Context, ticket *JackpotTicket) error
	GetTicket(ctx context.Context, id string) (*JackpotTicket, error)
	GetTickets(ctx context.Context, filters *TicketFilters) ([]*JackpotTicket, error)
	UpdateTicket(ctx context.Context, ticket *JackpotTicket) error

	// Contribution operations
	AddContribution(ctx context.Context, contribution *JackpotContribution) error
	GetContributions(ctx context.Context, jackpotID string) ([]*JackpotContribution, error)

	// Draw operations
	CreateDraw(ctx context.Context, draw *JackpotDraw) error
	GetDraws(ctx context.Context, jackpotID string) ([]*JackpotDraw, error)
	GetLatestDraw(ctx context.Context, jackpotID string) (*JackpotDraw, error)

	// Winner operations
	CreateWinner(ctx context.Context, winner *JackpotWinner) error
	GetWinners(ctx context.Context, filters *WinnerFilters) ([]*JackpotWinner, error)
	UpdateWinner(ctx context.Context, winner *JackpotWinner) error

	// Metrics
	GetMetrics(ctx context.Context) (*JackpotMetrics, error)
}

// JackpotFilters represents filters for jackpot queries
type JackpotFilters struct {
	Type     *JackpotType   `json:"type,omitempty"`
	Status   *JackpotStatus `json:"status,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
	From     *time.Time     `json:"from,omitempty"`
	To       *time.Time     `json:"to,omitempty"`
	Limit    int            `json:"limit,omitempty"`
	Offset   int            `json:"offset,omitempty"`
}

// TicketFilters represents filters for ticket queries
type TicketFilters struct {
	JackpotID *string       `json:"jackpot_id,omitempty"`
	UserID    *string       `json:"user_id,omitempty"`
	Status    *TicketStatus `json:"status,omitempty"`
	From      *time.Time    `json:"from,omitempty"`
	To        *time.Time    `json:"to,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
}

// WinnerFilters represents filters for winner queries
type WinnerFilters struct {
	JackpotID *string    `json:"jackpot_id,omitempty"`
	UserID    *string    `json:"user_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}
