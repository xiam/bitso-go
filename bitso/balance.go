package bitso

// Balance represents the balance of a given currency.
type Balance struct {
	Currency          Currency `json:"currency"`
	Total             Monetary `json:"total"`
	Locked            Monetary `json:"locked"`
	Available         Monetary `json:"available"`
	PendingDeposit    Monetary `json:"pending_deposit"`
	PendingWithdrawal Monetary `json:"pending_withdrawal"`
}
