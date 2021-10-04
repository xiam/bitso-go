package bitso

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	return s.fromString(z)
}

func (s *OrderStatus) String() string {
	if z, ok := statusNames[*s]; ok {
		return z
	}
	panic("unsupported status")
}

func (s OrderStatus) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *OrderStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return s.fromString(value.(string))
}

func (s *OrderStatus) fromString(z string) error {
	for status, name := range statusNames {
		if z == name {
			*s = status
			return nil
		}
	}
	return errors.New("unsupported status")
}
