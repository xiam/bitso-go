package bitso

type AvailableBook struct {
	// Order book symbol
	Book string `json:"book"`

	// Minimum amount of major when placing orders
	MinimumAmount string `json:"minimum_amount"`
	// Maximum amount of major when placing orders
	MaximumAmount string `json:"maximum_amount"`

	// Minimum price when placing orders
	MinimumPrice string `json:"minimum_price"`
	// Maximum price when placing orders
	MaximumPrice string `json:"maximum_price"`

	// Minimum value amount (amount*price) when placing orders
	MinimumValue string `json:"minimun_value"`

	// Maximum value amount (amount*price) when placing orders
	MaximumValue string `json:"maximum_value"`
}
