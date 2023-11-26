package testing

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xiam/bitso-go/bitso"
)

func TestPublicAPI(t *testing.T) {
	c := bitso.NewClient()
	c.SetAPIBaseURL("https://sandbox.bitso.com/api")

	t.Run("AvailableBooks", func(t *testing.T) {
		books, err := c.AvailableBooks()
		assert.NoError(t, err)
		assert.NotNil(t, books)
	})

	t.Run("Ticker", func(t *testing.T) {
		tickers, err := c.Tickers()
		assert.NoError(t, err)
		assert.NotNil(t, tickers)

		ticker, err := c.Ticker(bitso.NewBook(bitso.ETH, bitso.MXN))
		assert.NoError(t, err)
		assert.NotNil(t, ticker)
	})

	t.Run("OrderBook", func(t *testing.T) {
		orderBook, err := c.OrderBook(url.Values{
			"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
		})
		assert.NoError(t, err)
		assert.NotNil(t, orderBook)
	})

	t.Run("Trades", func(t *testing.T) {
		_, err := c.Trades(nil)
		assert.Error(t, err)

		ticker, err := c.Trades(url.Values{
			"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
		})
		assert.NoError(t, err)
		assert.NotNil(t, ticker)
	})
}
