package bitso

// Trade represents a recent trade from the specified book
type Trade struct {
	Book      Book     `json:"book"`
	CreatedAt Time     `json:"created_at"`
	Amount    Monetary `json:"amount"`
	MakerSide Side     `json:"maker_side"`
	Price     Monetary `json:"price"`
	TID       uint64   `json:"tid"`
}
