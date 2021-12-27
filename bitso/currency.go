package bitso

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

// Currency represents currencies
type Currency string

// Currencies
const (
	CurrencyNone Currency = ""

	AAVE = "aave"
	ARS  = "ars"
	AXS  = "axs"
	BAT  = "bat"
	BCH  = "bch"
	BRL  = "brl"
	BTC  = "btc"
	CHZ  = "chz"
	COMP = "comp"
	DAI  = "dai"
	DYDX = "dydx"
	ETH  = "eth"
	LINK = "link"
	LTC  = "ltc"
	MANA = "mana"
	MXN  = "mxn"
	SAND = "sand"
	TUSD = "tusd"
	UNI  = "uni"
	USD  = "usd"
	USDT = "usdt"
	XRP  = "xrp"
	YFI  = "yfi"
)

func ToCurrency(name string) Currency {
	return Currency(strings.ToLower(name))
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *Currency) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	*c = ToCurrency(z)
	return nil
}

func (c Currency) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *Currency) Scan(value interface{}) error {
	*c = ToCurrency(value.(string))
	return nil
}
