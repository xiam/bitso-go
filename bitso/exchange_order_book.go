package bitso

// BookFlatRate represents the default maker/taker fees for a book.
type BookFlatRate struct {
	Maker Monetary `json:"maker"`
	Taker Monetary `json:"taker"`
}

// BookFeesTier represents a volume-based fee tier.
type BookFeesTier struct {
	Volume Monetary `json:"volume"`
	Maker  Monetary `json:"maker"`
	Taker  Monetary `json:"taker"`
}

// BookFees represents the fee structure for a book.
type BookFees struct {
	FlatRate  BookFlatRate   `json:"flat_rate"`
	Structure []BookFeesTier `json:"structure"`
}

// ExchangeOrderBook represents order placement limits on books.
type ExchangeOrderBook struct {
	// Order book symbol
	Book Book `json:"book"`

	// Default chart type (depth, candle, hollow, line, trading view)
	DefaultChart string `json:"default_chart"`

	// Minimum amount of major when placing orders
	MinimumAmount Monetary `json:"minimum_amount"`
	// Maximum amount of major when placing orders
	MaximumAmount Monetary `json:"maximum_amount"`

	// Minimum price when placing orders
	MinimumPrice Monetary `json:"minimum_price"`
	// Maximum price when placing orders
	MaximumPrice Monetary `json:"maximum_price"`

	// Minimum value amount (amount*price) when placing orders
	MinimumValue Monetary `json:"minimum_value"`

	// Maximum value amount (amount*price) when placing orders
	MaximumValue Monetary `json:"maximum_value"`

	// Minimum price increment between consecutive bid/offer prices
	TickSize Monetary `json:"tick_size"`

	// Fee structure for this book
	Fees BookFees `json:"fees"`
}
