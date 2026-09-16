package bitso

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	apiBaseURL = "https://bitso.com/api"
	apiVersion = "v3"
)

type apiRoute struct {
	path               string
	productRootBaseURL string
}

var apiRouteV4 = apiRoute{path: "v4"}

// DefaultHTTPClientTimeout is the request timeout used by NewClient.
const DefaultHTTPClientTimeout = 30 * time.Second

const (
	LogLevelPanic = zerolog.PanicLevel
	LogLevelFatal = zerolog.FatalLevel
	LogLevelError = zerolog.ErrorLevel
	LogLevelWarn  = zerolog.WarnLevel
	LogLevelInfo  = zerolog.InfoLevel
	LogLevelDebug = zerolog.DebugLevel
	LogLevelTrace = zerolog.TraceLevel
)

const defaultTickets = 1

// maxResponseSize limits the maximum response body size to prevent DoS attacks
const maxResponseSize = 10 * 1024 * 1024 // 10MB

const (
	nonceV2SaltMin = 100000
	nonceV2SaltMax = 999999
)

var (
	// Burst rate is disabled by default
	defaultBurstRate = time.Second * 0
)

type nonceGenerator func() (string, error)

// A Client is a Bitso API consumer
type Client struct {
	client *http.Client
	logger zerolog.Logger

	baseURL       string
	version       string
	customBaseURL bool

	key       string
	apiSecret string
	nonce     nonceGenerator

	tickets chan struct{}

	burstRate time.Duration

	mu sync.RWMutex
}

type clientConfig struct {
	client        *http.Client
	logger        zerolog.Logger
	baseURL       string
	version       string
	customBaseURL bool
	key           string
	apiSecret     string
	nonce         nonceGenerator
	tickets       chan struct{}
	burstRate     time.Duration
}

// NewClient creates and returns a new Bitso API client.
func NewClient() *Client {
	c := &Client{
		client:    defaultHTTPClient(),
		tickets:   make(chan struct{}, defaultTickets),
		baseURL:   strings.TrimPrefix(apiBaseURL, "/") + "/",
		logger:    zerolog.New(os.Stderr).With().Timestamp().Logger(),
		version:   apiVersion,
		nonce:     generateNonceV2,
		burstRate: defaultBurstRate,
	}

	c.SetLogLevel(LogLevelInfo)

	for i := 0; i < defaultTickets; i++ {
		c.tickets <- struct{}{}
	}
	return c
}

func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: DefaultHTTPClientTimeout}
}

func (c *Client) snapshot() clientConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	client := c.client
	if client == nil {
		client = defaultHTTPClient()
	}

	return clientConfig{
		client:        client,
		logger:        c.logger,
		baseURL:       c.baseURL,
		version:       c.version,
		customBaseURL: c.customBaseURL,
		key:           c.key,
		apiSecret:     c.apiSecret,
		nonce:         c.nonce,
		tickets:       c.tickets,
		burstRate:     c.burstRate,
	}
}

func generateNonceV2() (string, error) {
	saltRange := big.NewInt(nonceV2SaltMax - nonceV2SaltMin + 1)
	salt, err := rand.Int(rand.Reader, saltRange)
	if err != nil {
		return "", fmt.Errorf("generate nonce v2 salt: %w", err)
	}

	return fmt.Sprintf("%d%06d", time.Now().UnixMilli(), salt.Int64()+nonceV2SaltMin), nil
}

func (c *Client) endpointURL(endpoint string) (*url.URL, error) {
	return c.snapshot().endpointURL(endpoint)
}

func (cfg clientConfig) endpointURL(endpoint string) (*url.URL, error) {
	return cfg.endpointURLForRoute(apiRoute{path: cfg.version}, endpoint)
}

func (c *Client) endpointURLForRoute(route apiRoute, endpoint string) (*url.URL, error) {
	return c.snapshot().endpointURLForRoute(route, endpoint)
}

func (cfg clientConfig) endpointURLForRoute(route apiRoute, endpoint string) (*url.URL, error) {
	path := strings.Trim(route.path, "/")
	if path == "" {
		path = cfg.version
	}
	endpoint = strings.TrimLeft(endpoint, "/")

	if route.productRootBaseURL != "" {
		baseURL, err := cfg.productRootBaseURL(route.productRootBaseURL)
		if err != nil {
			return nil, err
		}
		return url.Parse(strings.TrimRight(baseURL, "/") + "/" + path + "/" + endpoint)
	}

	return url.Parse(cfg.baseURL + path + "/" + endpoint)
}

