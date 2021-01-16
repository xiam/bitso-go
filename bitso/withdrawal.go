package bitso

// Funding represents the withdrawals of the user
type Withdrawal struct {
	WID       string                 `json:"wid"`
	Status    string                 `json:"status"`
	CreatedAt Time                   `json:"created_at"`
	Currency  Currency               `json:"currency"`
	Method    string                 `json:"method"`
	Amount    Monetary               `json:"amount"`
	Details   map[string]interface{} `json:"details"`
}
