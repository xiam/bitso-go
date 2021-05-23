package bitso

import (
	"encoding/json"
	"fmt"
)

// OrderStatus represents the status of open orders.
type OrderStatus uint8

// List of order statuses.
const (
	OrderStatusNone OrderStatus = iota

	OrderStatusOpen
	OrderStatusQueued
	OrderStatusPartialFill
)

var statusNames = map[OrderStatus]string{
	OrderStatusOpen:        "open",
	OrderStatusPartialFill: "partially filled",
	OrderStatusQueued:      "queued",
}

// MarshalJSON implements json.Marshaler
func (s OrderStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (s *OrderStatus) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	for k, v := range statusNames {
		if z == v {
			*s = k
			return nil
		}
	}
	return fmt.Errorf("unsupported status: %q", z)
}

func (s *OrderStatus) String() string {
	if z, ok := statusNames[*s]; ok {
		return z
	}
	panic("unsupported status")
}
