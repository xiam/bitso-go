package bitso

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// mockServer creates a test server that returns predefined responses
func mockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClient()
	client.SetAPIBaseURL(server.URL + "/api")
	return server, client
}

// successResponse wraps payload in a success envelope
func successResponse(payload interface{}) []byte {
	resp := map[string]interface{}{
		"success": true,
		"payload": payload,
	}
	data, _ := json.Marshal(resp)
	return data
}

// errorResponse creates an error envelope
func errorResponse(code int, message string) []byte {
	resp := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

func testSignature(secret, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

func TestNewClient(t *testing.T) {
	c := NewClient()

	require.NotNil(t, c, "NewClient should return non-nil client")
	assert.NotNil(t, c.client, "HTTP client should be initialized")
	assert.NotSame(t, http.DefaultClient, c.client, "Default client should not reuse http.DefaultClient")
	assert.Equal(t, DefaultHTTPClientTimeout, c.client.Timeout, "Default HTTP client should have a finite timeout")
	assert.NotNil(t, c.tickets, "Tickets channel should be initialized")
	assert.Equal(t, "https://bitso.com/api/", c.baseURL, "Default base URL should be set")
	assert.Equal(t, "v3", c.version, "Default version should be v3")
	assert.Equal(t, time.Duration(0), c.burstRate, "Default burst rate should be 0")
}

func TestSetHTTPClient(t *testing.T) {
	c := NewClient()
	custom := &http.Client{Timeout: 123 * time.Millisecond}

	c.SetHTTPClient(custom)
	assert.Same(t, custom, c.HTTPClient())

	c.SetHTTPClient(nil)
	require.NotNil(t, c.HTTPClient())
	assert.NotSame(t, custom, c.HTTPClient())
	assert.Equal(t, DefaultHTTPClientTimeout, c.HTTPClient().Timeout)
}

func TestCustomHTTPClientControlsRequests(t *testing.T) {
	c := NewClient()
	c.SetAPIBaseURL("https://example.test/custom")

	called := false
	custom := &http.Client{
		Timeout: 123 * time.Millisecond,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			called = true
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "example.test", req.URL.Host)
			assert.Equal(t, "/custom/v3/available_books", req.URL.Path)

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(successResponse([]map[string]interface{}{})))),
				Request:    req,
			}, nil
		}),
	}

	c.SetHTTPClient(custom)
	books, err := c.AvailableBooks()

	require.NoError(t, err)
	assert.Empty(t, books)
	assert.True(t, called)
	assert.Same(t, custom, c.HTTPClient())
	assert.Equal(t, 123*time.Millisecond, custom.Timeout)
}

func TestHTTPRequestTimeoutReturnsError(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write(successResponse([]interface{}{}))
	})
	defer server.Close()

	client.SetHTTPClient(&http.Client{Timeout: 20 * time.Millisecond})
	_, err := client.AvailableBooks()

	require.Error(t, err)
	var urlErr *url.Error
	require.ErrorAs(t, err, &urlErr)
	assert.True(t, urlErr.Timeout(), "expected timeout error, got %v", err)
}

func TestSetAuth(t *testing.T) {
	c := NewClient()

	c.SetAuth("test-key", "tsecret")

	assert.Equal(t, "test-key", c.key)
	assert.Equal(t, "tsecret", c.apiSecret)
}

func TestSetAPIBaseURL(t *testing.T) {
	c := NewClient()

	t.Run("without trailing slash", func(t *testing.T) {
		c.SetAPIBaseURL("https://sandbox.bitso.com/api")
		assert.Equal(t, "https://sandbox.bitso.com/api/", c.APIBaseURL())
	})

	t.Run("with trailing slash", func(t *testing.T) {
		c.SetAPIBaseURL("https://sandbox.bitso.com/api/")
		assert.Equal(t, "https://sandbox.bitso.com/api/", c.APIBaseURL())
	})
}

func TestBurstRate(t *testing.T) {
	c := NewClient()

	assert.Equal(t, time.Duration(0), c.BurstRate())

	c.SetBurstRate(100 * time.Millisecond)
	assert.Equal(t, 100*time.Millisecond, c.BurstRate())
}

func TestClientMutableConfigurationConcurrentAccess(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(successResponse([]map[string]interface{}{}))
	})
	defer server.Close()

	client.SetHTTPClient(server.Client())

	const iterations = 100
	logLevels := []zerolog.Level{LogLevelError, LogLevelInfo, LogLevelDebug, LogLevelTrace}

	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				client.SetAuth(
					fmt.Sprintf("key-%d-%d", worker, j),
					fmt.Sprintf("secret-%d-%d", worker, j),
				)
				client.SetAPIBaseURL(server.URL + "/api")
				client.SetLogLevel(logLevels[(worker+j)%len(logLevels)])
				client.SetHTTPClient(server.Client())
				if j%2 == 0 {
					client.SetBurstRate(0)
				} else {
					client.SetBurstRate(time.Nanosecond)
				}
			}
		}(i)
	}

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				_, _ = client.AvailableBooks()
				_ = client.APIBaseURL()
				_ = client.HTTPClient()
				_ = client.BurstRate()
			}
		}()
	}

	wg.Wait()
}

func TestClientConfigurationLockNotHeldDuringHTTPDo(t *testing.T) {
	client := NewClient()
	client.SetAPIBaseURL("https://example.test/api")

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})

	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			close(requestStarted)
			<-releaseRequest

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(string(successResponse([]map[string]interface{}{})))),
				Request:    req,
			}, nil
		}),
	})

	errCh := make(chan error, 1)
	go func() {
		_, err := client.AvailableBooks()
		errCh <- err
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for request to reach transport")
	}

	configUpdated := make(chan struct{})
	go func() {
		client.SetAuth("updated-key", "updated-secret")
		client.SetAPIBaseURL("https://sandbox.bitso.com/api")
		client.SetLogLevel(LogLevelTrace)
		client.SetHTTPClient(nil)
		client.SetBurstRate(0)
		close(configUpdated)
	}()

	select {
	case <-configUpdated:
	case <-time.After(time.Second):
		t.Fatal("client configuration mutex was held during HTTP Do")
	}

	close(releaseRequest)

	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for request to finish")
	}
}

func TestAvailableBooks(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/api/v3/available_books", r.URL.Path)

			payload := []map[string]interface{}{
				{
					"book":           "btc_mxn",
					"minimum_amount": "0.001",
					"maximum_amount": "1000.0",
					"minimum_price":  "100.0",
					"maximum_price":  "10000000.0",
					"minimun_value":  "10.0",
					"maximum_value":  "1000000.0",
				},
			}
			w.Write(successResponse(payload))
		})
		defer server.Close()

		books, err := client.AvailableBooks()

		require.NoError(t, err)
		require.Len(t, books, 1)
		assert.Equal(t, "btc_mxn", books[0].Book.String())
	})

	t.Run("api error", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse(101, "Invalid API key"))
		})
		defer server.Close()

		books, err := client.AvailableBooks()

		require.Error(t, err)
		assert.Nil(t, books)
		assert.Contains(t, err.Error(), "Invalid API key")
	})
}

