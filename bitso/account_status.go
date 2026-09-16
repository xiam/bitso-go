package bitso

// AccountStatus represents the user's account state, KYC document status,
// profile metadata, and transaction limits.
type AccountStatus struct {
	AccountCreationDate     Time     `json:"account_creation_date"`
	BusinessName            string   `json:"business_name"`
	BornInResidence         string   `json:"born_in_residence"`
	CashDepositAllowance    Monetary `json:"cash_deposit_allowance"`
	CellphoneNumber         string   `json:"cellphone_number"`
	CellphoneNumberStored   string   `json:"cellphone_number_stored"`
	ClientID                string   `json:"client_id"`
	CountryOfResidence      string   `json:"country_of_residence"`
	DailyLimit              Monetary `json:"daily_limit"`
	DailyRemaining          Monetary `json:"daily_remaining"`
	DateOfBirth             string   `json:"date_of_birth"`
	Email                   string   `json:"email"`
	EmailStored             string   `json:"email_stored"`
	EnabledTwoFactorMethods []string `json:"enabled_two_factor_methods"`
	EntityType              string   `json:"entity_type"`
	FirstName               string   `json:"first_name"`
	Gender                  string   `json:"gender"`
	GravatarImg             string   `json:"gravatar_img"`
	LastName                string   `json:"last_name"`
	MonthlyLimit            Monetary `json:"monthly_limit"`
	MonthlyRemaining        Monetary `json:"monthly_remaining"`
	OfficialID              string   `json:"official_id"`
	OriginOfFunds           string   `json:"origin_of_funds"`
	PreferredCurrency       Currency `json:"preferred_currency"`
	ProofOfResidency        string   `json:"proof_of_residency"`
	// Deprecated: Bitso keeps referral_code only for backwards compatibility.
	ReferralCode            string   `json:"referral_code"`
	SecondLastName          string   `json:"second_last_name"`
	SignedContract          string   `json:"signed_contract"`
	Status                  string   `json:"status"`
	TaxPayerType            string   `json:"tax_payer_type"`
	UserDefaultFiatCurrency Currency `json:"user_default_fiat_currency"`
	VerificationLevel       int      `json:"verification_level"`
}
