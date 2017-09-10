package bitso

type Fee struct {
	Book       Book     `json:"book"`
	FeeDecimal Monetary `json:"fee_decimal"`
	FeePercent Monetary `json:"fee_percent"`
}

type WithdrawalFees map[string]Monetary
