package bitso

import (
	"time"
)

// Ticker holds trading information from an specific book.
type Ticker struct {
	// Order book symbol
	Book Book `json:"book"`

	// Last 24 hours volume
	Volume Monetary `json:"volume"`

	// Last 24 hours price high
	High Monetary `json:"high"`

	// Last traded price
	Last Monetary `json:"last"`

	// Last 24 hours price low
	Low Monetary `json:"low"`

	// Last 24 hours volume weighted average price
	Vwap Monetary `json:"vwap"`

	// Lowest sell order
	Ask Monetary `json:"ask"`

	// Highest buy order
	Bid Monetary `json:"bid"`

	// When this ticker was generated
	CreatedAt time.Time `json:"created_at"`
}
