package testing

import (
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/mazingstudio/bitso-go/bitso"
	"github.com/stretchr/testify/assert"
)

var (
	apiKey    = os.Getenv("API_KEY")
	apiSecret = os.Getenv("API_SECRET")
)

func testAvailableBooks(t *testing.T, c *bitso.Client) {
	books, err := c.AvailableBooks()
	assert.NoError(t, err)
	assert.NotNil(t, books)
}

func testTicker(t *testing.T, c *bitso.Client) {
	tickers, err := c.Tickers()
	assert.NoError(t, err)
	assert.NotNil(t, tickers)

	ticker, err := c.Ticker(bitso.NewBook(bitso.ETH, bitso.MXN))
	assert.NoError(t, err)
	assert.NotNil(t, ticker)
}

func testOrderBook(t *testing.T, c *bitso.Client) {
	orderBook, err := c.OrderBook(url.Values{
		"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
	})
	assert.NoError(t, err)
	assert.NotNil(t, orderBook)
}

func testTrades(t *testing.T, c *bitso.Client) {
	_, err := c.Trades(nil)
	assert.Error(t, err)

	ticker, err := c.Trades(url.Values{
		"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
	})
	assert.NoError(t, err)
	assert.NotNil(t, ticker)
}

func testAuth(t *testing.T, c *bitso.Client) {
	_, err := c.Balances(nil)
	assert.Error(t, err)

	c.SetAPIKey(apiKey)
	c.SetAPISecret(apiSecret)

	_, err = c.Balances(nil)
	assert.NoError(t, err)
}

func testBalances(t *testing.T, c *bitso.Client) {
	balances, err := c.Balances(nil)

	assert.NoError(t, err)
	assert.NotNil(t, balances)
}

func testFees(t *testing.T, c *bitso.Client) {
	fees, err := c.Fees(nil)

	assert.NoError(t, err)
	assert.NotNil(t, fees)
}

func testLedger(t *testing.T, c *bitso.Client) {
	ledger, err := c.Ledger(nil)

	assert.NoError(t, err)
	assert.NotNil(t, ledger)
}

func testLedgerByOperation(t *testing.T, c *bitso.Client) {
	ledger, err := c.LedgerByOperation(bitso.OperationFunding, nil)

	assert.NoError(t, err)
	assert.NotNil(t, ledger)
}

func testFundings(t *testing.T, c *bitso.Client) {
	fundings, err := c.Fundings(nil)

	assert.NoError(t, err)
	assert.NotNil(t, fundings)
}

func testMyTrades(t *testing.T, c *bitso.Client) {
	trades, err := c.MyTrades(nil)

	assert.NoError(t, err)
	assert.NotNil(t, trades)
}

func testMyOpenOrders(t *testing.T, c *bitso.Client) {
	_, err := c.MyOpenOrders(nil)
	assert.Error(t, err)

	orders, err := c.MyOpenOrders(url.Values{
		"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
	})

	assert.NoError(t, err)
	assert.NotNil(t, orders)
}

func testOrderTrades(t *testing.T, c *bitso.Client) {
	orders, err := c.MyOpenOrders(url.Values{
		"book": {bitso.NewBook(bitso.ETH, bitso.MXN).String()},
	})
	assert.NoError(t, err)
	assert.NotNil(t, orders)

	for _, order := range orders {
		orderTrades, err := c.OrderTrades(order.OID, nil)
		assert.NoError(t, err)
		assert.NotNil(t, orderTrades)

		order, err := c.LookupOrder(order.OID)
		assert.NoError(t, err)
		assert.NotNil(t, order)
	}
}

var publicTests = []func(*testing.T, *bitso.Client){
	testAvailableBooks,
	testTicker,
	testOrderBook,
	testTrades,
}

var privateTests = []func(*testing.T, *bitso.Client){
	testAvailableBooks,
	testTicker,
	testOrderBook,
	testTrades,

	testAuth,
	testBalances,
	testFees,
	testLedger,
	testLedgerByOperation,
	testFundings,
	testMyTrades,
	testMyOpenOrders,
	testOrderTrades,
}

func TestPublicProdAPI(t *testing.T) {
	c := bitso.NewClient(nil)
	for _, testFn := range publicTests {
		testFn(t, c)
	}
}

func TestPublicDevAPI(t *testing.T) {
	c := bitso.NewClient(nil)
	c.SetAPIPrefix("https://api-dev.bitso.com/")
	for _, testFn := range publicTests {
		testFn(t, c)
	}
}

func TestPrivateProdAPI(t *testing.T) {
	if apiKey == "" || apiSecret == "" {
		t.Skip("no key or secret provided")
	}

	c := bitso.NewClient(nil)
	c.SetBurstRate(time.Millisecond * 500)
	for _, testFn := range privateTests {
		testFn(t, c)
	}
}
