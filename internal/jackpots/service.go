package jackpots

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/betting-platform/internal/core/domain"
)

// Repository interfaces to avoid import cycle
type JackpotRepository interface {
	Create(ctx context.Context, jackpot *Jackpot) error
	GetByID(ctx context.Context, id string) (*Jackpot, error)
	Update(ctx context.Context, jackpot *Jackpot) error
	GetActive(ctx context.Context) ([]*Jackpot, error)
	CreateTicket(ctx context.Context, ticket *JackpotTicket) error
	DeleteTicket(ctx context.Context, ticketID string) error
	GetActiveTickets(ctx context.Context, jackpotID string) ([]*JackpotTicket, error)
	GetUserTickets(ctx context.Context, userID string) ([]*JackpotTicket, error)
	UpdateTicketStatus(ctx context.Context, ticketID string, status TicketStatus, prize decimal.Decimal) error
}

type SportBetRepository interface {
	Create(ctx context.Context, bet interface{}) error
}

// JackpotService manages jackpot games and payouts
type JackpotService struct {
	jackpotRepo   JackpotRepository
	betRepo       SportBetRepository
	walletService WalletService
	eventBus      EventBus
	rng           *rand.Rand
	mu            sync.RWMutex
}

// WalletService interface for wallet operations
type WalletService interface {
	Credit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
	Debit(ctx context.Context, userID string, amount decimal.Decimal, movement Movement) (*Transaction, error)
}

// Movement represents a wallet movement
type Movement struct {
	UserID        string                 `json:"user_id"`
	Amount        decimal.Decimal        `json:"amount"`
	Type          domain.TransactionType `json:"type"`
	ReferenceID   *string                `json:"reference_id,omitempty"`
	ReferenceType string                 `json:"reference_type"`
	Description   string                 `json:"description"`
	ProviderName  string                 `json:"provider_name"`
	ProviderTxnID string                 `json:"provider_txn_id"`
	CountryCode   string                 `json:"country_code"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	ID string `json:"id"`
}

// EventBus interface for publishing events
type EventBus interface {
	Publish(topic string, data interface{}) error
}

// NewJackpotService creates a new jackpot service
func NewJackpotService(
	jackpotRepo JackpotRepository,
	betRepo SportBetRepository,
	walletService WalletService,
	eventBus EventBus,
) *JackpotService {
	return &JackpotService{
		jackpotRepo:   jackpotRepo,
		betRepo:       betRepo,
		walletService: walletService,
		eventBus:      eventBus,
		rng:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Jackpot represents a progressive jackpot game
type Jackpot struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Type             JackpotType     `json:"type"`
	CurrentAmount    decimal.Decimal `json:"current_amount"`
	SeedAmount       decimal.Decimal `json:"seed_amount"`
	ContributionRate decimal.Decimal `json:"contribution_rate"` // Percentage of bets that contribute
	MinBet           decimal.Decimal `json:"min_bet"`
	MaxBet           decimal.Decimal `json:"max_bet"`
	Status           JackpotStatus   `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	LastWonAt        *time.Time      `json:"last_won_at,omitempty"`
	LastWonBy        string          `json:"last_won_by,omitempty"`
	NextDrawAt       *time.Time      `json:"next_draw_at,omitempty"`
}

type JackpotType string

const (
	JackpotTypeDaily       JackpotType = "DAILY"
	JackpotTypeWeekly      JackpotType = "WEEKLY"
	JackpotTypeMonthly     JackpotType = "MONTHLY"
	JackpotTypeProgressive JackpotType = "PROGRESSIVE"
	JackpotTypeMystery     JackpotType = "MYSTERY"
)

type JackpotStatus string

const (
	JackpotStatusActive  JackpotStatus = "ACTIVE"
	JackpotStatusPaused  JackpotStatus = "PAUSED"
	JackpotStatusSettled JackpotStatus = "SETTLED"
	JackpotStatusExpired JackpotStatus = "EXPIRED"
)