func TestTickers(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/api/v3/ticker", r.URL.Path)

		payload := []map[string]interface{}{
			{
				"book":       "btc_mxn",
				"volume":     "100.5",
				"high":       "500000.00",
				"last":       "480000.00",
				"low":        "450000.00",
				"vwap":       "475000.00",
				"ask":        "480100.00",
				"bid":        "479900.00",
				"created_at": "2024-01-15T10:30:00+00:00",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	tickers, err := client.Tickers()

	require.NoError(t, err)
	require.Len(t, tickers, 1)
	assert.Equal(t, "btc_mxn", tickers[0].Book.String())
	assert.Equal(t, "100.5", string(tickers[0].Volume))
}

func TestTicker(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.RawQuery, "book=btc_mxn")

		payload := map[string]interface{}{
			"book":       "btc_mxn",
			"volume":     "100.5",
			"high":       "500000.00",
			"last":       "480000.00",
			"low":        "450000.00",
			"vwap":       "475000.00",
			"ask":        "480100.00",
			"bid":        "479900.00",
			"created_at": "2024-01-15T10:30:00+00:00",
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	book := NewBook(BTC, MXN)
	ticker, err := client.Ticker(book)

	require.NoError(t, err)
	require.NotNil(t, ticker)
	assert.Equal(t, "btc_mxn", ticker.Book.String())
}

func TestTrades(t *testing.T) {
	t.Run("success with params", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.RawQuery, "book=eth_mxn")

			payload := []map[string]interface{}{
				{
					"book":       "eth_mxn",
					"created_at": "2024-01-15T10:30:00+00:00",
					"amount":     "1.5",
					"maker_side": "buy",
					"price":      "35000.00",
					"tid":        12345,
				},
			}
			w.Write(successResponse(payload))
		})
		defer server.Close()

		params := url.Values{"book": {"eth_mxn"}}
		trades, err := client.Trades(params)

		require.NoError(t, err)
		require.Len(t, trades, 1)
		assert.Equal(t, "eth_mxn", trades[0].Book.String())
		assert.Equal(t, uint64(12345), trades[0].TID.Uint64())
	})

	t.Run("error without book param", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write(errorResponse(303, "The field book is missing"))
		})
		defer server.Close()

		trades, err := client.Trades(nil)

		require.Error(t, err)
		assert.Nil(t, trades)
	})
}

func TestOrderBook(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)

		payload := map[string]interface{}{
			"asks": []map[string]interface{}{
				{"book": "btc_mxn", "price": "500000.00", "amount": "0.5"},
			},
			"bids": []map[string]interface{}{
				{"book": "btc_mxn", "price": "499000.00", "amount": "1.0"},
			},
			"updated_at": "2024-01-15T10:30:00+00:00",
			"sequence":   "12345",
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	params := url.Values{"book": {"btc_mxn"}}
	orderBook, err := client.OrderBook(params)

	require.NoError(t, err)
	require.NotNil(t, orderBook)
	assert.Len(t, orderBook.Asks, 1)
	assert.Len(t, orderBook.Bids, 1)
	assert.Equal(t, "12345", orderBook.Sequence)
}

func TestBalances(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		// Verify auth header is present
		assert.NotEmpty(t, r.Header.Get("Authorization"))

		payload := map[string]interface{}{
			"balances": []map[string]interface{}{
				{
					"currency":           "btc",
					"total":              "1.5",
					"locked":             "0.5",
					"available":          "1.0",
					"pending_deposit":    "0.0",
					"pending_withdrawal": "0.0",
				},
				{
					"currency":           "mxn",
					"total":              "50000.00",
					"locked":             "10000.00",
					"available":          "40000.00",
					"pending_deposit":    "0.0",
					"pending_withdrawal": "0.0",
				},
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	balances, err := client.Balances(nil)

	require.NoError(t, err)
	require.Len(t, balances, 2)
	assert.Equal(t, Currency(BTC), balances[0].Currency)
	assert.Equal(t, "1.5", string(balances[0].Total))
	assert.Equal(t, Currency(MXN), balances[1].Currency)
}

func TestAccountStatus(t *testing.T) {
	const nonce = "1234567890123456789"
	const key = "test-key"
	const secret = "tsecret"

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "AccountStatus should use GET")
		assert.Equal(t, "/api/v3/account_status", r.URL.Path, "AccountStatus should use the documented endpoint")
		assert.Empty(t, r.URL.RawQuery, "AccountStatus should not send query parameters")

		expectedSignature := testSignature(secret, nonce+"GET"+"/api/v3/account_status")
		assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "AccountStatus should sign the request URI")

		payload := map[string]interface{}{
			"account_creation_date":      "2022-01-09T17:07:24+0000",
			"business_name":              "",
			"born_in_residence":          "1",
			"cash_deposit_allowance":     "5300.00",
			"cellphone_number":           "verified",
			"cellphone_number_stored":    "+525555555555",
			"client_id":                  "1234",
			"country_of_residence":       "MX",
			"daily_limit":                "5300.00",
			"daily_remaining":            "3300.00",
			"date_of_birth":              "01/01/1990",
			"email":                      "verified",
			"email_stored":               "claude.shannon@example.com",
			"enabled_two_factor_methods": []string{"totp", "email", "fido"},
			"entity_type":                "N/A",
			"first_name":                 "Claude",
			"gender":                     "F",
			"gravatar_img":               "https://secure.gravatar.com/avatar/example",
			"last_name":                  "Shannon",
			"monthly_limit":              "32000.00",
			"monthly_remaining":          "31000.00",
			"official_id":                "submitted",
			"origin_of_funds":            "unsubmitted",
			"preferred_currency":         "mxn",
			"proof_of_residency":         "submitted",
			"referral_code":              "aauv",
			"second_last_name":           "Monet",
			"signed_contract":            "unsubmitted",
			"status":                     "active",
			"tax_payer_type":             "person",
			"user_default_fiat_currency": "mxn",
			"verification_level":         5,
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) {
		return nonce, nil
	}

	status, err := client.AccountStatus()

	require.NoError(t, err, "get account status")
	require.NotNil(t, status, "account status should decode")
	assert.Equal(t, "1234", status.ClientID, "client_id should decode")
	assert.Equal(t, "Claude", status.FirstName, "first_name should decode")
	assert.Equal(t, "Shannon", status.LastName, "last_name should decode")
	assert.Equal(t, "Monet", status.SecondLastName, "second_last_name should decode")
	assert.Equal(t, "active", status.Status, "status should decode")
	assert.Equal(t, Monetary("5300.00"), status.DailyLimit, "daily_limit should decode")
	assert.Equal(t, Monetary("32000.00"), status.MonthlyLimit, "monthly_limit should decode")
	assert.Equal(t, Monetary("3300.00"), status.DailyRemaining, "daily_remaining should decode")
	assert.Equal(t, Monetary("31000.00"), status.MonthlyRemaining, "monthly_remaining should decode")
	assert.Equal(t, Monetary("5300.00"), status.CashDepositAllowance, "cash_deposit_allowance should decode")
	assert.Equal(t, "verified", status.CellphoneNumber, "cellphone_number should decode")
	assert.Equal(t, "+525555555555", status.CellphoneNumberStored, "cellphone_number_stored should decode")
	assert.Equal(t, "verified", status.Email, "email should decode")
	assert.Equal(t, "claude.shannon@example.com", status.EmailStored, "email_stored should decode")
	assert.Equal(t, "submitted", status.OfficialID, "official_id should decode")
	assert.Equal(t, "submitted", status.ProofOfResidency, "proof_of_residency should decode")
	assert.Equal(t, "unsubmitted", status.SignedContract, "signed_contract should decode")
	assert.Equal(t, "unsubmitted", status.OriginOfFunds, "origin_of_funds should decode")
	assert.Equal(t, 5, status.VerificationLevel, "verification_level should decode as a number")
	assert.Equal(t, "aauv", status.ReferralCode, "deprecated referral_code should decode")
	assert.Equal(t, "MX", status.CountryOfResidence, "country_of_residence should decode")
	assert.Equal(t, "https://secure.gravatar.com/avatar/example", status.GravatarImg, "gravatar_img should decode")
	assert.Equal(t, "2022-01-09T17:07:24+0000", status.AccountCreationDate.String(), "account_creation_date should decode")
	assert.Equal(t, Currency(MXN), status.PreferredCurrency, "preferred_currency should decode")
	assert.Equal(t, []string{"totp", "email", "fido"}, status.EnabledTwoFactorMethods, "enabled_two_factor_methods should decode as an array")
	assert.Equal(t, "", status.BusinessName, "business_name should decode")
	assert.Equal(t, "F", status.Gender, "gender should decode")
	assert.Equal(t, Currency(MXN), status.UserDefaultFiatCurrency, "user_default_fiat_currency should decode")
	assert.Equal(t, "01/01/1990", status.DateOfBirth, "date_of_birth should decode")
	assert.Equal(t, "1", status.BornInResidence, "born_in_residence should decode")
	assert.Equal(t, "person", status.TaxPayerType, "tax_payer_type should decode")
	assert.Equal(t, "N/A", status.EntityType, "entity_type should decode")
}

