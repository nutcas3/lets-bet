package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/shopspring/decimal"
)

// Errors accumulates validation errors.
type Errors map[string][]string

func (e Errors) Add(field, msg string) {
	e[field] = append(e[field], msg)
}

func (e Errors) HasAny() bool {
	return len(e) > 0
}

func (e Errors) Error() string {
	var sb strings.Builder
	for f, msgs := range e {
		sb.WriteString(fmt.Sprintf("%s: %s; ", f, strings.Join(msgs, ", ")))
	}
	return strings.TrimSuffix(sb.String(), "; ")
}

// DecodeJSON reads and decodes a JSON body into v with a 1 MiB cap.
// The response writer is passed so http.MaxBytesReader can surface 413 responses
// to the client when the limit is exceeded.
func DecodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("empty request body")
		}
		return fmt.Errorf("invalid json: %w", err)
	}
	return nil
}

// --- common validators ---

var (
	kenyaPhoneRegex = regexp.MustCompile(`^(?:\+?254|0)(7\d{8}|1\d{8})$`)
	emailRegex      = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// Required returns an error if s is empty.
func Required(field, s string) (string, bool) {
	if strings.TrimSpace(s) == "" {
		return field + " is required", false
	}
	return "", true
}

// MinLen validates minimum length.
func MinLen(field, s string, n int) (string, bool) {
	if len(s) < n {
		return fmt.Sprintf("%s must be at least %d characters", field, n), false
	}
	return "", true
}

// MaxLen validates maximum length.
func MaxLen(field, s string, n int) (string, bool) {
	if len(s) > n {
		return fmt.Sprintf("%s must be at most %d characters", field, n), false
	}
	return "", true
}

// Email validates a valid email address.
func Email(field, s string) (string, bool) {
	if !emailRegex.MatchString(s) {
		return field + " is not a valid email", false
	}
	return "", true
}

// KenyaPhone validates a Kenyan phone number (+254..., 254..., 07..., 01...).
func KenyaPhone(field, s string) (string, bool) {
	if !kenyaPhoneRegex.MatchString(s) {
		return field + " is not a valid Kenyan phone number", false
	}
	return "", true
}

// PositiveDecimal ensures the amount is greater than zero.
func PositiveDecimal(field string, d decimal.Decimal) (string, bool) {
	if d.LessThanOrEqual(decimal.Zero) {
		return field + " must be greater than zero", false
	}
	return "", true
}

// DecimalInRange validates min <= d <= max.
func DecimalInRange(field string, d, min, max decimal.Decimal) (string, bool) {
	if d.LessThan(min) || d.GreaterThan(max) {
		return fmt.Sprintf("%s must be between %s and %s", field, min.String(), max.String()), false
	}
	return "", true
}

// In validates that s is one of allowed values.
func In(field, s string, allowed []string) (string, bool) {
	if slices.Contains(allowed, s) {
		return "", true
	}
	return fmt.Sprintf("%s must be one of %s", field, strings.Join(allowed, ", ")), false
}