// JackpotTicket represents a jackpot entry
type JackpotTicket struct {
	ID        string          `json:"id"`
	JackpotID string          `json:"jackpot_id"`
	UserID    string          `json:"user_id"`
	BetAmount decimal.Decimal `json:"bet_amount"`
	Numbers   []int           `json:"numbers"`
	Status    TicketStatus    `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	DrawnAt   *time.Time      `json:"drawn_at,omitempty"`
	Won       bool            `json:"won"`
	Prize     decimal.Decimal `json:"prize"`
}

type TicketStatus string

const (
	TicketStatusActive  TicketStatus = "ACTIVE"
	TicketStatusDrawn   TicketStatus = "DRAWN"
	TicketStatusWon     TicketStatus = "WON"
	TicketStatusExpired TicketStatus = "EXPIRED"
)

// CreateJackpot creates a new jackpot game
func (s *JackpotService) CreateJackpot(ctx context.Context, jackpot *Jackpot) error {
	// Validate jackpot configuration
	if err := s.validateJackpot(jackpot); err != nil {
		return fmt.Errorf("invalid jackpot configuration: %w", err)
	}

	// Set initial values
	jackpot.ID = s.generateID()
	jackpot.Status = JackpotStatusActive
	jackpot.CreatedAt = time.Now()

	// Set next draw time based on type
	switch jackpot.Type {
	case JackpotTypeDaily:
		nextDraw := time.Now().Add(24 * time.Hour)
		jackpot.NextDrawAt = &nextDraw
	case JackpotTypeWeekly:
		nextDraw := time.Now().Add(7 * 24 * time.Hour)
		jackpot.NextDrawAt = &nextDraw
	case JackpotTypeMonthly:
		nextDraw := time.Now().Add(30 * 24 * time.Hour)
		jackpot.NextDrawAt = &nextDraw
	}

	// Save to repository
	err := s.jackpotRepo.Create(ctx, jackpot)
	if err != nil {
		return fmt.Errorf("failed to create jackpot: %w", err)
	}

	// Publish event
	s.publishEvent("jackpot.created", map[string]interface{}{
		"jackpot_id":  jackpot.ID,
		"type":        jackpot.Type,
		"seed_amount": jackpot.SeedAmount,
	})

	return nil
}

// validateJackpot validates jackpot configuration
func (s *JackpotService) validateJackpot(jackpot *Jackpot) error {
	if jackpot.Name == "" {
		return fmt.Errorf("jackpot name is required")
	}

	if jackpot.SeedAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("seed amount must be greater than zero")
	}

	if jackpot.ContributionRate.LessThan(decimal.Zero) || jackpot.ContributionRate.GreaterThan(decimal.NewFromInt(100)) {
		return fmt.Errorf("contribution rate must be between 0 and 100")
	}

	if jackpot.MinBet.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("minimum bet must be greater than zero")
	}

	if jackpot.MaxBet.LessThan(jackpot.MinBet) {
		return fmt.Errorf("maximum bet must be greater than minimum bet")
	}

	return nil
}

// ContributeToJackpot adds contribution from a bet to jackpot pool
func (s *JackpotService) ContributeToJackpot(ctx context.Context, jackpotID string, betAmount decimal.Decimal) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get jackpot
	jackpot, err := s.jackpotRepo.GetByID(ctx, jackpotID)
	if err != nil {
		return fmt.Errorf("failed to get jackpot: %w", err)
	}

	if jackpot.Status != JackpotStatusActive {
		return fmt.Errorf("jackpot is not active")
	}

	// Calculate contribution
	contribution := betAmount.Mul(jackpot.ContributionRate).Div(decimal.NewFromInt(100))

	// Update jackpot amount
	jackpot.CurrentAmount = jackpot.CurrentAmount.Add(contribution)

	// Save updated jackpot
	err = s.jackpotRepo.Update(ctx, jackpot)
	if err != nil {
		return fmt.Errorf("failed to update jackpot: %w", err)
	}

	// Publish contribution event
	s.publishEvent("jackpot.contribution", map[string]interface{}{
		"jackpot_id":   jackpotID,
		"contribution": contribution,
		"new_amount":   jackpot.CurrentAmount,
	})

	return nil
}

// BuyJackpotTicket purchases a ticket for a jackpot game
func (s *JackpotService) BuyJackpotTicket(ctx context.Context, jackpotID string, userID string, betAmount decimal.Decimal, numbers []int) (*JackpotTicket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get jackpot
	jackpot, err := s.jackpotRepo.GetByID(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jackpot: %w", err)
	}

	if jackpot.Status != JackpotStatusActive {
		return nil, fmt.Errorf("jackpot is not active")
	}

	// Validate bet amount
	if betAmount.LessThan(jackpot.MinBet) {
		return nil, fmt.Errorf("bet amount is below minimum")
	}

	if betAmount.GreaterThan(jackpot.MaxBet) {
		return nil, fmt.Errorf("bet amount exceeds maximum")
	}

	// Validate numbers based on jackpot type
	if err := s.validateNumbers(jackpot.Type, numbers); err != nil {
		return nil, fmt.Errorf("invalid numbers: %w", err)
	}

	// Create ticket
	ticket := &JackpotTicket{
		ID:        s.generateID(),
		JackpotID: jackpotID,
		UserID:    userID,
		BetAmount: betAmount,
		Numbers:   numbers,
		Status:    TicketStatusActive,
		CreatedAt: time.Now(),
	}

	// Save ticket
	err = s.jackpotRepo.CreateTicket(ctx, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket: %w", err)
	}

	// Contribute to jackpot pool
	err = s.ContributeToJackpot(ctx, jackpotID, betAmount)
	if err != nil {
		// Rollback ticket creation on contribution failure
		s.jackpotRepo.DeleteTicket(ctx, ticket.ID)
		return nil, fmt.Errorf("failed to contribute to jackpot: %w", err)
	}

	// Publish ticket purchase event
	s.publishEvent("jackpot.ticket.purchased", map[string]interface{}{
		"ticket_id":  ticket.ID,
		"jackpot_id": jackpotID,
		"user_id":    userID,
		"bet_amount": betAmount,
		"numbers":    numbers,
	})

	return ticket, nil
}

// validateNumbers validates ticket numbers based on jackpot type
func (s *JackpotService) validateNumbers(jackpotType JackpotType, numbers []int) error {
	switch jackpotType {
	case JackpotTypeDaily, JackpotTypeWeekly, JackpotTypeMonthly:
		// Traditional lottery - 6 numbers from 1-49
		if len(numbers) != 6 {
			return fmt.Errorf("must select exactly 6 numbers")
		}
		for _, num := range numbers {
			if num < 1 || num > 49 {
				return fmt.Errorf("numbers must be between 1 and 49")
			}
		}
	case JackpotTypeProgressive:
		// Progressive jackpot - 5 numbers from 1-35
		if len(numbers) != 5 {
			return fmt.Errorf("must select exactly 5 numbers")
		}
		for _, num := range numbers {
			if num < 1 || num > 35 {
				return fmt.Errorf("numbers must be between 1 and 35")
			}
		}
	case JackpotTypeMystery:
		// Mystery jackpot - random number of numbers
		if len(numbers) < 3 || len(numbers) > 8 {
			return fmt.Errorf("must select between 3 and 8 numbers")
		}
		for _, num := range numbers {
			if num < 1 || num > 50 {
				return fmt.Errorf("numbers must be between 1 and 50")
			}
		}
	}

	// Check for duplicates
	seen := make(map[int]bool)
	for _, num := range numbers {
		if seen[num] {
			return fmt.Errorf("duplicate numbers not allowed")
		}
		seen[num] = true
	}

	return nil
}

// DrawJackpot performs a jackpot draw
func (s *JackpotService) DrawJackpot(ctx context.Context, jackpotID string) (*JackpotResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get jackpot
	jackpot, err := s.jackpotRepo.GetByID(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get jackpot: %w", err)
	}

	if jackpot.Status != JackpotStatusActive {
		return nil, fmt.Errorf("jackpot is not active")
	}

	// Get all active tickets
	tickets, err := s.jackpotRepo.GetActiveTickets(ctx, jackpotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickets: %w", err)
	}

	if len(tickets) == 0 {
		return nil, fmt.Errorf("no active tickets for jackpot")
	}

	// Generate winning numbers
	winningNumbers := s.generateWinningNumbers(jackpot.Type)

	// Check for winners
	winners := s.determineWinners(tickets, winningNumbers)

	// Create result
	result := &JackpotResult{
		JackpotID:      jackpotID,
		WinningNumbers: winningNumbers,
		TotalTickets:   len(tickets),
		Winners:        winners,
		DrawnAt:        time.Now(),
	}

	// Process payouts
	if len(winners) > 0 {
		err = s.processJackpotPayouts(ctx, jackpot, winners)
		if err != nil {
			log.Printf("Error processing jackpot payouts: %v", err)
		}
	} else {
		// No winners - roll over to next draw
		jackpot.CurrentAmount = jackpot.CurrentAmount.Add(jackpot.SeedAmount)
		s.jackpotRepo.Update(ctx, jackpot)
	}

	// Update jackpot status
	jackpot.Status = JackpotStatusSettled
	jackpot.LastWonAt = &result.DrawnAt
	if len(winners) > 0 {
		jackpot.LastWonBy = winners[0].UserID
	}

	// Set next draw time
	s.setNextDrawTime(jackpot)
	jackpot.Status = JackpotStatusActive // Reactivate for next draw

	err = s.jackpotRepo.Update(ctx, jackpot)
	if err != nil {
		log.Printf("Error updating jackpot after draw: %v", err)
	}

	// Publish draw event
	s.publishEvent("jackpot.drawn", map[string]interface{}{
		"jackpot_id":      jackpotID,
		"winning_numbers": winningNumbers,
		"total_tickets":   len(tickets),
		"winners_count":   len(winners),
		"new_amount":      jackpot.CurrentAmount,
	})

	return result, nil
}

// JackpotResult represents the result of a jackpot draw
type JackpotResult struct {
	JackpotID      string          `json:"jackpot_id"`
	WinningNumbers []int           `json:"winning_numbers"`
	TotalTickets   int             `json:"total_tickets"`
	Winners        []JackpotWinner `json:"winners"`
	DrawnAt        time.Time       `json:"drawn_at"`
}

// JackpotWinner represents a jackpot winner
type JackpotWinner struct {
	UserID   string          `json:"user_id"`
	TicketID string          `json:"ticket_id"`
	Numbers  []int           `json:"numbers"`
	Matches  int             `json:"matches"`
	Prize    decimal.Decimal `json:"prize"`
}

// generateWinningNumbers generates random winning numbers
func (s *JackpotService) generateWinningNumbers(jackpotType JackpotType) []int {
	var maxNum, count int

	switch jackpotType {
	case JackpotTypeDaily, JackpotTypeWeekly, JackpotTypeMonthly:
		maxNum, count = 49, 6
	case JackpotTypeProgressive:
		maxNum, count = 35, 5
	case JackpotTypeMystery:
		maxNum, count = 50, s.rng.Intn(6)+3 // 3-8 numbers
	}

	numbers := make([]int, 0, count)
	seen := make(map[int]bool)

	for len(numbers) < count {
		num := s.rng.Intn(maxNum) + 1
		if !seen[num] {
			numbers = append(numbers, num)
			seen[num] = true
		}
	}

	return numbers
}

// determineWinners finds winners based on matching numbers
func (s *JackpotService) determineWinners(tickets []*JackpotTicket, winningNumbers []int) []JackpotWinner {
	var winners []JackpotWinner

	for _, ticket := range tickets {
		matches := s.countMatches(ticket.Numbers, winningNumbers)
		if matches >= 3 { // Minimum 3 matches to win
			prize := s.calculatePrize(matches, len(winningNumbers))
			winners = append(winners, JackpotWinner{
				UserID:   ticket.UserID,
				TicketID: ticket.ID,
				Numbers:  ticket.Numbers,
				Matches:  matches,
				Prize:    prize,
			})
		}
	}

	return winners
}

// countMatches counts how many numbers match
func (s *JackpotService) countMatches(ticketNumbers, winningNumbers []int) int {
	matches := 0
	winningSet := make(map[int]bool)
	for _, num := range winningNumbers {
		winningSet[num] = true
	}

	for _, num := range ticketNumbers {
		if winningSet[num] {
			matches++
		}
	}

	return matches
}

// calculatePrize calculates prize based on matches
func (s *JackpotService) calculatePrize(matches, totalNumbers int) decimal.Decimal {
	// Prize calculation based on match percentage
	matchPercentage := float64(matches) / float64(totalNumbers)

	// Base prize amounts (can be configured per jackpot)
	basePrizes := map[float64]decimal.Decimal{
		0.5: decimal.NewFromInt(100),    // 50% match
		0.6: decimal.NewFromInt(500),    // 60% match
		0.7: decimal.NewFromInt(2000),   // 70% match
		0.8: decimal.NewFromInt(10000),  // 80% match
		0.9: decimal.NewFromInt(50000),  // 90% match
		1.0: decimal.NewFromInt(100000), // 100% match
	}

	prize, exists := basePrizes[matchPercentage]
	if !exists {
		prize = decimal.NewFromInt(50) // Minimum prize
	}

	return prize
}

// processJackpotPayouts processes payouts to winners
func (s *JackpotService) processJackpotPayouts(ctx context.Context, jackpot *Jackpot, winners []JackpotWinner) error {
	for _, winner := range winners {
		movement := Movement{
			UserID:        winner.UserID,
			Amount:        winner.Prize,
			Type:          domain.TransactionTypeBetWon,
			ReferenceID:   &jackpot.ID,
			ReferenceType: "jackpot",
			Description:   fmt.Sprintf("Jackpot win from %s", jackpot.Name),
			ProviderName:  "jackpot",
			ProviderTxnID: fmt.Sprintf("jackpot-%s-%s", jackpot.ID, winner.TicketID),
			CountryCode:   "KE",
		}

		_, err := s.walletService.Credit(ctx, winner.UserID, winner.Prize, movement)
		if err != nil {
			log.Printf("Error paying jackpot winner %s: %v", winner.UserID, err)
			continue
		}

		// Update ticket status
		err = s.jackpotRepo.UpdateTicketStatus(ctx, winner.TicketID, TicketStatusWon, winner.Prize)
		if err != nil {
			log.Printf("Error updating ticket status for %s: %v", winner.TicketID, err)
		}
	}

	return nil
}

// setNextDrawTime sets the next draw time based on jackpot type
func (s *JackpotService) setNextDrawTime(jackpot *Jackpot) {
	var nextDraw time.Time

	switch jackpot.Type {
	case JackpotTypeDaily:
		nextDraw = time.Now().Add(24 * time.Hour)
	case JackpotTypeWeekly:
		nextDraw = time.Now().Add(7 * 24 * time.Hour)
	case JackpotTypeMonthly:
		nextDraw = time.Now().Add(30 * 24 * time.Hour)
	case JackpotTypeProgressive, JackpotTypeMystery:
		// Progressive and mystery jackpots draw when threshold is reached
		nextDraw = time.Now().Add(1 * time.Hour) // Check every hour
	}

	jackpot.NextDrawAt = &nextDraw
}

// GetActiveJackpots returns all active jackpots
func (s *JackpotService) GetActiveJackpots(ctx context.Context) ([]*Jackpot, error) {
	return s.jackpotRepo.GetActive(ctx)
}

// GetUserTickets returns all tickets for a user
func (s *JackpotService) GetUserTickets(ctx context.Context, userID string) ([]*JackpotTicket, error) {
	return s.jackpotRepo.GetUserTickets(ctx, userID)
}

// generateID generates a unique ID
func (s *JackpotService) generateID() string {
	return fmt.Sprintf("jp_%d_%d", time.Now().Unix(), s.rng.Intn(10000))
}

// publishEvent publishes an event to the event bus
func (s *JackpotService) publishEvent(topic string, data interface{}) {
	if s.eventBus != nil {
		err := s.eventBus.Publish(topic, data)
		if err != nil {
			log.Printf("Error publishing jackpot event %s: %v", topic, err)
		}
	}
}
