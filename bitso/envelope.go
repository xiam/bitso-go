package bitso

import (
	"time"
)

// Envelope represents a common response envelope from Bitso API.
type Envelope struct {
	Success bool `json:"success"`
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
		Asks      []Ask     `json:"asks"`
		Bids      []Bid     `json:"bids"`
		UpdatedAt time.Time `json:"updated_at"`
		Sequence  string    `json:"sequence"`
	} `json:"payload"`
}
