package bitso

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

// Monetary represents a monetary value
type Monetary string

// NewMonetary validates a decimal string and returns it as Monetary.
//
// The input string is preserved exactly so callers can keep exchange-required
// scale, trailing zeros, and precision-sensitive fractional values.
func NewMonetary(in string) (Monetary, error) {
	if strings.TrimSpace(in) != in {
		return "", fmt.Errorf("invalid monetary value %q", in)
	}
	if _, err := decimal.NewFromString(in); err != nil {
		return "", fmt.Errorf("invalid monetary value %q: %w", in, err)
	}
	return Monetary(in), nil
}

// NewMonetaryFromDecimal converts a decimal.Decimal value into Monetary.
func NewMonetaryFromDecimal(in decimal.Decimal) Monetary {
	return Monetary(in.String())
}

// Float64 returns the monetary value as a float64
//
// Deprecated: use Decimal for exact decimal math, or Float64E when a float64
// conversion is explicitly needed and parse errors must be handled.
func (m Monetary) Float64() float64 {
	v, _ := strconv.ParseFloat(string(m), 64)
	return v
}

// Float64E returns the monetary value as a float64, reporting parse errors.
//
// Decimal should be preferred for monetary calculations. Float64E is provided
// for integrations that must interoperate with float64 APIs.
func (m Monetary) Float64E() (float64, error) {
	return strconv.ParseFloat(string(m), 64)
}

func (m Monetary) Decimal() (decimal.Decimal, error) {
	return decimal.NewFromString(string(m))
}

// ToMonetary converts a float64 value into Monetary
//
// Deprecated: use NewMonetary with a string literal or NewMonetaryFromDecimal
// to avoid float64 rounding and formatting loss.
func ToMonetary(in float64) Monetary {
	return Monetary(fmt.Sprintf("%f", in))
}
