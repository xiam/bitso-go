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

	AAVE    = "aave"
	ADA     = "ada"
	ALGO    = "algo"
	APE     = "ape"
	ARB     = "arb"
	ARS     = "ars"
	ATOM    = "atom"
	AVAX    = "avax"
	AXS     = "axs"
	BAL     = "bal"
	BAR     = "bar"
	BAT     = "bat"
	BCH     = "bch"
	BONK    = "bonk"
	BRL     = "brl"
	BRL1    = "brl1"
	BTC     = "btc"
	CHZ     = "chz"
	COMP    = "comp"
	COP     = "cop"
	CRV     = "crv"
	DOGE    = "doge"
	DOT     = "dot"
	DYDX    = "dydx"
	ENJ     = "enj"
	ETH     = "eth"
	EUR     = "eur"
	FET     = "fet"
	FLOKI   = "floki"
	GALA    = "gala"
	GRT     = "grt"
	HBAR    = "hbar"
	HYPE    = "hype"
	LDO     = "ldo"
	LINK    = "link"
	LRC     = "lrc"
	LTC     = "ltc"
	MANA    = "mana"
	MXN     = "mxn"
	NEAR    = "near"
	NEIRO   = "neiro"
	OMG     = "omg"
	ONDO    = "ondo"
	PAXG    = "paxg"
	PEPE    = "pepe"
	POL     = "pol"
	POPCAT  = "popcat"
	PSG     = "psg"
	PYUSD   = "pyusd"
	QNT     = "qnt"
	RENDER  = "render"
	RLUSD   = "rlusd"
	S       = "s"
	SAND    = "sand"
	SHIB    = "shib"
	SKY     = "sky"
	SNX     = "snx"
	SOL     = "sol"
	SUSHI   = "sushi"
	TIGRES  = "tigres"
	TON     = "ton"
	TRX     = "trx"
	TUSD    = "tusd"
	UNI     = "uni"
	USD     = "usd"
	USDS    = "usds"
	USDT    = "usdt"
	VIRTUAL = "virtual"
	WIF     = "wif"
	XLM     = "xlm"
	XRP     = "xrp"
	YFI     = "yfi"
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
