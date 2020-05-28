package bitso

import (
	"encoding/json"
	"fmt"
)

// Currency represents currencies
type Currency uint8

// Currencies
const (
	CurrencyNone Currency = iota

	ARS
	BAT
	BCH
	BTC
	DAI
	ETH
	GNT
	LTC
	MANA
	MXN
	TUSD
	USD
	XRP
)

var currencyNames = map[Currency]string{
	ARS:  "ars",
	BAT:  "bat",
	BCH:  "bch",
	BTC:  "btc",
	DAI:  "dai",
	ETH:  "eth",
	GNT:  "gnt",
	LTC:  "ltc",
	MANA: "mana",
	MXN:  "mxn",
	TUSD: "tusd",
	USD:  "usd",
	XRP:  "xrp",
}

func getCurrencyByName(name string) (*Currency, error) {
	for c, n := range currencyNames {
		if n == name {
			return &c, nil
		}
	}
	return nil, fmt.Errorf("no such currency: %q", name)
}

// MarshalJSON implements json.Marshaler
func (c Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (c *Currency) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	for k, v := range currencyNames {
		if v == z {
			*c = k
			return nil
		}
	}
	return fmt.Errorf("unsupported currency: %v", z)
}

func (c Currency) String() string {
	if z, ok := currencyNames[c]; ok {
		return z
	}
	panic("unsupported currency")
}
