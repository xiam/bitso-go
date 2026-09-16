package bitso

import "encoding/json"

// Fee represents a Bitso fee.
type Fee struct {
	Book Book `json:"book"`

	MakerFeeDecimal     Monetary `json:"maker_fee_decimal"`
	MakerFeePercent     Monetary `json:"maker_fee_percent"`
	TakerFeeDecimal     Monetary `json:"taker_fee_decimal"`
	TakerFeePercent     Monetary `json:"taker_fee_percent"`
	VolumeCurrency      Currency `json:"volume_currency"`
	CurrentVolume       Monetary `json:"current_volume"`
	NextVolume          Monetary `json:"next_volume"`
	NextMakerFeePercent Monetary `json:"next_maker_fee_percent"`
	NextTakerFeePercent Monetary `json:"next_taker_fee_percent"`

	// Deprecated: use MakerFeeDecimal.
	FeeDecimal Monetary `json:"fee_decimal"`
	// Deprecated: use MakerFeePercent.
	FeePercent Monetary `json:"fee_percent"`
	// Deprecated: use NextMakerFeePercent.
	NextFee Monetary `json:"nextFee"`
	// Deprecated: use NextTakerFeePercent.
	NextTakerFee Monetary `json:"nextTakerFee"`
}

// UnmarshalJSON decodes the current snake_case fee fields while preserving the
// legacy camelCase nextVolume response field.
func (f *Fee) UnmarshalJSON(in []byte) error {
	type fee Fee

	var decoded fee
	if err := json.Unmarshal(in, &decoded); err != nil {
		return err
	}

	if decoded.NextVolume == "" {
		var legacy struct {
			NextVolume Monetary `json:"nextVolume"`
		}
		if err := json.Unmarshal(in, &legacy); err != nil {
			return err
		}
		decoded.NextVolume = legacy.NextVolume
	}

	*f = Fee(decoded)
	return nil
}

// DepositFee represents a Bitso deposit fee.
type DepositFee struct {
	Currency Currency `json:"currency"`
	Method   string   `json:"method"`
	Fee      Monetary `json:"fee"`
	IsFixed  bool     `json:"is_fixed"`
}

// CustomerFees represents a list of fees that Bitso
// charges the user.
type CustomerFees struct {
	Fees           []Fee               `json:"fees"`
	WithdrawalFees map[string]Monetary `json:"withdrawal_fees"`
	DepositFees    []DepositFee        `json:"deposit_fees"`
}