func TestAccountStatusEmptyOptionalValues(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "AccountStatus should use GET")
		assert.Equal(t, "/api/v3/account_status", r.URL.Path, "AccountStatus should use the documented endpoint")

		payload := map[string]interface{}{
			"account_creation_date":      "2022-01-09T17:07:24+0000",
			"business_name":              "",
			"enabled_two_factor_methods": []string{},
			"entity_type":                "",
			"preferred_currency":         "",
			"referral_code":              "",
			"user_default_fiat_currency": "",
			"verification_level":         0,
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	status, err := client.AccountStatus()

	require.NoError(t, err, "get account status with empty optional values")
	require.NotNil(t, status, "account status should decode")
	assert.Empty(t, status.EnabledTwoFactorMethods, "empty enabled_two_factor_methods should decode")
	assert.Equal(t, CurrencyNone, status.PreferredCurrency, "empty preferred_currency should decode as CurrencyNone")
	assert.Equal(t, CurrencyNone, status.UserDefaultFiatCurrency, "empty user_default_fiat_currency should decode as CurrencyNone")
	assert.Equal(t, "", status.ReferralCode, "empty referral_code should decode")
	assert.Equal(t, 0, status.VerificationLevel, "zero verification_level should decode")
}

func TestRequestCurrencyConversionQuote(t *testing.T) {
	const nonce = "1234567890123456789"
	const key = "test-key"
	const secret = "tsecret"
	const expectedBody = `{"from_currency":"mxn","to_currency":"usd","spend_amount":"1000.00"}`

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "RequestCurrencyConversionQuote should POST")
		assert.Equal(t, "/api/v4/currency_conversions", r.URL.Path, "RequestCurrencyConversionQuote should use the v4 endpoint")
		assert.Empty(t, r.URL.RawQuery, "RequestCurrencyConversionQuote should not send query parameters")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "RequestCurrencyConversionQuote should send JSON")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "read RequestCurrencyConversionQuote body")
		assert.Equal(t, expectedBody, string(body), "RequestCurrencyConversionQuote should send the expected JSON body")

		expectedSignature := testSignature(secret, nonce+"POST"+"/api/v4/currency_conversions"+expectedBody)
		assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "RequestCurrencyConversionQuote should sign the v4 request body")

		payload := map[string]interface{}{
			"id":            "zmAfx2rvNnv1Jn0Q",
			"from_amount":   "1000.00000000",
			"from_currency": "mxn",
			"to_amount":     "53.93125050",
			"to_currency":   "usd",
			"created":       1719956986165,
			"expires":       1719957016165,
			"rate":          "18.54",
			"plain_rate":    "18.36",
			"rate_currency": "mxn",
			"padding":       "0.0098",
			"book":          "xrp_mxn",
			"estimated_slippage": map[string]interface{}{
				"value":   "0.0000",
				"level":   "normal",
				"message": "",
			},
			"next_recurrent_events": map[string]interface{}{
				"DAILY":   1720043386165,
				"WEEKLY":  1720561786165,
				"MONTHLY": 1722635386165,
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) {
		return nonce, nil
	}

	quote, err := client.RequestCurrencyConversionQuote(&CurrencyConversionQuoteRequest{
		FromCurrency: Currency(MXN),
		ToCurrency:   Currency(USD),
		SpendAmount:  Monetary("1000.00"),
	})

	require.NoError(t, err, "request currency conversion quote")
	require.NotNil(t, quote, "currency conversion quote should decode")
	assert.Equal(t, "zmAfx2rvNnv1Jn0Q", quote.ID, "id should decode")
	assert.Equal(t, Monetary("1000.00000000"), quote.FromAmount, "from_amount should decode")
	assert.Equal(t, Currency(MXN), quote.FromCurrency, "from_currency should decode")
	assert.Equal(t, Monetary("53.93125050"), quote.ToAmount, "to_amount should decode")
	assert.Equal(t, Currency(USD), quote.ToCurrency, "to_currency should decode")
	assert.Equal(t, Monetary("0.0000"), quote.EstimatedSlippage.Value, "estimated_slippage.value should decode")
	assert.Equal(t, "normal", quote.EstimatedSlippage.Level, "estimated_slippage.level should decode")
	assert.Equal(t, "", quote.EstimatedSlippage.Message, "estimated_slippage.message should decode")
	assert.Equal(t, int64(1719956986165), quote.Created, "created should decode")
	assert.Equal(t, int64(1719957016165), quote.Expires, "expires should decode")
	assert.Equal(t, Monetary("18.54"), quote.Rate, "rate should decode")
	assert.Equal(t, Monetary("18.36"), quote.PlainRate, "plain_rate should decode")
	assert.Equal(t, Currency(MXN), quote.RateCurrency, "rate_currency should decode")
	assert.Equal(t, Monetary("0.0098"), quote.Padding, "padding should decode")
	assert.Equal(t, "xrp_mxn", quote.Book.String(), "book should decode")
	assert.Equal(t, int64(1720043386165), quote.NextRecurrentEvents["DAILY"], "next_recurrent_events should decode")
	assert.Equal(t, int64(1720561786165), quote.NextRecurrentEvents["WEEKLY"], "next_recurrent_events should decode")
	assert.Equal(t, int64(1722635386165), quote.NextRecurrentEvents["MONTHLY"], "next_recurrent_events should decode")
}

