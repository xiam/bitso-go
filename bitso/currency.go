package bitso

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Currency represents currencies
type Currency uint8

// Currencies
const (
	CurrencyNone Currency = iota

	XRP
	BCH
	BTC
	ETH
	MXN
	ETC
)

var currencyNames = map[Currency]string{
	XRP: "xrp",
	ETH: "eth",
	MXN: "mxn",
	BTC: "btc",
	BCH: "bch",
	ETC: "etc",
}

func getCurrencyByName(name string) (*Currency, error) {
	for c, n := range currencyNames {
		if n == name {
			return &c, nil
		}
	}
	return nil, errors.New("no such currency")
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
