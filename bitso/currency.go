package bitso

import (
	"errors"
)

// Currency represents currencies
type Currency uint8

// Currencies
const (
	XRP Currency = iota

	BCH
	BTC
	ETH
	MXN
)

var currencyNames = map[Currency]string{
	XRP: "xrp",
	ETH: "eth",
	MXN: "mxn",
	BTC: "btc",
	BCH: "bch",
}

func getCurrencyByName(name string) (*Currency, error) {
	for c, n := range currencyNames {
		if n == name {
			return &c, nil
		}
	}
	return nil, errors.New("no such currency")
}