func TestRequestCurrencyConversionQuoteWithReceiveAmount(t *testing.T) {
	const expectedBody = `{"from_currency":"mxn","to_currency":"usd","receive_amount":"100.00"}`

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "RequestCurrencyConversionQuote should POST")
		assert.Equal(t, "/api/v4/currency_conversions", r.URL.Path, "RequestCurrencyConversionQuote should use the v4 endpoint")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "read RequestCurrencyConversionQuote body")
		assert.Equal(t, expectedBody, string(body), "RequestCurrencyConversionQuote should send receive_amount when selected")

		w.Write(successResponse(map[string]interface{}{
			"id":            "receive-quote",
			"from_amount":   "1854.21860516",
			"from_currency": "mxn",
			"to_amount":     "100.00000000",
			"to_currency":   "usd",
			"created":       1719862339217,
			"expires":       1719862369217,
			"rate":          "18.54",
			"plain_rate":    "18.36",
			"rate_currency": "mxn",
			"padding":       "0.0098",
			"book":          "xrp_mxn",
			"estimated_slippage": map[string]interface{}{
				"value":   "0.0000",
				"level":   "normal",
				"message": "",
			},
			"next_recurrent_events": map[string]interface{}{},
		}))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	quote, err := client.RequestCurrencyConversionQuote(&CurrencyConversionQuoteRequest{
		FromCurrency:  Currency(MXN),
		ToCurrency:    Currency(USD),
		ReceiveAmount: Monetary("100.00"),
	})

	require.NoError(t, err, "request currency conversion quote with receive_amount")
	require.NotNil(t, quote, "currency conversion quote should decode")
	assert.Equal(t, "receive-quote", quote.ID, "id should decode")
}

func TestExecuteCurrencyConversion(t *testing.T) {
	const nonce = "1234567890123456789"
	const key = "test-key"
	const secret = "tsecret"

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method, "ExecuteCurrencyConversion should PUT")
		assert.Equal(t, "/api/v4/currency_conversions/quote-123", r.URL.Path, "ExecuteCurrencyConversion should use the v4 quote endpoint")
		assert.Empty(t, r.URL.RawQuery, "ExecuteCurrencyConversion should not send query parameters")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "read ExecuteCurrencyConversion body")
		assert.Empty(t, body, "ExecuteCurrencyConversion should not send a body")

		expectedSignature := testSignature(secret, nonce+"PUT"+"/api/v4/currency_conversions/quote-123")
		assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "ExecuteCurrencyConversion should sign the v4 request URI")

		w.Write(successResponse(map[string]interface{}{
			"oid": "7316",
		}))
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) {
		return nonce, nil
	}

	oid, err := client.ExecuteCurrencyConversion("quote-123")

	require.NoError(t, err, "execute currency conversion")
	assert.Equal(t, "7316", oid, "ExecuteCurrencyConversion should return the conversion id")
}

func TestCurrencyConversionStatus(t *testing.T) {
	const nonce = "1234567890123456789"
	const key = "test-key"
	const secret = "tsecret"

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "CurrencyConversionStatus should GET")
		assert.Equal(t, "/api/v4/currency_conversions/7316", r.URL.Path, "CurrencyConversionStatus should use the v4 conversion endpoint")
		assert.Empty(t, r.URL.RawQuery, "CurrencyConversionStatus should not send query parameters")

		expectedSignature := testSignature(secret, nonce+"GET"+"/api/v4/currency_conversions/7316")
		assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "CurrencyConversionStatus should sign the v4 request URI")

		payload := map[string]interface{}{
			"id":            "7316",
			"from_amount":   "1854.21860516",
			"from_currency": "mxn",
			"to_amount":     "100.00000000",
			"to_currency":   "usd",
			"created":       1719862355209,
			"expires":       1719862385209,
			"rate":          "18.54218605",
			"plain_rate":    "18.36",
			"rate_currency": "mxn",
			"book":          "xrp_mxn",
			"status":        "queued",
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) {
		return nonce, nil
	}

	conversion, err := client.CurrencyConversionStatus("7316")

	require.NoError(t, err, "get currency conversion status")
	require.NotNil(t, conversion, "currency conversion should decode")
	assert.Equal(t, "7316", conversion.ID, "id should decode")
	assert.Equal(t, Monetary("1854.21860516"), conversion.FromAmount, "from_amount should decode")
	assert.Equal(t, Currency(MXN), conversion.FromCurrency, "from_currency should decode")
	assert.Equal(t, Monetary("100.00000000"), conversion.ToAmount, "to_amount should decode")
	assert.Equal(t, Currency(USD), conversion.ToCurrency, "to_currency should decode")
	assert.Equal(t, int64(1719862355209), conversion.Created, "created should decode")
	assert.Equal(t, int64(1719862385209), conversion.Expires, "expires should decode")
	assert.Equal(t, Monetary("18.54218605"), conversion.Rate, "rate should decode")
	assert.Equal(t, Monetary("18.36"), conversion.PlainRate, "plain_rate should decode")
	assert.Equal(t, Currency(MXN), conversion.RateCurrency, "rate_currency should decode")
	assert.Equal(t, "xrp_mxn", conversion.Book.String(), "book should decode")
	assert.Equal(t, CurrencyConversionStatusQueued, conversion.Status, "status should decode")
}

