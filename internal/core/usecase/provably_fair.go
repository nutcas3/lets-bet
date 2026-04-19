package usecase

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

// HouseEdgePercent is the configured house edge (3% by default). The published
// commitment / server seed lets players independently verify this value.
const HouseEdgePercent = 3

// MaxCrashMultiplier caps payouts to protect liquidity.
const MaxCrashMultiplier = 10000.0

// ProvablyFairService produces provably-fair crash outcomes using the standard
// Bustabit-style algorithm:
//
//  1. A 32-byte server seed is generated with crypto/rand before the round.
//  2. sha256(serverSeed) is published as a commitment before bets open.
//  3. At round resolution, HMAC_SHA256(serverSeed, "<round>:<clientSeed>")
//     is taken. The first 52 bits form an integer r in [0, 2^52). With
//     probability HouseEdgePercent/100 the round is instantly busted at 1.00.
//     Otherwise the crash point is (2^52 / r) rounded to two decimals.
//
// The server seed is revealed to players after the round; they hash it and
// compare against the earlier commitment. Anyone can then recompute the crash.
type ProvablyFairService struct{}

// NewProvablyFairService constructs a new provably-fair service.
func NewProvablyFairService() *ProvablyFairService {
	return &ProvablyFairService{}
}

// GenerateServerSeed returns a cryptographically random 32-byte seed, hex-encoded.
func (s *ProvablyFairService) GenerateServerSeed() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate server seed: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// HashServerSeed returns the public commitment published before the round.
func (s *ProvablyFairService) HashServerSeed(serverSeed string) string {
	h := sha256.Sum256([]byte(serverSeed))
	return hex.EncodeToString(h[:])
}

// CalculateCrashPoint derives the crash multiplier for a round.
func (s *ProvablyFairService) CalculateCrashPoint(serverSeed, clientSeed string, roundNumber int64) decimal.Decimal {
	msg := strconv.FormatInt(roundNumber, 10) + ":" + clientSeed
	mac := hmac.New(sha256.New, []byte(serverSeed))
	mac.Write([]byte(msg))
	sum := mac.Sum(nil)

	// Instant-bust slot: (HouseEdge % of rounds) crash at 1.00x exactly.
	if int(binary.BigEndian.Uint32(sum[28:32]))%100 < HouseEdgePercent {
		return decimal.NewFromFloat(1.00)
	}

	// Use the first 52 bits (6.5 bytes) as r ∈ [0, 2^52). 52 bits is the max
	// integer precision of float64 without loss.
	r := binary.BigEndian.Uint64(sum[0:8]) >> 12
	const mod = 1 << 52

	// Standard Bustabit formula. r=0 would blow up, but that happens with
	// probability 2^-52 and the instant-bust branch usually catches it; guard
	// anyway for safety.
	if r == 0 {
		return decimal.NewFromFloat(MaxCrashMultiplier)
	}
	crash := math.Floor(float64(mod)/float64(r)*100) / 100

	if crash < 1.00 {
		crash = 1.00
	}
	if crash > MaxCrashMultiplier {
		crash = MaxCrashMultiplier
	}
	return decimal.NewFromFloat(crash)
}

// VerifyCrashPoint returns true when the claimed crash matches what the seeds
// produce. Used by the /verify endpoint after a round is settled.
func (s *ProvablyFairService) VerifyCrashPoint(serverSeed, clientSeed string, roundNumber int64, claimedCrash decimal.Decimal) bool {
	return s.CalculateCrashPoint(serverSeed, clientSeed, roundNumber).Equal(claimedCrash)
}
