package jackpots

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/betting-platform/internal/infrastructure/id"
	"github.com/shopspring/decimal"
)

var jackpotGenerator *id.SnowflakeGenerator

func init() {
	var err error
	jackpotGenerator, err = id.ServiceTypeGenerator("jackpot")
	if err != nil {
		panic(fmt.Sprintf("Failed to create jackpot ID generator: %v", err))
	}
}

// generateID generates a unique ID for jackpot records
func generateID() string {
	return jackpotGenerator.GenerateID()
}

// validateJackpot validates jackpot configuration
func validateJackpot(jackpot *Jackpot) error {
	if jackpot.Name == "" {
		return fmt.Errorf("jackpot name is required")
	}

	if jackpot.SeedAmount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("seed amount must be greater than zero")
	}

	if jackpot.MinBet.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("minimum bet must be greater than zero")
	}

	if jackpot.MaxBet.LessThan(jackpot.MinBet) {
		return fmt.Errorf("maximum bet must be greater than or equal to minimum bet")
	}

	if jackpot.ContributionRate.LessThanOrEqual(decimal.Zero) || jackpot.ContributionRate.GreaterThan(decimal.NewFromInt(1)) {
		return fmt.Errorf("contribution rate must be between 0 and 1")
	}

	return nil
}

// validateTicketPurchase validates ticket purchase
func validateTicketPurchase(jackpot *Jackpot, betAmount decimal.Decimal, numbers []int) error {
	if jackpot.Status != JackpotStatusActive {
		return fmt.Errorf("jackpot is not active")
	}

	if betAmount.LessThan(jackpot.MinBet) {
		return fmt.Errorf("bet amount is below minimum")
	}

	if betAmount.GreaterThan(jackpot.MaxBet) {
		return fmt.Errorf("bet amount exceeds maximum")
	}

	if len(numbers) != 6 {
		return fmt.Errorf("exactly 6 numbers are required")
	}

	for _, num := range numbers {
		if num < 1 || num > 49 {
			return fmt.Errorf("numbers must be between 1 and 49")
		}
	}

	// Check for duplicate numbers
	seen := make(map[int]bool)
	for _, num := range numbers {
		if seen[num] {
			return fmt.Errorf("duplicate numbers are not allowed")
		}
		seen[num] = true
	}

	return nil
}

// generateWinningNumbers generates random winning numbers
func generateWinningNumbers(rng *rand.Rand, count int) []int {
	numbers := make([]int, count)
	seen := make(map[int]bool)

	for i := range count {
		for {
			num := rng.Intn(49) + 1 // 1-49
			if !seen[num] {
				numbers[i] = num
				seen[num] = true
				break
			}
		}
	}

	return numbers
}

// findWinner finds the winning ticket among all tickets
func findWinner(tickets []*JackpotTicket, winningNumbers []int) *JackpotTicket {
	for _, ticket := range tickets {
		if isWinner(ticket.Numbers, winningNumbers) {
			return ticket
		}
	}
	return nil
}

// isWinner checks if a ticket matches the winning numbers
func isWinner(ticketNumbers, winningNumbers []int) bool {
	if len(ticketNumbers) != len(winningNumbers) {
		return false
	}

	ticketSet := make(map[int]bool)
	for _, num := range ticketNumbers {
		ticketSet[num] = true
	}

	for _, num := range winningNumbers {
		if !ticketSet[num] {
			return false
		}
	}

	return true
}

// _calculateOdds calculates the odds of winning
func _calculateOdds(totalNumbers int, numbersToChoose int) *big.Rat {
	// Calculate combinations: C(totalNumbers, numbersToChoose)
	numerator := big.NewInt(1)
	denominator := big.NewInt(1)

	for i := range numbersToChoose {
		numerator.Mul(numerator, big.NewInt(int64(totalNumbers-i)))
		denominator.Mul(denominator, big.NewInt(int64(i+1)))
	}

	return new(big.Rat).SetFrac(numerator, denominator)
}

// _formatOdds formats odds as a string
func _formatOdds(odds *big.Rat) string {
	if odds.IsInt() {
		return fmt.Sprintf("1 in %s", odds.Denom().String())
	}
	return fmt.Sprintf("1 in %.2f", float64(odds.Denom().Int64())/float64(odds.Num().Int64()))
}

// _calculateExpectedValue calculates expected value of a ticket
func _calculateExpectedValue(jackpotAmount, ticketCost decimal.Decimal, totalTickets int64) decimal.Decimal {
	if totalTickets == 0 {
		return decimal.Zero
	}

	odds := decimal.NewFromInt(totalTickets)
	expectedWin := jackpotAmount.Div(odds)
	return expectedWin.Sub(ticketCost)
}

// _getJackpotProgress calculates jackpot progress as percentage
func _getJackpotProgress(currentAmount, targetAmount decimal.Decimal) decimal.Decimal {
	if targetAmount.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return currentAmount.Div(targetAmount).Mul(decimal.NewFromInt(100))
}

// _getTimeUntilDraw returns time until next draw
func _getTimeUntilDraw(nextDrawTime time.Time) string {
	duration := time.Until(nextDrawTime)

	if duration <= 0 {
		return "Now"
	}

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

// _formatAmount formats decimal amount for display
func _formatAmount(amount decimal.Decimal) string {
	return amount.StringFixedBank(2)
}

// _validateNumbers validates number selection
func _validateNumbers(numbers []int, min, max int) error {
	if len(numbers) == 0 {
		return fmt.Errorf("at least one number is required")
	}

	seen := make(map[int]bool)
	for _, num := range numbers {
		if num < min || num > max {
			return fmt.Errorf("number %d is out of range (%d-%d)", num, min, max)
		}
		if seen[num] {
			return fmt.Errorf("duplicate number %d is not allowed", num)
		}
		seen[num] = true
	}

	return nil
}

// _generateQuickPick generates random quick pick numbers
func _generateQuickPick(rng *rand.Rand, count, min, max int) []int {
	if count > (max - min + 1) {
		count = max - min + 1
	}

	numbers := make([]int, 0, count)
	available := make([]int, max-min+1)
	for i := range available {
		available[i] = min + i
	}

	for i := 0; i < count && len(available) > 0; i++ {
		idx := rng.Intn(len(available))
		num := available[idx]
		numbers = append(numbers, num)
		available = append(available[:idx], available[idx+1:]...)
	}

	return numbers
}

// _calculatePrizeTier calculates prize for different matching numbers
func _calculatePrizeTier(jackpotAmount decimal.Decimal, matches int) decimal.Decimal {
	switch matches {
	case 6:
		return jackpotAmount // Full jackpot
	case 5:
		return jackpotAmount.Div(decimal.NewFromInt(100)) // 1% of jackpot
	case 4:
		return jackpotAmount.Div(decimal.NewFromInt(1000)) // 0.1% of jackpot
	case 3:
		return jackpotAmount.Div(decimal.NewFromInt(10000)) // 0.01% of jackpot
	default:
		return decimal.Zero
	}
}