func TestCurrencyConversionRoutesUseV4WithCustomAPIBaseURL(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/custom/api/v3/available_books":
			assert.Equal(t, "GET", r.Method, "AvailableBooks should keep using GET")
			w.Write(successResponse([]map[string]interface{}{}))
		case "/custom/api/v4/currency_conversions":
			assert.Equal(t, "POST", r.Method, "RequestCurrencyConversionQuote should use POST")
			w.Write(successResponse(map[string]interface{}{
				"id":            "quote-custom-base",
				"from_amount":   "1000.00000000",
				"from_currency": "mxn",
				"to_amount":     "53.93125050",
				"to_currency":   "usd",
				"created":       1719956986165,
				"expires":       1719957016165,
				"rate":          "18.54",
				"plain_rate":    "18.36",
				"rate_currency": "mxn",
				"padding":       "0.0098",
				"book":          "xrp_mxn",
				"estimated_slippage": map[string]interface{}{
					"value":   "0.0000",
					"level":   "normal",
					"message": "",
				},
				"next_recurrent_events": map[string]interface{}{},
			}))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.SetAPIBaseURL(server.URL + "/custom/api")
	client.SetAuth("test-key", "tsecret")

	books, err := client.AvailableBooks()
	require.NoError(t, err, "available books should still use v3")
	assert.Empty(t, books, "available books response should decode")

	quote, err := client.RequestCurrencyConversionQuote(&CurrencyConversionQuoteRequest{
		FromCurrency: Currency(MXN),
		ToCurrency:   Currency(USD),
		SpendAmount:  Monetary("1000.00"),
	})
	require.NoError(t, err, "request currency conversion quote should use v4")
	assert.Equal(t, "quote-custom-base", quote.ID, "quote should decode")
	assert.Equal(t, []string{"/custom/api/v3/available_books", "/custom/api/v4/currency_conversions"}, paths, "custom API base URL should preserve v3 and v4 routes")
}

func TestCurrencyConversionQuoteValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *CurrencyConversionQuoteRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "request is required",
		},
		{
			name: "missing from currency",
			request: &CurrencyConversionQuoteRequest{
				ToCurrency:  Currency(USD),
				SpendAmount: Monetary("100.00"),
			},
			wantErr: "from_currency is required",
		},
		{
			name: "missing to currency",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency: Currency(MXN),
				SpendAmount:  Monetary("100.00"),
			},
			wantErr: "to_currency is required",
		},
		{
			name: "missing amount",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency: Currency(MXN),
				ToCurrency:   Currency(USD),
			},
			wantErr: "exactly one of spend_amount or receive_amount is required",
		},
		{
			name: "both amounts",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency:  Currency(MXN),
				ToCurrency:    Currency(USD),
				SpendAmount:   Monetary("100.00"),
				ReceiveAmount: Monetary("10.00"),
			},
			wantErr: "exactly one of spend_amount or receive_amount is required",
		},
		{
			name: "invalid spend amount",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency: Currency(MXN),
				ToCurrency:   Currency(USD),
				SpendAmount:  Monetary("not-a-number"),
			},
			wantErr: "spend_amount must be a valid decimal",
		},
		{
			name: "zero spend amount",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency: Currency(MXN),
				ToCurrency:   Currency(USD),
				SpendAmount:  Monetary("0"),
			},
			wantErr: "spend_amount must be positive",
		},
		{
			name: "negative receive amount",
			request: &CurrencyConversionQuoteRequest{
				FromCurrency:  Currency(MXN),
				ToCurrency:    Currency(USD),
				ReceiveAmount: Monetary("-1"),
			},
			wantErr: "receive_amount must be positive",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient()
			client.SetHTTPClient(&http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					t.Fatalf("validation error should happen before sending request to %s", req.URL.String())
					return nil, nil
				}),
			})

			quote, err := client.RequestCurrencyConversionQuote(tc.request)

			require.Error(t, err, "invalid quote request should fail")
			assert.Contains(t, err.Error(), tc.wantErr, "validation error should explain the invalid field")
			assert.Nil(t, quote, "invalid quote request should not return a quote")
		})
	}
}

func TestCurrencyConversionIDValidation(t *testing.T) {
	client := NewClient()
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("validation error should happen before sending request to %s", req.URL.String())
			return nil, nil
		}),
	})

	oid, err := client.ExecuteCurrencyConversion(" ")
	require.Error(t, err, "empty quote id should fail")
	assert.Contains(t, err.Error(), "quote_id is required", "quote_id validation should explain the invalid field")
	assert.Empty(t, oid, "invalid quote id should not return a conversion id")

	conversion, err := client.CurrencyConversionStatus("")
	require.Error(t, err, "empty conversion id should fail")
	assert.Contains(t, err.Error(), "conversion_id is required", "conversion_id validation should explain the invalid field")
	assert.Nil(t, conversion, "invalid conversion id should not return a conversion")
}

