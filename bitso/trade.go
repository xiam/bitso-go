package bitso

import "encoding/json"

// Trade represents a recent trade from the specified book
type Trade struct {
	Book      Book      `json:"book"`
	CreatedAt Time      `json:"created_at"`
	Amount    Monetary  `json:"amount"`
	MakerSide OrderSide `json:"maker_side"`
	Price     Monetary  `json:"price"`
	TID       TID       `json:"tid"`
}

// UserTrade represents a trade made by the user
type UserTrade struct {
	Book         Book     `json:"book"`
	Major        Monetary `json:"major"`
	CreatedAt    Time     `json:"created_at"`
	Minor        Monetary `json:"minor"`
	FeesAmount   Monetary `json:"fees_amount"`
	FeesCurrency Currency `json:"fees_currency"`
	Price        Monetary `json:"price"`
	TID          TID      `json:"tid"`
	OID          string   `json:"oid"`
	OriginID     string   `json:"origin_id"`

	Side            OrderSide `json:"side"`
	MakerSide       OrderSide `json:"maker_side"`
	MajorCurrency   Currency  `json:"major_currency"`
	MinorCurrency   Currency  `json:"minor_currency"`
	MarginOrderType string    `json:"margin_order_type"`
}

// UserOrderTrade represents a trade made by the user
type UserOrderTrade struct {
	Book         Book     `json:"book"`
	Major        Monetary `json:"major"`
	CreatedAt    Time     `json:"created_at"`
	Minor        Monetary `json:"minor"`
	FeesAmount   Monetary `json:"fees_amount"`
	FeesCurrency Currency `json:"fees_currency"`
	Price        Monetary `json:"price"`
	TID          TID      `json:"tid"`
	OID          string   `json:"oid"`
	OriginID     string   `json:"origin_id"`

	Side          OrderSide `json:"side"`
	MakerSide     OrderSide `json:"maker_side"`
	MajorCurrency Currency  `json:"major_currency"`
	MinorCurrency Currency  `json:"minor_currency"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (t *UserOrderTrade) UnmarshalJSON(in []byte) error {
	type userOrderTrade UserOrderTrade

	var decoded userOrderTrade
	if err := json.Unmarshal(in, &decoded); err != nil {
		return err
	}
	*t = UserOrderTrade(decoded)

	if t.FeesCurrency == CurrencyNone {
		var legacy struct {
			FeesCurrency Currency `json:"currency"`
		}
		if err := json.Unmarshal(in, &legacy); err != nil {
			return err
		}
		t.FeesCurrency = legacy.FeesCurrency
	}

	return nil
}
