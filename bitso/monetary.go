package bitso

import (
	"strconv"
)

type Monetary string

func (m *Monetary) Float64() float64 {
	v, _ := strconv.ParseFloat(string(*m), 64)
	return v
}