func TestFees(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]interface{}{
			"fees": []map[string]interface{}{
				{
					"book":        "btc_mxn",
					"fee_decimal": "0.0065",
					"fee_percent": "0.65",
				},
			},
			"withdrawal_fees": map[string]interface{}{
				"btc": "0.0001",
				"eth": "0.005",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	fees, err := client.Fees(nil)

	require.NoError(t, err)
	require.NotNil(t, fees)
	assert.Len(t, fees.Fees, 1)
	assert.Equal(t, "btc_mxn", fees.Fees[0].Book.String())
	assert.Contains(t, fees.WithdrawalFees, "btc")
}

func TestLedger(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/ledger", r.URL.Path)

		payload := []map[string]interface{}{
			{
				"eid":        "abc123",
				"operation":  "trade",
				"created_at": "2024-01-15T10:30:00+00:00",
				"balance_updates": []map[string]interface{}{
					{"currency": "btc", "amount": "-0.1"},
					{"currency": "mxn", "amount": "50000.00"},
				},
				"details": map[string]interface{}{
					"tid": "12345",
				},
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	transactions, err := client.Ledger(nil)

	require.NoError(t, err)
	require.Len(t, transactions, 1)
	assert.Equal(t, "abc123", transactions[0].EID)
	assert.Equal(t, OperationTrade, transactions[0].Operation)
	assert.Len(t, transactions[0].BalanceUpdates, 2)
}

func TestLedgerByOperation(t *testing.T) {
	testCases := []struct {
		operation    Operation
		expectedPath string
	}{
		{OperationFunding, "/ledger/fundings"},
		{OperationWithdrawal, "/ledger/withdrawals"},
		{OperationTrade, "/ledger/trades"},
		{OperationFee, "/ledger/fees"},
	}

	for _, tc := range testCases {
		t.Run(tc.operation.String(), func(t *testing.T) {
			server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/v3"+tc.expectedPath, r.URL.Path)
				w.Write(successResponse([]map[string]interface{}{}))
			})
			defer server.Close()

			client.SetAuth("test-key", "tsecret")
			_, err := client.LedgerByOperation(tc.operation, nil)

			require.NoError(t, err)
		})
	}
}

func TestFundings(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		payload := []map[string]interface{}{
			{
				"fid":        "fund123",
				"currency":   "btc",
				"method":     "Bitcoin Network",
				"amount":     "0.5",
				"status":     "complete",
				"created_at": "2024-01-15T10:30:00+00:00",
				"details":    map[string]interface{}{"txid": "abc123"},
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	fundings, err := client.Fundings(nil)

	require.NoError(t, err)
	require.Len(t, fundings, 1)
	assert.Equal(t, "fund123", fundings[0].FID)
	assert.Equal(t, Currency(BTC), fundings[0].Currency)
}

func TestWithdrawals(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		payload := []map[string]interface{}{
			{
				"wid":        "withdraw123",
				"currency":   "mxn",
				"method":     "SPEI",
				"amount":     "10000.00",
				"status":     "complete",
				"created_at": "2024-01-15T10:30:00+00:00",
				"details":    map[string]interface{}{"clabe": "1234567890"},
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	withdrawals, err := client.Withdrawals(nil)

	require.NoError(t, err)
	require.Len(t, withdrawals, 1)
	assert.Equal(t, "withdraw123", withdrawals[0].WID)
	assert.Equal(t, Currency(MXN), withdrawals[0].Currency)
}

func TestMyTrades(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		payload := []map[string]interface{}{
			{
				"book":          "btc_mxn",
				"major":         "-0.1",
				"created_at":    "2024-01-15T10:30:00+00:00",
				"minor":         "50000.00",
				"fees_amount":   "325.00",
				"fees_currency": "mxn",
				"price":         "500000.00",
				"tid":           12345,
				"oid":           "order123",
				"side":          "sell",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	trades, err := client.MyTrades(nil)

	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.Equal(t, "btc_mxn", trades[0].Book.String())
	assert.Equal(t, "order123", trades[0].OID)
	assert.Equal(t, OrderSideSell, trades[0].Side)
}

func TestOrderTrades(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/order_trades/order123", r.URL.Path)

		payload := []map[string]interface{}{
			{
				"book":        "btc_mxn",
				"major":       "-0.05",
				"created_at":  "2024-01-15T10:30:00+00:00",
				"minor":       "25000.00",
				"fees_amount": "162.50",
				"currency":    "mxn",
				"price":       "500000.00",
				"tid":         12346,
				"oid":         "order123",
				"side":        "sell",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	trades, err := client.OrderTrades("order123", nil)

	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.Equal(t, "order123", trades[0].OID)
}

func TestMyOpenOrders(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		payload := []map[string]interface{}{
			{
				"book":            "btc_mxn",
				"original_amount": "0.1",
				"unfilled_amount": "0.05",
				"original_value":  "50000.00",
				"created_at":      "2024-01-15T10:30:00+00:00",
				"updated_at":      "2024-01-15T10:35:00+00:00",
				"price":           "500000.00",
				"oid":             "order123",
				"side":            "sell",
				"status":          "partially filled",
				"type":            "limit",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	orders, err := client.MyOpenOrders(nil)

	require.NoError(t, err)
	require.Len(t, orders, 1)
	assert.Equal(t, "order123", orders[0].OID)
	assert.Equal(t, OrderStatusPartialFill, orders[0].Status)
	assert.Equal(t, OrderSideSell, orders[0].Side)
}

func TestLookupOrder(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v3/orders/order123", r.URL.Path)

			payload := []map[string]interface{}{
				{
					"book":            "btc_mxn",
					"original_amount": "0.1",
					"unfilled_amount": "0.0",
					"original_value":  "50000.00",
					"created_at":      "2024-01-15T10:30:00+00:00",
					"updated_at":      "2024-01-15T10:35:00+00:00",
					"price":           "500000.00",
					"oid":             "order123",
					"side":            "buy",
					"status":          "completed",
					"type":            "limit",
				},
			}
			w.Write(successResponse(payload))
		})
		defer server.Close()

		client.SetAuth("test-key", "tsecret")
		order, err := client.LookupOrder("order123")

		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, "order123", order.OID)
		assert.Equal(t, OrderStatusCompleted, order.Status)
	})

	t.Run("not found", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write(successResponse([]map[string]interface{}{}))
		})
		defer server.Close()

		client.SetAuth("test-key", "tsecret")
		order, err := client.LookupOrder("nonexistent")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "no such order")
	})
}

func TestLookupOrders(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/orders/order1,order2", r.URL.Path)

		payload := []map[string]interface{}{
			{"oid": "order1", "book": "btc_mxn", "status": "completed", "side": "buy",
				"original_amount": "0.1", "unfilled_amount": "0.0", "price": "500000.00",
				"created_at": "2024-01-15T10:30:00+00:00", "updated_at": "2024-01-15T10:30:00+00:00"},
			{"oid": "order2", "book": "eth_mxn", "status": "open", "side": "sell",
				"original_amount": "1.0", "unfilled_amount": "1.0", "price": "35000.00",
				"created_at": "2024-01-15T10:30:00+00:00", "updated_at": "2024-01-15T10:30:00+00:00"},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	orders, err := client.LookupOrders([]string{"order1", "order2"})

	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "order1", orders[0].OID)
	assert.Equal(t, "order2", orders[1].OID)
}

func TestPlaceOrder(t *testing.T) {
	major, err := NewMonetary("0.10000000")
	require.NoError(t, err, "create major amount")
	price, err := NewMonetary("500000.00")
	require.NoError(t, err, "create limit price")
	stop, err := NewMonetary("490000.00")
	require.NoError(t, err, "create stop price")
	zeroSlippage := 0.0
	nonZeroSlippage := 0.5

	tests := []struct {
		name         string
		order        *OrderPlacement
		expectedBody string
		wantSlippage *float64
	}{
		{
			name: "omits optional fields",
			order: &OrderPlacement{
				Book:  *NewBook(BTC, MXN),
				Side:  OrderSideBuy,
				Type:  OrderTypeLimit,
				Major: major,
				Price: price,
			},
			expectedBody: `{
				"book": "btc_mxn",
				"side": "buy",
				"type": "limit",
				"major": "0.10000000",
				"price": "500000.00"
			}`,
		},
		{
			name: "sends zero slippage as number",
			order: &OrderPlacement{
				Book:              *NewBook(BTC, MXN),
				Side:              OrderSideBuy,
				Type:              OrderTypeMarket,
				SlippageTolerance: &zeroSlippage,
			},
			expectedBody: `{
				"book": "btc_mxn",
				"side": "buy",
				"type": "market",
				"slippage_tolerance": 0
			}`,
			wantSlippage: &zeroSlippage,
		},
		{
			name: "sends optional order fields",
			order: &OrderPlacement{
				Book:              *NewBook(BTC, MXN),
				Side:              OrderSideBuy,
				Type:              OrderTypeLimit,
				Major:             major,
				Price:             price,
				OriginID:          "client-order-123",
				Stop:              stop,
				TimeInForce:       OrderTimeInForcePostOnly,
				SlippageTolerance: &nonZeroSlippage,
				MarginOrderType:   MarginOrderTypeCrossMargin,
			},
			expectedBody: `{
				"book": "btc_mxn",
				"side": "buy",
				"type": "limit",
				"major": "0.10000000",
				"price": "500000.00",
				"origin_id": "client-order-123",
				"stop": "490000.00",
				"time_in_force": "postonly",
				"slippage_tolerance": 0.5,
				"margin_order_type": "CROSS_MARGIN"
			}`,
			wantSlippage: &nonZeroSlippage,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method, "PlaceOrder should POST to /orders")
				assert.Equal(t, "/api/v3/orders/", r.URL.Path, "PlaceOrder should use the order placement endpoint")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "PlaceOrder should send JSON")

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err, "read PlaceOrder request body")
				assert.JSONEq(t, tc.expectedBody, string(body), "PlaceOrder should send the expected order placement request")

				var payload map[string]interface{}
				require.NoError(t, json.Unmarshal(body, &payload), "decode PlaceOrder request body")
				if tc.wantSlippage == nil {
					assert.NotContains(t, payload, "slippage_tolerance", "nil slippage_tolerance should be omitted")
				} else {
					assert.Equal(t, *tc.wantSlippage, payload["slippage_tolerance"], "slippage_tolerance should be a JSON number")
				}

				w.Write(successResponse(map[string]interface{}{
					"oid": "new-order-123",
				}))
			})
			defer server.Close()

			client.SetAuth("test-key", "tsecret")
			oid, err := client.PlaceOrder(tc.order)

			require.NoError(t, err, "place order")
			assert.Equal(t, "new-order-123", oid, "PlaceOrder should return the new order id")
		})
	}
}

