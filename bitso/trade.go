package bitso

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
	Book         Book      `json:"book"`
	Major        Monetary  `json:"major"`
	CreatedAt    Time      `json:"created_at"`
	Minor        Monetary  `json:"minor"`
	FeesAmount   Monetary  `json:"fees_amount"`
	FeesCurrency Currency  `json:"currency"`
	Price        Monetary  `json:"price"`
	TID          TID       `json:"tid"`
	OID          string    `json:"oid"`
	Side         OrderSide `json:"side"`
}

// UserOrderTrade represents a trade made by the user
type UserOrderTrade struct {
	Book         Book      `json:"book"`
	Major        Monetary  `json:"major"`
	CreatedAt    Time      `json:"created_at"`
	Minor        Monetary  `json:"minor"`
	FeesAmount   Monetary  `json:"fees_amount"`
	FeesCurrency Currency  `json:"currency"`
	Price        Monetary  `json:"price"`
	TID          TID       `json:"tid"`
	OID          string    `json:"oid"`
	Side         OrderSide `json:"side"`
}
