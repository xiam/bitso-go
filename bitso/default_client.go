package bitso

import (
	"net/url"
)

func Trades(params url.Values) (*TradesResponse, error) {
	return DefaultClient.Trades(params)
}