func TestModifyOrder(t *testing.T) {
	const nonce = "173134920012345678"
	const key = "test-key"
	const secret = "tsecret"

	major, err := NewMonetary("0.10000000")
	require.NoError(t, err, "create major amount")
	price, err := NewMonetary("500000.00")
	require.NoError(t, err, "create price")
	stop, err := NewMonetary("490000.00")
	require.NoError(t, err, "create stop")

	tests := []struct {
		name          string
		call          func(*Client) (string, error)
		expectedPath  string
		expectedQuery string
		expectedBody  string
	}{
		{
			name: "path oid",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("order 123", &OrderModification{
					Major:  major,
					Price:  price,
					Cancel: true,
				})
			},
			expectedPath: "/api/v4/orders/order%20123",
			expectedBody: `{"major":"0.10000000","price":"500000.00","cancel":true}`,
		},
		{
			name: "query oid",
			call: func(client *Client) (string, error) {
				return client.ModifyOrderByQueryOID("order 456", &OrderModification{
					Price: price,
				})
			},
			expectedPath:  "/api/v4/orders",
			expectedQuery: "oid=order+456",
			expectedBody:  `{"price":"500000.00"}`,
		},
		{
			name: "query origin id",
			call: func(client *Client) (string, error) {
				return client.ModifyOrderByOriginID("client/order 789", &OrderModification{
					Stop: stop,
				})
			},
			expectedPath:  "/api/v4/orders",
			expectedQuery: "origin_id=client%2Forder+789",
			expectedBody:  `{"stop":"490000.00"}`,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PATCH", r.Method, "ModifyOrder should PATCH")
				assert.Equal(t, tc.expectedPath, r.URL.EscapedPath(), "ModifyOrder should use the expected v4 path")
				assert.Equal(t, tc.expectedQuery, r.URL.RawQuery, "ModifyOrder should use the expected query")
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "ModifyOrder should send JSON")

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err, "read ModifyOrder request body")
				assert.Equal(t, tc.expectedBody, string(body), "ModifyOrder should send the expected JSON body")

				message := nonce + r.Method + r.URL.RequestURI() + tc.expectedBody
				expectedSignature := testSignature(secret, message)
				assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "ModifyOrder should sign the exact PATCH body")

				w.Write(successResponse(map[string]interface{}{
					"oid": "modified-order-123",
				}))
			})
			defer server.Close()

			client.SetAuth(key, secret)
			client.nonce = func() (string, error) {
				return nonce, nil
			}

			oid, err := tc.call(client)

			require.NoError(t, err, "modify order")
			assert.Equal(t, "modified-order-123", oid, "ModifyOrder should return the modified order id")
		})
	}
}

func TestModifyOrderValidation(t *testing.T) {
	major, err := NewMonetary("0.10000000")
	require.NoError(t, err, "create major amount")
	minor, err := NewMonetary("50000.00")
	require.NoError(t, err, "create minor amount")
	price, err := NewMonetary("500000.00")
	require.NoError(t, err, "create price")

	tests := []struct {
		name    string
		call    func(*Client) (string, error)
		wantErr string
	}{
		{
			name: "path oid is required",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("", &OrderModification{Price: price})
			},
			wantErr: "oid is required",
		},
		{
			name: "query oid is required",
			call: func(client *Client) (string, error) {
				return client.ModifyOrderByQueryOID(" ", &OrderModification{Price: price})
			},
			wantErr: "oid is required",
		},
		{
			name: "origin id is required",
			call: func(client *Client) (string, error) {
				return client.ModifyOrderByOriginID("", &OrderModification{Price: price})
			},
			wantErr: "origin_id is required",
		},
		{
			name: "modification is required",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("order123", nil)
			},
			wantErr: "order modification is required",
		},
		{
			name: "empty modification is rejected",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("order123", &OrderModification{})
			},
			wantErr: "order modification requires at least one of major, minor, price, or stop",
		},
		{
			name: "cancel alone is rejected",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("order123", &OrderModification{Cancel: true})
			},
			wantErr: "cancel cannot be sent alone",
		},
		{
			name: "major and minor are mutually exclusive",
			call: func(client *Client) (string, error) {
				return client.ModifyOrder("order123", &OrderModification{Major: major, Minor: minor})
			},
			wantErr: "major and minor cannot both be set",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
				t.Fatalf("ModifyOrder validation should fail before sending %s %s", r.Method, r.URL.RequestURI())
			})
			defer server.Close()

			oid, err := tc.call(client)

			require.Error(t, err, "invalid ModifyOrder input should return an error")
			assert.Empty(t, oid, "invalid ModifyOrder input should not return an oid")
			assert.Contains(t, err.Error(), tc.wantErr, "ModifyOrder validation error should explain the invalid input")
		})
	}
}

