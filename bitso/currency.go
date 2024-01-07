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

	AAVE   = "aave"
	ADA    = "ada"
	ALGO   = "algo"
	APE    = "ape"
	ARB    = "arb"
	ARS    = "ars"
	ATOM   = "atom"
	AVAX   = "avax"
	AXS    = "axs"
	BAL    = "bal"
	BAR    = "bar"
	BAT    = "bat"
	BCH    = "bch"
	BRL    = "brl"
	BTC    = "btc"
	CHZ    = "chz"
	COMP   = "comp"
	CRV    = "crv"
	DAI    = "dai"
	DOGE   = "doge"
	DOT    = "dot"
	DYDX   = "dydx"
	ENJ    = "enj"
	ETH    = "eth"
	EUR    = "eur"
	FTM    = "ftm"
	GALA   = "gala"
	GRT    = "grt"
	LDO    = "ldo"
	LINK   = "link"
	LRC    = "lrc"
	LTC    = "ltc"
	MANA   = "mana"
	MATIC  = "matic"
	MKR    = "mkr"
	MXN    = "mxn"
	NEAR   = "near"
	OMG    = "omg"
	PAXG   = "paxg"
	PEPE   = "pepe"
	PSG    = "psg"
	QNT    = "qnt"
	SAND   = "sand"
	SHIB   = "shib"
	SNX    = "snx"
	SOL    = "sol"
	SUSHI  = "sushi"
	TIGRES = "tigres"
	TRX    = "trx"
	TUSD   = "tusd"
	UNI    = "uni"
	USD    = "usd"
	USDT   = "usdt"
	XLM    = "xlm"
	XRP    = "xrp"
	YFI    = "yfi"
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
