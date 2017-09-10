package bitso

// Envelope represents a common response envelope from Bitso API.
type Envelope struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// AvailableBooksResponse represents the response from /v3/available_books.
type AvailableBooksResponse struct {
	Envelope
	Payload []ExchangeOrderBook `json:"payload"`
}

// TickerResponse
type TickerResponse struct {
	Envelope
	Payload Ticker `json:"payload"`
}

// OrderBookResponse
type OrderBookResponse struct {
	Envelope
	Payload struct {
		Asks      []Ask  `json:"asks"`
		Bids      []Bid  `json:"bids"`
		UpdatedAt Time   `json:"updated_at"`
		Sequence  string `json:"sequence"`
	} `json:"payload"`
}

// TradesResponse represents a response from /v3/trades
type TradesResponse struct {
	Envelope
	Payload []Trade `json:"payload"`
}

// Balance represents a response from /v3/balance
type BalanceResponse struct {
	Envelope
	Payload struct {
		Balances []Balance `json:"balances"`
	} `json:"payload"`
}

// Fees represents a response from /v3/fees
type FeesResponse struct {
	Envelope
	Payload struct {
		Fees           []Fee          `json:"fees"`
		WithdrawalFees WithdrawalFees `json:"withdrawal_fees"`
	} `json:"payload"`
}

// Ledger represents a response from /v3/ledger
type LedgerResponse struct {
	Envelope
	Payload []Transaction `json:"payload"`
}

// FundingsResponse represents a response from /v3/fundings/
type FundingsResponse struct {
	Envelope
	Payload []Funding `json:"payload"`
}

// UserTradesResponse represents a response from /v3/user_trades/
type UserTradesResponse struct {
	Envelope
	Payload []UserTrade `json:"payload"`
}

// UserOrderTradesResponse represents a response from /v3/user_order_trades/
type UserOrderTradesResponse struct {
	Envelope
	Payload []UserOrderTrade `json:"payload"`
}

// OrdersResponse represents a response from /v3/open_orders or /v3/lookup_orders
type OrdersResponse struct {
	Envelope
	Payload []UserOrder `json:"payload"`
}

// NewOrderResponse represents a response from /v3/orders
type NewOrderResponse struct {
	Envelope
	Payload struct {
		OID string `json:"oid"`
	} `json:"payload"`
}
