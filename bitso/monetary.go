package bitso

import (
	"fmt"
	"strconv"
)

// Monetary represents a monetary value
type Monetary string

// Float64 returns the monetary value as a float64
func (m *Monetary) Float64() float64 {
	v, _ := strconv.ParseFloat(string(*m), 64)
	return v
}

// ToMonetary converts a float64 value into Monetary
func ToMonetary(in float64) Monetary {
	return Monetary(fmt.Sprintf("%f", in))
}
