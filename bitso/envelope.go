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

// TradesResponse represents the response from /v3/trades
type TradesResponse struct {
	Envelope
	Payload []Trade `json:"payload"`
}
