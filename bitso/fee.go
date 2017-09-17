package bitso

// Fee represents a Bitso fee.
type Fee struct {
	Book       Book     `json:"book"`
	FeeDecimal Monetary `json:"fee_decimal"`
	FeePercent Monetary `json:"fee_percent"`
}

// CustomerFees represents a list of fees that Bitso
// charges the user.
type CustomerFees struct {
	Fees           []Fee               `json:"fees"`
	WithdrawalFees map[string]Monetary `json:"withdrawal_fees"`
}