func (cfg clientConfig) productRootBaseURL(defaultBaseURL string) (string, error) {
	if !cfg.customBaseURL {
		return strings.TrimRight(defaultBaseURL, "/") + "/", nil
	}

	u, err := url.Parse(cfg.baseURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimRight(u.Path, "/")
	if path == "/api" {
		path = ""
	} else if strings.HasSuffix(path, "/api") {
		path = strings.TrimSuffix(path, "/api")
	}

	u.Path = strings.TrimRight(path, "/") + "/"
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

// SetLogLevel sets the log level for the client.
func (c *Client) SetLogLevel(level zerolog.Level) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger = c.logger.Level(level)
}

// SetAuth sets the user key and secret to use for private API calls.
func (c *Client) SetAuth(key, secret string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.key = key
	c.apiSecret = secret
}

// SetHTTPClient sets the HTTP client used for API requests.
//
// Passing nil restores the default client with DefaultHTTPClientTimeout.
func (c *Client) SetHTTPClient(client *http.Client) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if client == nil {
		client = defaultHTTPClient()
	}
	c.client = client
}

// HTTPClient returns the HTTP client used for API requests.
func (c *Client) HTTPClient() *http.Client {
	return c.snapshot().client
}

// SetAPIBaseURL sets the API prefix
func (c *Client) SetAPIBaseURL(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.baseURL = strings.TrimRight(prefix, "/") + "/"
	c.customBaseURL = true
}

// APIBaseURL returns the API prefix
func (c *Client) APIBaseURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.baseURL
}

