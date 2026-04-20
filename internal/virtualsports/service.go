package virtualsports

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/betting-platform/internal/core/domain"
	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/betting-platform/internal/infrastructure/repository/postgres"
	"github.com/shopspring/decimal"
)

// NewVirtualSportsService creates a new virtual sports service
func NewVirtualSportsService(
	matchRepo postgres.MatchRepository,
	marketRepo postgres.BettingMarketRepository,
	outcomeRepo postgres.MarketOutcomeRepository,
	betRepo postgres.SportBetRepository,
	walletService WalletService,
	eventBus EventBus,
) *VirtualSportsService {
	gameIDGenerator, err := id.ServiceTypeGenerator("virtualsports")
	if err != nil {
		panic(fmt.Sprintf("Failed to create virtual sports ID generator: %v", err))
	}

	return &VirtualSportsService{
		matchRepo:       matchRepo,
		marketRepo:      marketRepo,
		outcomeRepo:     outcomeRepo,
		betRepo:         betRepo,
		walletService:   walletService,
		eventBus:        eventBus,
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
		gameIDGenerator: gameIDGenerator,
		games:           make(map[string]*VirtualGame),
		schedules:       make(map[string]*GameSchedule),
	}
}

// CreateVirtualGame creates a new virtual sports game
func (s *VirtualSportsService) CreateVirtualGame(ctx context.Context, gameType domain.Sport) (*VirtualGame, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate teams
	homeTeam := s.generateTeam(gameType)
	awayTeam := s.generateTeam(gameType)

	// Create game
	game := &VirtualGame{
		ID:        s.generateGameID(),
		Sport:     string(gameType),
		Name:      fmt.Sprintf("%s vs %s", homeTeam.Name, awayTeam.Name),
		Status:    "upcoming",
		StartTime: time.Now().Add(5 * time.Minute),
		HomeTeam:  homeTeam,
		AwayTeam:  awayTeam,
		Odds:      s.generateOdds(homeTeam, awayTeam),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.games[game.ID] = game
	return game, nil
}

// StartGame starts a virtual sports game
func (s *VirtualSportsService) StartGame(ctx context.Context, gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[gameID]
	if !exists {
		return fmt.Errorf("game not found")
	}

	if game.Status != "upcoming" {
		return fmt.Errorf("game cannot be started")
	}

	game.Status = "live"
	game.StartTime = time.Now()
	game.Score = &VirtualScore{HomeScore: 0, AwayScore: 0}

	// Start game simulation in background
	go s.simulateGame(ctx, game)

	return nil
}

// GetGame retrieves a virtual sports game
func (s *VirtualSportsService) GetGame(ctx context.Context, gameID string) (*VirtualGame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	return game, nil
}

// GetActiveGames retrieves all active virtual sports games
func (s *VirtualSportsService) GetActiveGames(ctx context.Context) ([]*VirtualGame, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var activeGames []*VirtualGame
	for _, game := range s.games {
		if game.Status == "live" || game.Status == "upcoming" {
			activeGames = append(activeGames, game)
		}
	}

	return activeGames, nil
}

// PlaceBet places a bet on a virtual sports game
func (s *VirtualSportsService) PlaceBet(ctx context.Context, req *BetRequest) (*BetResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[req.GameID]
	if !exists {
		return &BetResponse{
			Success: false,
			Message: "game not found",
		}, nil
	}

	if game.Status != "upcoming" {
		return &BetResponse{
			Success: false,
			Message: "betting not allowed for this game",
		}, nil
	}

	// Validate odds
	validOdds := false
	for _, odds := range game.Odds {
		for _, outcome := range odds.Outcomes {
			if outcome.OutcomeID == req.OutcomeID && outcome.IsActive {
				if !req.Odds.Equals(outcome.Odds) {
					return &BetResponse{
						Success: false,
						Message: "odds have changed",
					}, nil
				}
				validOdds = true
				break
			}
		}
		if validOdds {
			break
		}
	}

	if !validOdds {
		return &BetResponse{
			Success: false,
			Message: "invalid outcome or odds",
		}, nil
	}

	// Calculate potential win
	potentialWin := req.Amount.Mul(req.Odds)

	// Debit wallet
	movement := Movement{
		UserID:        req.UserID,
		Amount:        req.Amount,
		Type:          domain.TransactionTypeWithdrawal,
		ReferenceID:   &req.GameID,
		ReferenceType: "virtual_sports_bet",
		Description:   fmt.Sprintf("Bet on %s", game.Name),
		ProviderName:  "virtual_sports",
		CountryCode:   "KE",
	}

	_, err := s.walletService.Debit(ctx, req.UserID, req.Amount, movement)
	if err != nil {
		return &BetResponse{
			Success: false,
			Message: "failed to process payment",
		}, err
	}

	// Create bet record
	betID := s.generateBetID()
	betResponse := &BetResponse{
		BetID:        betID,
		Success:      true,
		Message:      "bet placed successfully",
		Amount:       req.Amount,
		Odds:         req.Odds,
		PotentialWin: potentialWin,
	}

	// Publish event
	if err := s.eventBus.Publish("virtual_sports.bet_placed", betResponse); err != nil {
		log.Printf("failed to publish bet placed event: %v", err)
	}

	return betResponse, nil
}

// GetMetrics returns service metrics
func (s *VirtualSportsService) GetMetrics(ctx context.Context) (*VirtualSportsMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalGames := int64(len(s.games))
	activeGames := int64(0)
	completedGames := int64(0)

	for _, game := range s.games {
		switch game.Status {
		case "live":
			activeGames++
		case "completed":
			completedGames++
		}
	}

	return &VirtualSportsMetrics{
		TotalGames:     totalGames,
		ActiveGames:    activeGames,
		CompletedGames: completedGames,
		LastActivity:   time.Now(),
	}, nil
}

// generateTeam generates a virtual team
func (s *VirtualSportsService) generateTeam(sport domain.Sport) *VirtualTeam {
	// Implementation stub
	return &VirtualTeam{
		ID:        fmt.Sprintf("team_%d", rand.Int63()),
		Name:      "Team",
		ShortName: "TM",
	}
}

// generateGameID generates a unique time-based deterministic game ID
func (s *VirtualSportsService) generateGameID() string {
	return fmt.Sprintf("game_%s", s.gameIDGenerator.GenerateID())
}

// generateOdds generates betting odds
func (s *VirtualSportsService) generateOdds(homeTeam, awayTeam *VirtualTeam) []VirtualOdds {
	// Implementation stub
	return []VirtualOdds{
		{
			MarketID:   "moneyline",
			MarketName: "Moneyline",
			Outcomes: []VirtualOutcome{
				{
					OutcomeID: "home",
					Name:      homeTeam.Name,
					Odds:      decimal.NewFromFloat(2.1),
				},
				{
					OutcomeID: "away",
					Name:      awayTeam.Name,
					Odds:      decimal.NewFromFloat(1.8),
				},
			},
		},
	}
}

// simulateGame simulates a virtual game
func (s *VirtualSportsService) simulateGame(ctx context.Context, game *VirtualGame) error {
	// Implementation stub
	return nil
}

// generateBetID generates a unique time-based deterministic bet ID
func (s *VirtualSportsService) generateBetID() string {
	return fmt.Sprintf("bet_%s", s.gameIDGenerator.GenerateID())
}
