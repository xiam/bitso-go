package bitso

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestNewClient(t *testing.T) {
	c := NewClient()

	require.NotNil(t, c, "NewClient should return non-nil client")
	assert.NotNil(t, c.client, "HTTP client should be initialized")
	assert.NotNil(t, c.tickets, "Tickets channel should be initialized")
	assert.Equal(t, "https://bitso.com/api/", c.baseURL, "Default base URL should be set")
	assert.Equal(t, "v3", c.version, "Default version should be v3")
	assert.Equal(t, time.Duration(0), c.burstRate, "Default burst rate should be 0")
}

func TestSetAuth(t *testing.T) {
	c := NewClient()

	c.SetAuth("test-key", "test-secret")

	assert.Equal(t, "test-key", c.key)
	assert.Equal(t, "test-secret", c.secret)
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

func TestAvailableBooks(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.True(t, strings.HasSuffix(r.URL.Path, "/available_books"))

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
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ticker"))

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

	client.SetAuth("test-key", "test-secret")
	balances, err := client.Balances(nil)

	require.NoError(t, err)
	require.Len(t, balances, 2)
	assert.Equal(t, Currency(BTC), balances[0].Currency)
	assert.Equal(t, "1.5", string(balances[0].Total))
	assert.Equal(t, Currency(MXN), balances[1].Currency)
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

	client.SetAuth("test-key", "test-secret")
	fees, err := client.Fees(nil)

	require.NoError(t, err)
	require.NotNil(t, fees)
	assert.Len(t, fees.Fees, 1)
	assert.Equal(t, "btc_mxn", fees.Fees[0].Book.String())
	assert.Contains(t, fees.WithdrawalFees, "btc")
}

func TestLedger(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.HasSuffix(r.URL.Path, "/ledger"))

		payload := []map[string]interface{}{
			{
				"eid":       "abc123",
				"operation": "trade",
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

	client.SetAuth("test-key", "test-secret")
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
				assert.True(t, strings.HasSuffix(r.URL.Path, tc.expectedPath))
				w.Write(successResponse([]map[string]interface{}{}))
			})
			defer server.Close()

			client.SetAuth("test-key", "test-secret")
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

	client.SetAuth("test-key", "test-secret")
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

	client.SetAuth("test-key", "test-secret")
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

	client.SetAuth("test-key", "test-secret")
	trades, err := client.MyTrades(nil)

	require.NoError(t, err)
	require.Len(t, trades, 1)
	assert.Equal(t, "btc_mxn", trades[0].Book.String())
	assert.Equal(t, "order123", trades[0].OID)
	assert.Equal(t, OrderSideSell, trades[0].Side)
}

func TestOrderTrades(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.Contains(r.URL.Path, "/order_trades/order123"))

		payload := []map[string]interface{}{
			{
				"book":          "btc_mxn",
				"major":         "-0.05",
				"created_at":    "2024-01-15T10:30:00+00:00",
				"minor":         "25000.00",
				"fees_amount":   "162.50",
				"currency":      "mxn",
				"price":         "500000.00",
				"tid":           12346,
				"oid":           "order123",
				"side":          "sell",
			},
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "test-secret")
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

	client.SetAuth("test-key", "test-secret")
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
			assert.True(t, strings.Contains(r.URL.Path, "/orders/order123"))

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

		client.SetAuth("test-key", "test-secret")
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

		client.SetAuth("test-key", "test-secret")
		order, err := client.LookupOrder("nonexistent")

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "no such order")
	})
}

func TestLookupOrders(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.True(t, strings.Contains(r.URL.Path, "/orders/order1,order2"))

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

	client.SetAuth("test-key", "test-secret")
	orders, err := client.LookupOrders([]string{"order1", "order2"})

	require.NoError(t, err)
	require.Len(t, orders, 2)
	assert.Equal(t, "order1", orders[0].OID)
	assert.Equal(t, "order2", orders[1].OID)
}

func TestPlaceOrder(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body OrderPlacement
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "btc_mxn", body.Book.String())
		assert.Equal(t, OrderSideBuy, body.Side)
		assert.Equal(t, OrderTypeLimit, body.Type)

		payload := map[string]interface{}{
			"oid": "new-order-123",
		}
		w.Write(successResponse(payload))
	})
	defer server.Close()

	client.SetAuth("test-key", "test-secret")
	order := &OrderPlacement{
		Book:  *NewBook(BTC, MXN),
		Side:  OrderSideBuy,
		Type:  OrderTypeLimit,
		Major: ToMonetary(0.1),
		Price: ToMonetary(500000),
	}
	oid, err := client.PlaceOrder(order)

	require.NoError(t, err)
	assert.Equal(t, "new-order-123", oid)
}

func TestCancelOrder(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/orders/order123"))

		w.Write(successResponse([]string{"order123"}))
	})
	defer server.Close()

	client.SetAuth("test-key", "test-secret")
	cancelled, err := client.CancelOrder("order123")

	require.NoError(t, err)
	require.Len(t, cancelled, 1)
	assert.Equal(t, "order123", cancelled[0])
}

func TestCancelOrders(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.True(t, strings.Contains(r.URL.Path, "/orders/order1,order2"))

		w.Write(successResponse([]string{"order1", "order2"}))
	})
	defer server.Close()

	client.SetAuth("test-key", "test-secret")
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
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		require.NotEmpty(t, auth)

		// Format: "Bitso key:nonce:signature"
		assert.True(t, strings.HasPrefix(auth, "Bitso "))
		parts := strings.Split(strings.TrimPrefix(auth, "Bitso "), ":")
		require.Len(t, parts, 3)
		assert.Equal(t, "test-key", parts[0])
		assert.NotEmpty(t, parts[1]) // nonce
		assert.NotEmpty(t, parts[2]) // signature

		w.Write(successResponse(map[string]interface{}{"balances": []interface{}{}}))
	})
	defer server.Close()

	client.SetAuth("test-key", "test-secret")
	_, err := client.Balances(nil)

	require.NoError(t, err)
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
	assert.Equal(t, "https://api.bitso.com/api/v3/balance", u.String())
}