// AvailableBooks returns a list of existing exchange order books and their
// respective order placement limits.
func (c *Client) AvailableBooks() ([]ExchangeOrderBook, error) {
	res := struct {
		Payload []ExchangeOrderBook `json:"payload"`
	}{}
	if err := c.getResponse("/available_books", nil, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// Tickers returns trading information from all books.
func (c *Client) Tickers() ([]Ticker, error) {
	res := struct {
		Payload []Ticker `json:"payload"`
	}{}
	if err := c.getResponse("/ticker", nil, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// Ticker returns trading information from the specified book.
func (c *Client) Ticker(book *Book) (*Ticker, error) {
	params := url.Values{
		"book": {book.String()},
	}
	res := struct {
		Payload Ticker `json:"payload"`
	}{}
	if err := c.getResponse("/ticker", params, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}

// Trades returns a list of recent trades from the specified book.
func (c *Client) Trades(params url.Values) ([]Trade, error) {
	res := struct {
		Payload []Trade `json:"payload"`
	}{}
	if err := c.getResponse("/trades", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// OrderBook returns a list of all open orders in the specified book.
func (c *Client) OrderBook(params url.Values) (*OrderBook, error) {
	res := struct {
		Payload OrderBook `json:"payload"`
	}{}
	if err := c.getResponse("/order_book", params, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}

// Balances returns information concerning the user’s balances for all supported
// currencies.
func (c *Client) Balances(params url.Values) ([]Balance, error) {
	res := struct {
		Payload struct {
			Balances []Balance `json:"balances"`
		} `json:"payload"`
	}{}
	if err := c.getResponse("/balance", params, &res); err != nil {
		return nil, err
	}
	return res.Payload.Balances, nil
}

// AccountStatus returns the user's account state, KYC document status, profile
// metadata, and transaction limits.
func (c *Client) AccountStatus() (*AccountStatus, error) {
	res := struct {
		Payload AccountStatus `json:"payload"`
	}{}
	if err := c.getResponse("/account_status", nil, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}

// Fees returns information on customer fees for all available order books,
// and withdrawal fees for applicable currencies.
func (c *Client) Fees(params url.Values) (*CustomerFees, error) {
	res := struct {
		Payload CustomerFees `json:"payload"`
	}{}
	if err := c.getResponse("/fees", params, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}

// Ledger returns a list of all the user's registered operations.
func (c *Client) Ledger(params url.Values) ([]Transaction, error) {
	res := struct {
		Payload []Transaction `json:"payload"`
	}{}
	if err := c.getResponse("/ledger", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// LedgerByOperation returns a list of all the user's registered operations.
func (c *Client) LedgerByOperation(op Operation, params url.Values) ([]Transaction, error) {
	optype := map[Operation]string{
		OperationFunding:    "fundings",
		OperationWithdrawal: "withdrawals",
		OperationTrade:      "trades",
		OperationFee:        "fees",
	}
	res := struct {
		Payload []Transaction `json:"payload"`
	}{}
	if err := c.getResponse("/ledger/"+optype[op], params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// Fundings returns detailed info on a user's fundings.
func (c *Client) Fundings(params url.Values) ([]Funding, error) {
	res := struct {
		Payload []Funding `json:"payload"`
	}{}
	if err := c.getResponse("/fundings/", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// Withdrawals returns detailed info on user's withdrawals
func (c *Client) Withdrawals(params url.Values) ([]Withdrawal, error) {
	res := struct {
		Payload []Withdrawal `json:"payload"`
	}{}
	if err := c.getResponse("/withdrawals", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// MyTrades returns a list of the user's trades.
func (c *Client) MyTrades(params url.Values) ([]UserTrade, error) {
	res := struct {
		Payload []UserTrade `json:"payload"`
	}{}
	if err := c.getResponse("/user_trades", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// OrderTrades returns a list of the user's order trades on a given order.
func (c *Client) OrderTrades(oid string, params url.Values) ([]UserOrderTrade, error) {
	res := struct {
		Payload []UserOrderTrade `json:"payload"`
	}{}
	if err := c.getResponse("/order_trades/"+oid, params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// MyOpenOrders a list of the user's open orders.
func (c *Client) MyOpenOrders(params url.Values) ([]UserOrder, error) {
	res := struct {
		Payload []UserOrder `json:"payload"`
	}{}
	if err := c.getResponse("/open_orders", params, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// LookupOrder returns details of an order given its order ID.
func (c *Client) LookupOrder(oid string) (*UserOrder, error) {
	orders, err := c.LookupOrders([]string{oid})
	if err != nil {
		return nil, err
	}
	if len(orders) > 0 {
		return &orders[0], nil
	}
	return nil, errors.New("no such order")
}

// LookupOrders returns a list of details for 1 or more orders
func (c *Client) LookupOrders(oids []string) ([]UserOrder, error) {
	res := struct {
		Payload []UserOrder `json:"payload"`
	}{}
	if err := c.getResponse("/orders/"+strings.Join(oids, ","), nil, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// CancelOrders cancels open order(s)
func (c *Client) CancelOrders(oids []string) ([]string, error) {
	var res struct {
		Payload []string `json:"payload"`
	}
	if err := c.deleteResponse("/orders/"+strings.Join(oids, ","), nil, &res); err != nil {
		return nil, err
	}
	return res.Payload, nil
}

// CancelOrder cancels an open order
func (c *Client) CancelOrder(oid string) ([]string, error) {
	return c.CancelOrders([]string{oid})
}

// PlaceOrder places a buy or sell order (both limit and market orders are
// available)
func (c *Client) PlaceOrder(order *OrderPlacement) (string, error) {
	var res struct {
		Payload struct {
			OID string `json:"oid"`
		} `json:"payload"`
	}
	if err := c.postResponse("/orders/", order, &res); err != nil {
		return "", err
	}
	return res.Payload.OID, nil
}

// BurstRate returns the current burst-rate limit.
func (c *Client) BurstRate() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.burstRate
}

// SetBurstRate sets the amount of time the client should wait in between
// requests.
func (c *Client) SetBurstRate(burstRate time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.burstRate = burstRate
}

func (cfg clientConfig) newRequest(logger *zerolog.Logger, method string, uri string, body io.Reader) (*http.Request, error) {
	var buf []byte

	if body != nil {
		var err error
		buf, err = io.ReadAll(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri, bytes.NewBuffer(buf))
	if err != nil {
		return nil, err
	}

	if method == "POST" || method == "PUT" || method == "PATCH" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
	}

	if cfg.key == "" && cfg.apiSecret == "" {
		// Return unsigned request
		return req, nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	generateNonce := cfg.nonce
	if generateNonce == nil {
		generateNonce = generateNonceV2
	}

	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}
	message := nonce + method + u.RequestURI() + string(buf)

	mac := hmac.New(sha256.New, []byte(cfg.apiSecret))
	mac.Write([]byte(message))

	signature := fmt.Sprintf("%x", mac.Sum(nil))

	authHeader := fmt.Sprintf("Bitso %s:%s:%s", cfg.key, nonce, signature)
	req.Header.Set("Authorization", authHeader)

	if logger.GetLevel() <= zerolog.TraceLevel {
		*logger = logger.With().
			Bool("auth.signed", true).
			Logger()
	}

	return req, nil
}

func (c *Client) doRequest(method string, endpoint string, params url.Values, body io.Reader, dest interface{}) error {
	return c.doRequestForRoute(method, apiRoute{}, endpoint, params, body, dest)
}

func (c *Client) doRequestForRoute(method string, route apiRoute, endpoint string, params url.Values, body io.Reader, dest interface{}) error {
	cfg := c.snapshot()

	logger := cfg.logger.With().
		Str("method", method).
		Str("endpoint", endpoint).
		Logger()

	u, err := cfg.endpointURLForRoute(route, endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	req, err := cfg.newRequest(&logger, method, u.String(), body)
	if err != nil {
		return err
	}

	// Apply burst-rate protection.
	if cfg.burstRate > 0 && cfg.tickets != nil {
		<-cfg.tickets
		ticker := time.NewTicker(cfg.burstRate)

		go func() {
			<-ticker.C
			ticker.Stop()

			cfg.tickets <- struct{}{}
		}()
	}

	res, err := cfg.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("request failed")
		return err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		logger.Error().Err(err).Msg("can not read response body")
		return err
	}

	logger = logger.With().
		Int("status", res.StatusCode).
		Logger()

	if logger.GetLevel() <= zerolog.DebugLevel {
		logger = logger.With().
			Str("body", string(buf)).
			Logger()
	}

	var env Envelope
	if err := json.Unmarshal(buf, &env); err != nil {
		logger.Error().Msg("can not unmarshal envelope")
		return err
	}

	if !env.Success {
		code, _ := strconv.Atoi(fmt.Sprintf("%v", env.Error.Code))
		logger.Error().
			Int("error.code", code).
			Msgf("api error: %s", env.Error.Message)
		return apiError(code, env.Error.Message)
	}

	if err := json.Unmarshal(buf, dest); err != nil {
		logger.Error().Msg("can not unmarshal payload")
		return err
	}

	return nil
}

func (c *Client) doDirectResponseForRoute(method string, route apiRoute, endpoint string, params url.Values, body io.Reader, dest interface{}, decodeAPIError func([]byte) error) error {
	return c.doDirectResponseForRouteWithHeaders(method, route, endpoint, params, body, nil, dest, decodeAPIError)
}

func (c *Client) doDirectResponseForRouteWithHeaders(method string, route apiRoute, endpoint string, params url.Values, body io.Reader, headers http.Header, dest interface{}, decodeAPIError func([]byte) error) error {
	cfg := c.snapshot()

	logger := cfg.logger.With().
		Str("method", method).
		Str("endpoint", endpoint).
		Logger()

	u, err := cfg.endpointURLForRoute(route, endpoint)
	if err != nil {
		return err
	}
	u.RawQuery = params.Encode()

	req, err := cfg.newRequest(&logger, method, u.String(), body)
	if err != nil {
		return err
	}
	for key, values := range headers {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	if cfg.burstRate > 0 && cfg.tickets != nil {
		<-cfg.tickets
		ticker := time.NewTicker(cfg.burstRate)

		go func() {
			<-ticker.C
			ticker.Stop()

			cfg.tickets <- struct{}{}
		}()
	}

	res, err := cfg.client.Do(req)
	if err != nil {
		logger.Error().Err(err).Msg("request failed")
		return err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(io.LimitReader(res.Body, maxResponseSize))
	if err != nil {
		logger.Error().Err(err).Msg("can not read response body")
		return err
	}

	logger = logger.With().
		Int("status", res.StatusCode).
		Logger()

	if logger.GetLevel() <= zerolog.DebugLevel {
		logger = logger.With().
			Str("body", string(buf)).
			Logger()
	}

	if decodeAPIError != nil {
		if err := decodeAPIError(buf); err != nil {
			logger.Error().Err(err).Msg("api error")
			return err
		}
	}
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("api error: status %d", res.StatusCode)
	}

	if err := json.Unmarshal(buf, dest); err != nil {
		logger.Error().Msg("can not unmarshal payload")
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

func (c *Client) getResponseForRoute(route apiRoute, endpoint string, params url.Values, dest interface{}) error {
	return c.doRequestForRoute("GET", route, endpoint, params, nil, dest)
}

func (c *Client) postResponse(endpoint string, body interface{}, dest interface{}) error {
	return c.postResponseForRoute(apiRoute{}, endpoint, body, dest)
}

func (c *Client) postResponseForRoute(route apiRoute, endpoint string, body interface{}, dest interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.doRequestForRoute("POST", route, endpoint, nil, bytes.NewBuffer(buf), dest)
}

func (c *Client) putResponseForRoute(route apiRoute, endpoint string, body interface{}, dest interface{}) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(buf)
	}
	return c.doRequestForRoute("PUT", route, endpoint, nil, reader, dest)
}

func (c *Client) patchResponseForRoute(route apiRoute, endpoint string, params url.Values, body interface{}, dest interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.doRequestForRoute("PATCH", route, endpoint, params, bytes.NewBuffer(buf), dest)
}
