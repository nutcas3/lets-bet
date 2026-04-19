package validation_test

import (
	"testing"

	"github.com/betting-platform/internal/infrastructure/validation"
	"github.com/shopspring/decimal"
)

func TestRequired(t *testing.T) {
	t.Parallel()
	if _, ok := validation.Required("name", "alice"); !ok {
		t.Fatal("Required(alice) = false, want true")
	}
	if _, ok := validation.Required("name", "  "); ok {
		t.Fatal("Required(whitespace) = true, want false")
	}
}

func TestEmail(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"alice@example.com":     true,
		"alice+label@example.io": true,
		"not-an-email":          false,
		"":                      false,
		"@example.com":          false,
	}
	for in, want := range cases {
		_, ok := validation.Email("email", in)
		if ok != want {
			t.Errorf("Email(%q) = %v, want %v", in, ok, want)
		}
	}
}

func TestKenyaPhone(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"254712345678": true,
		"0712345678":   true,
		"+254712345678": true,
		"1234":         false,
		"abc":          false,
	}
	for in, want := range cases {
		_, ok := validation.KenyaPhone("phone", in)
		if ok != want {
			t.Errorf("KenyaPhone(%q) = %v, want %v", in, ok, want)
		}
	}
}

func TestPositiveDecimal(t *testing.T) {
	t.Parallel()
	if _, ok := validation.PositiveDecimal("amount", decimal.NewFromInt(10)); !ok {
		t.Fatal("PositiveDecimal(10) = false")
	}
	if _, ok := validation.PositiveDecimal("amount", decimal.Zero); ok {
		t.Fatal("PositiveDecimal(0) = true, want false")
	}
	if _, ok := validation.PositiveDecimal("amount", decimal.NewFromInt(-1)); ok {
		t.Fatal("PositiveDecimal(-1) = true, want false")
	}
}
