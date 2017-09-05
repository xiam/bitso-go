package bitso

import (
	"time"
)

type Trade struct {
	Book      Book      `json:"book"`
	CreatedAt time.Time `json:"created_at"`
	Amount    Monetary  `json:"amount"`
	MakerSide Side      `json:"maker_side"`
	Price     Monetary  `json:"price"`
	TID       uint64    `json:"tid"`
}

type TradeResponse struct {
	Envelope
	Payload []Trade `json:"payload"`
}
