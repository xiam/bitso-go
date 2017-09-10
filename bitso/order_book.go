package bitso

import (
	"encoding/json"
	"errors"
)

type OrderType uint8

const (
	OrderTypeNone OrderType = iota

	OrderTypeMarket
	OrderTypeLimit
)

var orderTypes = map[OrderType]string{
	OrderTypeMarket: "market",
	OrderTypeLimit:  "limit",
}

func (o OrderType) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

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

type Order struct {
	// Order book symbol
	Book Book `json:"book"`
	// Price per unit of major
	Price Monetary `json:"price"`
	// Major amount in order
	Amount Monetary `json:"amount"`
	// Order ID	(only for unaggregated order)
	Oid string `json:"oid"`
}

type Ask struct {
	Order
}

type Bid struct {
	Order
}

type UserOrder struct {
	Book           Book      `json:"book"`
	OriginalAmount Monetary  `json:"original_amount"`
	UnfilledAmount Monetary  `json:"unfilled_amount"`
	OriginalValue  Monetary  `json:"original_value"`
	CreatedAt      Time      `json:"created_at"`
	UpdatedAt      Time      `json:"updated_at"`
	Price          Monetary  `json:"price"`
	OID            string    `json:"oid"`
	Side           OrderSide `json:"side"`
	Status         Status    `json:"status"`
	Type           string    `json:"type"`
}

type OrderPlacement struct {
	Book  Book      `json:"book"`
	Side  OrderSide `json:"side"`
	Type  OrderType `json:"type"`
	Major Monetary  `json:"major"`
	Minor Monetary  `json:"minor"`
	Price Monetary  `json:"price"`
}
