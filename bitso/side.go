package bitso

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// OrderSide tells whether an order is a buy or a sell.
type OrderSide uint8

// List of order sides.
const (
	OrderSideNone OrderSide = iota

	OrderSideBuy
	OrderSideSell
)

var orderSides = map[OrderSide]string{
	OrderSideBuy:  "buy",
	OrderSideSell: "sell",
}

// MarshalJSON implements json.Marshaler
func (s OrderSide) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (s *OrderSide) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	return s.fromString(z)
}

func (s *OrderSide) String() string {
	if z, ok := orderSides[*s]; ok {
		return z
	}
	panic("unsupported order side")
}

func (s OrderSide) Value() (driver.Value, error) {
	return s.String(), nil
}

func (s *OrderSide) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return s.fromString(value.(string))
}

func (s *OrderSide) fromString(z string) error {
	for side, name := range orderSides {
		if z == name {
			*s = side
			return nil
		}
	}
	return errors.New("unsupported order side")
}
