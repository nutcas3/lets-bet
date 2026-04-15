package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"math/big"

	"github.com/shopspring/decimal"
)

// ProvablyFairService handles cryptographic fairness for crash games
type ProvablyFairService struct{}

func NewProvablyFairService() *ProvablyFairService {
	return &ProvablyFairService{}
}

// GenerateServerSeed creates a random server seed
func (s *ProvablyFairService) GenerateServerSeed() string {
	// In production, use crypto/rand
	// This is simplified for demonstration
	return "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}

// HashServerSeed creates SHA-256 hash of server seed (public)
func (s *ProvablyFairService) HashServerSeed(serverSeed string) string {
	hash := sha256.Sum256([]byte(serverSeed))
	return hex.EncodeToString(hash[:])
}

// CalculateCrashPoint generates the crash multiplier using Provably Fair algorithm
// This is the CORE of trust in crash games
func (s *ProvablyFairService) CalculateCrashPoint(serverSeed, clientSeed string, roundNumber int64) decimal.Decimal {
	// 1. Create combined seed
	combined := serverSeed + "-" + clientSeed + "-" + string(rune(roundNumber))
	
	// 2. Generate HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(serverSeed))
	h.Write([]byte(combined))
	hash := h.Sum(nil)
	
	// 3. Convert first 8 bytes to integer
	hexValue := hex.EncodeToString(hash[:8])
	
	// 4. Convert hex to decimal
	val := new(big.Int)
	val.SetString(hexValue, 16)
	
	// 5. Calculate crash point using house edge formula
	// House edge: 1% (configurable)
	houseEdge := 0.01
	maxValue := math.Pow(2, 52) // 2^52 for precision
	
	// Convert to float for calculation
	floatVal := float64(val.Uint64())
	result := (maxValue - floatVal) / (maxValue * houseEdge)
	
	// 6. Apply floor function for fairness
	crashPoint := math.Floor(result * 100) / 100
	
	// 7. Ensure minimum crash point (e.g., 1.00x)
	if crashPoint < 1.00 {
		crashPoint = 1.00
	}
	
	// 8. Cap maximum (e.g., 10,000x for safety)
	if crashPoint > 10000.00 {
		crashPoint = 10000.00
	}
	
	return decimal.NewFromFloat(crashPoint)
}

// VerifyCrashPoint allows players to verify the result after the game
func (s *ProvablyFairService) VerifyCrashPoint(serverSeed, clientSeed string, roundNumber int64, claimedCrash decimal.Decimal) bool {
	calculated := s.CalculateCrashPoint(serverSeed, clientSeed, roundNumber)
	return calculated.Equal(claimedCrash)
}

// Alternative: Simple deterministic algorithm (used by many platforms)
func (s *ProvablyFairService) SimpleCrashPoint(serverSeed string, roundNumber int64) decimal.Decimal {
	combined := serverSeed + string(rune(roundNumber))
	hash := sha256.Sum256([]byte(combined))
	
	// Convert to 0-1 range
	hexStr := hex.EncodeToString(hash[:4])
	val := new(big.Int)
	val.SetString(hexStr, 16)
	
	// Max value for 4 bytes
	maxVal := new(big.Int).Exp(big.NewInt(256), big.NewInt(4), nil)
	
	// Calculate probability (0 to 1)
	prob := new(big.Float).Quo(
		new(big.Float).SetInt(val),
		new(big.Float).SetInt(maxVal),
	)
	
	probFloat, _ := prob.Float64()
	
	// Calculate crash point with house edge
	// Formula: 99 / (100 * probability)
	// This creates a distribution where lower crash points are more common
	if probFloat < 0.01 {
		probFloat = 0.01 // Prevent division by zero
	}
	
	crashPoint := 0.99 / probFloat
	
	// Floor to 2 decimals
	crashPoint = math.Floor(crashPoint * 100) / 100
	
	// Constraints
	if crashPoint < 1.00 {
		crashPoint = 1.00
	}
	if crashPoint > 100.00 {
		crashPoint = 100.00
	}
	
	return decimal.NewFromFloat(crashPoint)
}
