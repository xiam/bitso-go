package bitso

// ExchangeOrderBook represents order placement limits on books.
type ExchangeOrderBook struct {

	// Order book symbol
	Book Book `json:"book"`

	// Minimum amount of major when placing orders
	MinimumAmount Monetary `json:"minimum_amount"`
	// Maximum amount of major when placing orders
	MaximumAmount Monetary `json:"maximum_amount"`

	// Minimum price when placing orders
	MinimumPrice Monetary `json:"minimum_price"`
	// Maximum price when placing orders
	MaximumPrice Monetary `json:"maximum_price"`

	// Minimum value amount (amount*price) when placing orders
	MinimumValue Monetary `json:"minimun_value"`

	// Maximum value amount (amount*price) when placing orders
	MaximumValue Monetary `json:"maximum_value"`
}
