package bitso

// Balance represents the balance of a given currency.
type Balance struct {
	Currency  Currency `json:"currency"`
	Total     Monetary `json:"total"`
	Locked    Monetary `json:"locked"`
	Available Monetary `json:"available"`

	// Deprecated: current Bitso balance responses no longer document this field.
	PendingDeposit Monetary `json:"pending_deposit"`
	// Deprecated: current Bitso balance responses no longer document this field.
	PendingWithdrawal Monetary `json:"pending_withdrawal"`
}
