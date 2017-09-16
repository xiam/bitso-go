package bitso

import (
	"encoding/json"
	"errors"
)

// OrderType tells whether the order is a market or limit order.
type OrderType uint8

// List of order types.
const (
	OrderTypeNone OrderType = iota

	OrderTypeMarket
	OrderTypeLimit
)

var orderTypes = map[OrderType]string{
	OrderTypeMarket: "market",
	OrderTypeLimit:  "limit",
}

// MarshalJSON implements json.Marshaler
func (o OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (o *OrderType) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	for k, v := range orderTypes {
		if v == z {
			*o = k
			return nil
		}
	}
	return errors.New("unsupported order type")
}

func (o OrderType) String() string {
	if z, ok := orderTypes[o]; ok {
		return z
	}
	panic("unsupported order type")
}
