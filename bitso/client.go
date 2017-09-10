package bitso

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	debug = false
)

const (
	apiPrefix      = "https://api.bitso.com/"
	defaultTickets = 1
)

var (
	DefaultClient    = NewClient(http.DefaultClient)
	defaultBurstRate = time.Millisecond * 1500
)

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{
		client:  httpClient,
		tickets: make(chan struct{}, defaultTickets),

		version:   "v3",
		BurstRate: defaultBurstRate,
	}
	for i := 0; i < defaultTickets; i++ {
		c.tickets <- struct{}{}
	}
	return c
}

type Client struct {
	client *http.Client

	key     string
	secret  string
	version string

	tickets chan struct{}

	BurstRate time.Duration
}

func (c *Client) SetKey(key string) {
	c.key = key
}

func (c *Client) SetSecret(secret string) {
	c.secret = secret
}

func (c *Client) EndpointURL(endpoint string) (*url.URL, error) {
	return url.Parse(apiPrefix + c.version + "/" + endpoint)
}

func (c *Client) lock() {
	tk := time.NewTicker(c.BurstRate)
	<-tk.C
}

func (c *Client) Get(uri string) (*http.Response, error) {
	<-c.tickets

	ticker := time.NewTicker(c.BurstRate)
	go func() {
		<-ticker.C
		c.tickets <- struct{}{}
	}()

	return c.client.Get(uri)
}

func (c *Client) debugf(f string, a ...interface{}) {
	if !debug {
		return
	}
	log.Printf(f, a...)
}

func (c *Client) newRequest(method string, uri string, body io.Reader) (*http.Request, error) {
	nonce := time.Now().UnixNano()

	var buf []byte

	if body != nil {
		var err error
		buf, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	if c.key == "" && c.secret == "" {
		// Return unsigned request
		return req, nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	message := fmt.Sprintf("%d%s%s%s", nonce, method, u.Path, string(buf))

	mac := hmac.New(sha256.New, []byte(c.secret))
	mac.Write([]byte(message))
	signature := fmt.Sprintf("%x", mac.Sum(nil))

	authHeader := fmt.Sprintf("Bitso %s:%d:%s", c.key, nonce, signature)
	req.Header.Set("Authorization", authHeader)

	return req, err
}

func (c *Client) doRequest(method string, endpoint string, params url.Values, body io.Reader, dest interface{}) error {
	u, err := c.EndpointURL(endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	req, err := c.newRequest(method, u.String(), body)
	if err != nil {
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	c.debugf("res: %v", string(buf))

	if err := json.Unmarshal(buf, dest); err != nil {
		return err
	}

	return nil
}

func (c *Client) deleteResponse(endpoint string, params url.Values, dest interface{}) error {
	return c.doRequest("DELETE", endpoint, params, nil, dest)
}

func (c *Client) getResponse(endpoint string, params url.Values, dest interface{}) error {
	return c.doRequest("GET", endpoint, params, nil, dest)
}

func (c *Client) postResponse(endpoint string, body interface{}, dest interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.doRequest("POST", endpoint, nil, bytes.NewBuffer(buf), dest)
}

// AvailableBooks returns a list of existing exchange order books and their
// respective order placement limits.
func (c *Client) AvailableBooks() (*AvailableBooksResponse, error) {
	var res AvailableBooksResponse
	if err := c.getResponse("/available_books", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Ticker returns trading information from the specified book.
func (c *Client) Ticker(params url.Values) (*TickerResponse, error) {
	var res TickerResponse
	if err := c.getResponse("/ticker", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Trades returns a list of recent trades from the specified book.
func (c *Client) Trades(params url.Values) (*TradesResponse, error) {
	var res TradesResponse
	if err := c.getResponse("/trades", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// OrderBook returns a list of all open orders in the specified book.
func (c *Client) OrderBook(params url.Values) (*OrderBookResponse, error) {
	var res OrderBookResponse
	if err := c.getResponse("/order_book", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Balance returns information concerning the user’s balances for all supported
// currencies.
func (c *Client) Balance(params url.Values) (*BalanceResponse, error) {
	var res BalanceResponse
	if err := c.getResponse("/balance", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Fees returns information on customer fees for all available order books,
// and withdrawal fees for applicable currencies.
func (c *Client) Fees(params url.Values) (*FeesResponse, error) {
	var res FeesResponse
	if err := c.getResponse("/fees", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Ledger returns a list of all the user's registered operations.
func (c *Client) Ledger(params url.Values) (*LedgerResponse, error) {
	var res LedgerResponse
	if err := c.getResponse("/ledger", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// LedgerByOperation returns a list of all the user's registered operations.
func (c *Client) LedgerByOperation(op Operation, params url.Values) (*LedgerResponse, error) {
	optype := map[Operation]string{
		OperationFunding:    "fundings",
		OperationWithdrawal: "withdrawals",
		OperationTrade:      "trades",
		OperationFee:        "fees",
	}
	var res LedgerResponse
	if err := c.getResponse("/ledger/"+optype[op], params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// Fundings returns detailed info on a user's fundings.
func (c *Client) Fundings(params url.Values) (*FundingsResponse, error) {
	var res FundingsResponse
	if err := c.getResponse("/fundings", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// UserTrades returns a list of the user's trades.
func (c *Client) UserTrades(params url.Values) (*UserTradesResponse, error) {
	var res UserTradesResponse
	if err := c.getResponse("/user_trades", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// UserOrderTrades returns a list of the user's order trades.
func (c *Client) UserOrderTrades(oid string, params url.Values) (*UserOrderTradesResponse, error) {
	var res UserOrderTradesResponse
	if err := c.getResponse("/order_trades/"+oid, params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// OpenOrders a list of the user's open orders.
func (c *Client) OpenOrders(params url.Values) (*OrdersResponse, error) {
	var res OrdersResponse
	if err := c.getResponse("/open_orders", params, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// LookupOrders returns a list of details for 1 or more orders
func (c *Client) LookupOrders(oids []string) (*OrdersResponse, error) {
	var res OrdersResponse
	if err := c.getResponse("/lookup_orders/"+strings.Join(oids, "-"), nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// CancelOrders cancels open order(s)
func (c *Client) CancelOrders(oids []string) (*OrdersResponse, error) {
	var res OrdersResponse
	if err := c.deleteResponse("/orders/"+strings.Join(oids, "-"), nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// PlaceOrder places a buy or sell order (both limit and market orders are
// available)
func (c *Client) PlaceOrder(order *OrderPlacement) (*NewOrderResponse, error) {
	var res NewOrderResponse
	if err := c.postResponse("/orders/", order, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
