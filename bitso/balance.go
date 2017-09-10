package bitso

type Balance struct {
	Currency  Currency `json:"currency"`
	Total     Monetary `json:"total"`
	Locked    Monetary `json:"locked"`
	Available Monetary `json:"available"`
}
