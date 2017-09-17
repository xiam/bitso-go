package bitso

// OrderBook represents a response from /v3/order_book
type OrderBook struct {
	Asks      []Order `json:"asks"`
	Bids      []Order `json:"bids"`
	UpdatedAt Time    `json:"updated_at"`
	Sequence  string  `json:"sequence"`
}

// Order represents a public order.
type Order struct {
	// Order book symbol
	Book Book `json:"book"`
	// Price per unit of major
	Price Monetary `json:"price"`
	// Major amount in order
	Amount Monetary `json:"amount"`
	// Order ID	(only for unaggregated order)
	OID string `json:"oid"`
}

// UserOrder represents an order from the current user.
type UserOrder struct {
	Book Book `json:"book"`

	OriginalAmount Monetary `json:"original_amount"`
	UnfilledAmount Monetary `json:"unfilled_amount"`
	OriginalValue  Monetary `json:"original_value"`

	CreatedAt Time `json:"created_at"`
	UpdatedAt Time `json:"updated_at"`

	Price Monetary `json:"price"`
	OID   string   `json:"oid"`

	Side   OrderSide   `json:"side"`
	Status OrderStatus `json:"status"`
	Type   string      `json:"type"`
}

// OrderPlacement represents an order that can be set by the user.
type OrderPlacement struct {
	Book Book `json:"book"`

	Side OrderSide `json:"side"`
	Type OrderType `json:"type"`

	Major Monetary `json:"major,omitempty"`
	Minor Monetary `json:"minor,omitempty"`
	Price Monetary `json:"price,omitempty"`
}
