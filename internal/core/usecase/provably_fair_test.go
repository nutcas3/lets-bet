package usecase_test

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/betting-platform/internal/core/usecase"
	"github.com/shopspring/decimal"
)

func TestGenerateServerSeed_LengthAndUniqueness(t *testing.T) {
	t.Parallel()
	s := usecase.NewProvablyFairService()

	a, err := s.GenerateServerSeed()
	if err != nil {
		t.Fatalf("GenerateServerSeed: %v", err)
	}
	b, err := s.GenerateServerSeed()
	if err != nil {
		t.Fatalf("GenerateServerSeed: %v", err)
	}

	if len(a) != 64 {
		t.Fatalf("seed length = %d, want 64 hex chars", len(a))
	}
	if a == b {
		t.Fatalf("two seeds are identical: %s", a)
	}
}

func TestHashServerSeed_MatchesSHA256(t *testing.T) {
	t.Parallel()
	s := usecase.NewProvablyFairService()

	seed := "deadbeef"
	got := s.HashServerSeed(seed)

	want := sha256.Sum256([]byte(seed))
	if got != hex.EncodeToString(want[:]) {
		t.Fatalf("HashServerSeed = %s, want %s", got, hex.EncodeToString(want[:]))
	}
}

func TestCalculateCrashPoint_IsDeterministic(t *testing.T) {
	t.Parallel()
	s := usecase.NewProvablyFairService()

	a := s.CalculateCrashPoint("server-1", "client-1", 42)
	b := s.CalculateCrashPoint("server-1", "client-1", 42)

	if !a.Equal(b) {
		t.Fatalf("same inputs produced different outputs: %s vs %s", a, b)
	}
}

func TestCalculateCrashPoint_RespectsBounds(t *testing.T) {
	t.Parallel()
	s := usecase.NewProvablyFairService()

	min := decimal.NewFromFloat(1.0)
	max := decimal.NewFromFloat(usecase.MaxCrashMultiplier)

	for i := int64(1); i <= 500; i++ {
		got := s.CalculateCrashPoint("seed", "client", i)
		if got.LessThan(min) {
			t.Fatalf("round %d: crash %s < 1.00", i, got)
		}
		if got.GreaterThan(max) {
			t.Fatalf("round %d: crash %s > MaxCrashMultiplier", i, got)
		}
	}
}

func TestVerifyCrashPoint_RoundTrip(t *testing.T) {
	t.Parallel()
	s := usecase.NewProvablyFairService()

	server, err := s.GenerateServerSeed()
	if err != nil {
		t.Fatalf("GenerateServerSeed: %v", err)
	}
	client := "player-123"
	round := int64(777)

	crash := s.CalculateCrashPoint(server, client, round)

	if !s.VerifyCrashPoint(server, client, round, crash) {
		t.Fatalf("VerifyCrashPoint failed for crash=%s", crash)
	}

	// Any tweak to inputs must make verification fail.
	bumped := crash.Add(decimal.NewFromFloat(0.01))
	if s.VerifyCrashPoint(server, client, round, bumped) {
		t.Fatal("VerifyCrashPoint accepted an altered crash point")
	}
}
