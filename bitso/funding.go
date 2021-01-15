package bitso

// Funding represents the fundings of the user
type Funding struct {
	FID       string                 `json:"fid"`
	Currency  Currency               `json:"currency"`
	Method    string                 `json:"method"`
	Amount    Monetary               `json:"amount"`
	Status    string                 `json:"status"`
	CreatedAt Time                   `json:"created_at"`
	Details   map[string]interface{} `json:"details"`
}