func TestModifyOrderLeavesV3OrderRoutesUnchanged(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "LookupOrder should still use GET")
		assert.Equal(t, "/api/v3/orders/order123", r.URL.Path, "LookupOrder should keep the existing v3 route")

		w.Write(successResponse([]map[string]interface{}{
			{"oid": "order123"},
		}))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	order, err := client.LookupOrder("order123")

	require.NoError(t, err, "lookup order through existing v3 route")
	require.NotNil(t, order, "LookupOrder should decode the existing v3 response")
	assert.Equal(t, "order123", order.OID, "LookupOrder should return the v3 order id")
}

func TestCancelOrder(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v3/orders/order123", r.URL.Path)

		w.Write(successResponse([]string{"order123"}))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	cancelled, err := client.CancelOrder("order123")

	require.NoError(t, err)
	require.Len(t, cancelled, 1)
	assert.Equal(t, "order123", cancelled[0])
}

func TestCancelOrders(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/api/v3/orders/order1,order2", r.URL.Path)

		w.Write(successResponse([]string{"order1", "order2"}))
	})
	defer server.Close()

	client.SetAuth("test-key", "tsecret")
	cancelled, err := client.CancelOrders([]string{"order1", "order2"})

	require.NoError(t, err)
	require.Len(t, cancelled, 2)
}

func TestAPIError(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write(errorResponse(101, "Invalid API key"))
	})
	defer server.Close()

	_, err := client.AvailableBooks()

	require.Error(t, err)

	var apiErr *Error
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 101, apiErr.Code())
	assert.Contains(t, apiErr.Error(), "Invalid API key")
}

func TestAuthorizationHeader(t *testing.T) {
	t.Run("signs get request with deterministic nonce", func(t *testing.T) {
		const nonce = "173134920012345678"

		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			message := nonce + r.Method + r.URL.RequestURI()
			wantAuth := "Bitso test-key:" + nonce + ":" + testSignature("tsecret", message)
			assert.Equal(t, wantAuth, r.Header.Get("Authorization"))

			w.Write(successResponse(map[string]interface{}{"balances": []interface{}{}}))
		})
		defer server.Close()

		client.SetAuth("test-key", "tsecret")
		client.nonce = func() (string, error) { return nonce, nil }

		_, err := client.Balances(nil)

		require.NoError(t, err)
	})

	t.Run("includes supplied body in signature", func(t *testing.T) {
		const nonce = "173134920012399999"
		const body = `{"probe":true}`

		logger := zerolog.Nop()
		cfg := clientConfig{
			key:       "test-key",
			apiSecret: "tsecret",
			nonce:     func() (string, error) { return nonce, nil },
		}

		req, err := cfg.newRequest(&logger, "GET", "https://example.test/api/v3/balance", strings.NewReader(body))

		require.NoError(t, err)
		message := nonce + req.Method + req.URL.RequestURI() + body
		wantAuth := "Bitso test-key:" + nonce + ":" + testSignature("tsecret", message)
		assert.Equal(t, wantAuth, req.Header.Get("Authorization"))
	})
}

func TestGenerateNonceV2(t *testing.T) {
	before := time.Now().UnixMilli()
	nonce, err := generateNonceV2()
	after := time.Now().UnixMilli()

	require.NoError(t, err)
	require.Len(t, nonce, 19)
	for _, r := range nonce {
		assert.True(t, r >= '0' && r <= '9', "nonce %q contains non-digit %q", nonce, r)
	}

	timestamp, err := strconv.ParseInt(nonce[:13], 10, 64)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, timestamp, before)
	assert.LessOrEqual(t, timestamp, after)

	salt, err := strconv.Atoi(nonce[13:])
	require.NoError(t, err)
	assert.GreaterOrEqual(t, salt, nonceV2SaltMin)
	assert.LessOrEqual(t, salt, nonceV2SaltMax)
}

func TestTraceLogsRedactAuthMaterial(t *testing.T) {
	var logBuf bytes.Buffer
	var authHeader string
	var nonce string
	var signature string
	var signedMessage string

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		require.NotEmpty(t, authHeader)

		parts := strings.Split(strings.TrimPrefix(authHeader, "Bitso "), ":")
		require.Len(t, parts, 3)
		nonce = parts[1]
		signature = parts[2]
		signedMessage = nonce + r.Method + r.URL.RequestURI()

		w.Write(errorResponse(101, "Invalid API key"))
	})
	defer server.Close()

	client.SetAuth("sensitive-api-key", "sensitive-api-secret")
	client.logger = client.logger.Output(&logBuf).Level(LogLevelTrace)

	_, err := client.Balances(nil)
	require.Error(t, err)

	logOutput := logBuf.String()
	require.NotEmpty(t, logOutput)
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"endpoint":"/balance"`)
	assert.Contains(t, logOutput, `"auth.signed":true`)

	assert.NotContains(t, logOutput, authHeader)
	assert.NotContains(t, logOutput, "sensitive-api-key")
	assert.NotContains(t, logOutput, "sensitive-api-secret")
	assert.NotContains(t, logOutput, nonce)
	assert.NotContains(t, logOutput, signature)
	assert.NotContains(t, logOutput, signedMessage)
	assert.NotContains(t, logOutput, `"auth.nonce"`)
	assert.NotContains(t, logOutput, `"signature"`)
}

func TestUnauthenticatedRequest(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Public endpoint should work without auth
		assert.Empty(t, r.Header.Get("Authorization"))
		w.Write(successResponse([]interface{}{}))
	})
	defer server.Close()

	// Don't set auth
	books, err := client.AvailableBooks()

	require.NoError(t, err)
	assert.NotNil(t, books)
}

func TestInvalidJSONResponse(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	})
	defer server.Close()

	_, err := client.AvailableBooks()

	require.Error(t, err)
}

func TestEndpointURL(t *testing.T) {
	c := NewClient()
	c.SetAPIBaseURL("https://api.bitso.com/api")

	u, err := c.endpointURL("/balance")

	require.NoError(t, err)
	assert.Equal(t, "https://api.bitso.com/api/v3/balance", u.String(), "default endpointURL should use v3")

	u, err = c.endpointURLForRoute(apiRouteV4, "/currency_conversions")

	require.NoError(t, err)
	assert.Equal(t, "https://api.bitso.com/api/v4/currency_conversions", u.String(), "v4 route should share the configured API base URL")
}
