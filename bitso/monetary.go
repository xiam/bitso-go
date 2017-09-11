package bitso

import (
	"fmt"
	"strconv"
)

type Monetary string

func (m *Monetary) Float64() float64 {
	v, _ := strconv.ParseFloat(string(*m), 64)
	return v
}

func ToMonetary(in float64) Monetary {
	return Monetary(fmt.Sprintf("%f", in))
}
