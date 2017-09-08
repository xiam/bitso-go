package bitso

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
