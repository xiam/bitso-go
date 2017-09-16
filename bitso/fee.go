package bitso

// Fee represents a Bitso fee.
type Fee struct {
	Book       Book     `json:"book"`
	FeeDecimal Monetary `json:"fee_decimal"`
	FeePercent Monetary `json:"fee_percent"`
}

// WithdrawalFees represents a map of fees charged by withdrawals.
type WithdrawalFees map[Currency]Monetary

// CustomerFees represents a list of fees that Bitso
// charges the user.
type CustomerFees struct {
	Fees           []Fee          `json:"fees"`
	WithdrawalFees WithdrawalFees `json:"withdrawal_fees"`
}
