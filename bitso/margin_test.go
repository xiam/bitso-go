package bitso

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const marginAccountSummaryJSON = `{
	"status": "ACTIVE",
	"margin_level": {
		"current": "2.5",
		"initial": "1.25",
		"safe": "1.3",
		"maintenance": "1.1",
		"realized": "1.8",
		"unrealized": "2.5"
	},
	"margin_figure": {
		"currency": "usd",
		"total_assets_amount": "10000.00",
		"gross_assets_amount": "10500.00",
		"total_liabilities_amount": "4000.00",
		"trading_power_amount": "5000.00"
	},
	"balances": [
		{
			"currency": "btc",
			"total_amount": "0.50000000",
			"unrealized_amount": "0.45000000",
			"available_amount": "0.30000000"
		},
		{
			"currency": "usd",
			"total_amount": "-1000.00",
			"unrealized_amount": "-1050.00",
			"available_amount": "-1000.00"
		}
	]
}`

func TestMarginEndpointURLs(t *testing.T) {
	client := NewClient()

	u, err := client.endpointURL("/balance")
	require.NoError(t, err, "default endpointURL should build v3 URLs")
	assert.Equal(t, "https://bitso.com/api/v3/balance", u.String(), "default endpointURL should keep the legacy API root")

	u, err = client.endpointURLForRoute(apiRouteV4, "/currency_conversions")
	require.NoError(t, err, "v4 endpointURL should build v4 URLs")
	assert.Equal(t, "https://bitso.com/api/v4/currency_conversions", u.String(), "v4 route should keep the legacy API root")

	u, err = client.endpointURLForRoute(apiRouteRFQ, "/pairs")
	require.NoError(t, err, "RFQ endpointURL should build product-root URLs")
	assert.Equal(t, "https://api.bitso.com/rfq/v1/pairs", u.String(), "RFQ route should keep using its product root")

	u, err = client.endpointURLForRoute(apiRouteMarginV1Alpha, "/currencies")
	require.NoError(t, err, "margin v1 endpointURL should build margin URLs")
	assert.Equal(t, "https://api.bitso.com/margin_trading/v1alpha/currencies", u.String(), "margin v1 route should use the margin product root")

	u, err = client.endpointURLForRoute(apiRouteMarginV2Alpha, "/account")
	require.NoError(t, err, "margin v2 endpointURL should build margin URLs")
	assert.Equal(t, "https://api.bitso.com/margin_trading/v2alpha/account", u.String(), "margin v2 route should use the margin product root")

	client.SetAPIBaseURL("https://example.test/custom/api")

	u, err = client.endpointURL("/balance")
	require.NoError(t, err, "custom default endpointURL should build v3 URLs")
	assert.Equal(t, "https://example.test/custom/api/v3/balance", u.String(), "custom base should still apply to v3 routes")

	u, err = client.endpointURLForRoute(apiRouteV4, "/currency_conversions")
	require.NoError(t, err, "custom v4 endpointURL should build v4 URLs")
	assert.Equal(t, "https://example.test/custom/api/v4/currency_conversions", u.String(), "custom base should still apply to v4 routes")

	u, err = client.endpointURLForRoute(apiRouteRFQ, "/pairs")
	require.NoError(t, err, "custom RFQ endpointURL should build product-root URLs")
	assert.Equal(t, "https://example.test/custom/rfq/v1/pairs", u.String(), "custom RFQ base should strip the legacy /api suffix")

	u, err = client.endpointURLForRoute(apiRouteMarginV1Alpha, "/currencies")
	require.NoError(t, err, "custom margin v1 endpointURL should build margin URLs")
	assert.Equal(t, "https://example.test/custom/margin_trading/v1alpha/currencies", u.String(), "custom margin base should strip the legacy /api suffix")

	u, err = client.endpointURLForRoute(apiRouteMarginV2Alpha, "/account")
	require.NoError(t, err, "custom margin v2 endpointURL should build margin URLs")
	assert.Equal(t, "https://example.test/custom/margin_trading/v2alpha/account", u.String(), "custom margin base should strip the legacy /api suffix")
}

func TestCreateMarginAccountDefaultBaseSignsEmptyBody(t *testing.T) {
	const nonce = "1731349200123456789"
	const key = "test-key"
	const secret = "tsecret"

	client := NewClient()
	client.SetAuth(key, secret)
	client.nonce = func() (string, error) { return nonce, nil }
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method, "CreateMarginAccount should POST")
			assert.Equal(t, "api.bitso.com", req.URL.Host, "CreateMarginAccount should use the margin product host by default")
			assert.Equal(t, "/margin_trading/v2alpha/account", req.URL.Path, "CreateMarginAccount should use the current v2alpha account endpoint")
			assert.Empty(t, req.URL.RawQuery, "CreateMarginAccount should not send query parameters")
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "CreateMarginAccount should identify JSON requests")

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err, "read CreateMarginAccount body")
			assert.Empty(t, string(body), "CreateMarginAccount should send no JSON body")

			expectedSignature := testSignature(secret, nonce+"POST"+"/margin_trading/v2alpha/account")
			assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, req.Header.Get("Authorization"), "CreateMarginAccount should sign an empty request body")

			return marginHTTPResponse(req, http.StatusCreated, marginAccountSummaryJSON), nil
		}),
	})

	summary, err := client.CreateMarginAccount()

	require.NoError(t, err, "CreateMarginAccount should decode the root-level account summary")
	require.NotNil(t, summary, "CreateMarginAccount should return a summary")
	assertMarginAccountSummary(t, summary)
}

func TestMarginAccountSummary(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginAccountSummary should GET")
		assert.Equal(t, "/margin_trading/v2alpha/account", r.URL.Path, "MarginAccountSummary should use the current v2alpha account endpoint")
		assert.Empty(t, r.URL.RawQuery, "MarginAccountSummary should not send query parameters")

		w.Write([]byte(marginAccountSummaryJSON))
	})
	defer server.Close()

	summary, err := client.MarginAccountSummary()

	require.NoError(t, err, "MarginAccountSummary should decode the root-level account summary")
	require.NotNil(t, summary, "MarginAccountSummary should return a summary")
	assertMarginAccountSummary(t, summary)
}

func TestMarginCurrencies(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginCurrencies should GET")
		assert.Equal(t, "/margin_trading/v1alpha/currencies", r.URL.Path, "MarginCurrencies should use the margin currencies endpoint")
		assert.Empty(t, r.URL.RawQuery, "MarginCurrencies should not send query parameters")

		w.Write([]byte(`{
			"currencies": [
				{
					"currency_code": "btc",
					"interest_rate": "0.0001",
					"interest_rate_apy": "1.40117",
					"discount_factor": "0.9",
					"lending_pool_size": "50.00000000"
				},
				{
					"currency_code": "usd",
					"interest_rate": "0.00008",
					"interest_rate_apy": "1.01531",
					"discount_factor": "1.0",
					"lending_pool_size": "1000000.00"
				}
			]
		}`))
	})
	defer server.Close()

	currencies, err := client.MarginCurrencies()

	require.NoError(t, err, "MarginCurrencies should decode a root-level currencies response")
	require.Len(t, currencies, 2, "MarginCurrencies should return all currencies from the response")
	assert.Equal(t, Currency(BTC), currencies[0].CurrencyCode, "currency_code should decode")
	assert.Equal(t, Monetary("0.0001"), currencies[0].InterestRate, "interest_rate should decode")
	assert.Equal(t, Monetary("1.40117"), currencies[0].InterestRateAPY, "interest_rate_apy should decode")
	assert.Equal(t, Monetary("0.9"), currencies[0].DiscountFactor, "discount_factor should decode")
	assert.Equal(t, Monetary("50.00000000"), currencies[0].LendingPoolSize, "lending_pool_size should decode")
	assert.Equal(t, Currency(USD), currencies[1].CurrencyCode, "second currency_code should decode")
}

func TestMarginMovements(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginMovements should GET")
		assert.Equal(t, "/margin_trading/v1alpha/movements", r.URL.Path, "MarginMovements should use the movements endpoint")
		assert.Equal(t, "50", r.URL.Query().Get("page_size"), "MarginMovements should send page_size when positive")
		assert.Equal(t, "next-token", r.URL.Query().Get("page_token"), "MarginMovements should send page_token when present")

		w.Write([]byte(`{
			"movements": [
				{
					"id": "22c19e48-280d-4775-a9d9-b971f83cd09d",
					"type": "DEPOSIT",
					"amount": "1000.50",
					"currency": "btc",
					"created_at": "2025-12-03T10:15:30Z"
				},
				{
					"id": "33d29f59-391e-5886-ba0a-c082g94de0e",
					"type": "WITHDRAWAL",
					"amount": "500.25",
					"currency": "btc",
					"created_at": "2025-12-02T08:20:15Z"
				}
			],
			"next_page_token": "following-token"
		}`))
	})
	defer server.Close()

	movements, err := client.MarginMovements(&MarginPagination{
		PageSize:  50,
		PageToken: "next-token",
	})

	require.NoError(t, err, "MarginMovements should decode a root-level movement list")
	require.NotNil(t, movements, "MarginMovements should return a movement list")
	require.Len(t, movements.Movements, 2, "MarginMovements should return all movement entries")
	assert.Equal(t, "following-token", movements.NextPageToken, "next_page_token should decode")
	assert.Equal(t, MarginMovementTypeDeposit, movements.Movements[0].Type, "movement type should decode")
	assert.Equal(t, Monetary("1000.50"), movements.Movements[0].Amount, "movement amount should decode")
	assert.Equal(t, Currency(BTC), movements.Movements[0].Currency, "movement currency should decode")
	assert.Equal(t, time.Date(2025, 12, 3, 10, 15, 30, 0, time.UTC), movements.Movements[0].CreatedAt.Time(), "movement created_at should decode RFC3339 Z timestamps")
}

func TestMarginMovement(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginMovement should GET")
		assert.Equal(t, "/margin_trading/v1alpha/movements/movement%2F123", r.URL.EscapedPath(), "MarginMovement should path-escape movement IDs")
		assert.Empty(t, r.URL.RawQuery, "MarginMovement should not send query parameters")

		w.Write([]byte(`{
			"id": "movement/123",
			"type": "LIQUIDATION_TRANSFER",
			"amount": "20.00000000",
			"currency": "usd",
			"created_at": "2025-12-03T10:15:30Z"
		}`))
	})
	defer server.Close()

	movement, err := client.MarginMovement("movement/123")

	require.NoError(t, err, "MarginMovement should decode a root-level movement response")
	require.NotNil(t, movement, "MarginMovement should return the movement")
	assert.Equal(t, "movement/123", movement.ID, "movement ID should decode")
	assert.Equal(t, MarginMovementTypeLiquidationTransfer, movement.Type, "all documented movement types should decode")
	assert.Equal(t, Currency(USD), movement.Currency, "movement currency should decode")
}

func TestProcessMarginMovementSignsBodyAndHeaderDoesNotLeak(t *testing.T) {
	const nonce = "1731349200123456789"
	const key = "test-key"
	const secret = "tsecret"
	const expectedBody = `{"type":"DEPOSIT","amount":"1000.5000","currency":"btc"}`

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /margin_trading/v1alpha/movements":
			assert.Equal(t, "idem-123", r.Header.Get("X-Idempotency-Key"), "ProcessMarginMovement should set the idempotency header on the processing request")
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "ProcessMarginMovement should send JSON")

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err, "read ProcessMarginMovement body")
			assert.Equal(t, expectedBody, string(body), "ProcessMarginMovement should send the documented request body")

			expectedSignature := testSignature(secret, nonce+"POST"+"/margin_trading/v1alpha/movements"+expectedBody)
			assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "ProcessMarginMovement should sign the exact request body")

			w.Write([]byte(`{
				"id": "movement-123",
				"type": "DEPOSIT",
				"amount": "1000.5000",
				"currency": "btc",
				"created_at": "2025-12-03T10:15:30Z"
			}`))
		case "GET /margin_trading/v1alpha/movements":
			assert.Empty(t, r.Header.Get("X-Idempotency-Key"), "idempotency header should not leak into later margin requests")
			assert.Empty(t, r.URL.RawQuery, "empty pagination should omit query parameters")
			w.Write([]byte(`{"movements":[]}`))
		default:
			t.Fatalf("unexpected margin request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) { return nonce, nil }

	movement, err := client.ProcessMarginMovement("idem-123", &MarginMovementRequest{
		Type:     MarginMovementTypeDeposit,
		Amount:   Monetary("1000.5000"),
		Currency: BTC,
	})
	require.NoError(t, err, "ProcessMarginMovement should decode a root-level movement response")
	require.NotNil(t, movement, "ProcessMarginMovement should return the movement")
	assert.Equal(t, "movement-123", movement.ID, "processed movement ID should decode")
	assert.Equal(t, MarginMovementTypeDeposit, movement.Type, "processed movement type should decode")

	movements, err := client.MarginMovements(&MarginPagination{})
	require.NoError(t, err, "later margin requests should still succeed")
	require.Empty(t, movements.Movements, "empty movement pages should decode")
}

func TestMarginLoans(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginLoans should GET")
		assert.Equal(t, "/margin_trading/v1alpha/loans", r.URL.Path, "MarginLoans should use the loans endpoint")
		assert.Equal(t, "25", r.URL.Query().Get("page_size"), "MarginLoans should send page_size when positive")
		assert.Equal(t, "loan-token", r.URL.Query().Get("page_token"), "MarginLoans should send page_token when present")

		w.Write([]byte(`{
			"loans": [
				{
					"currency": "usd",
					"principal": "5000.00",
					"interest_rate": "0.0001",
					"interest_rate_apy": "1.40117",
					"accrued_interest": "1.25",
					"last_accrual_at": "2025-02-27T10:00:00Z",
					"next_accrual_after": "2025-02-27T11:00:00Z"
				},
				{
					"currency": "btc",
					"principal": "0.01000000",
					"interest_rate": "0.00012",
					"interest_rate_apy": "1.86090",
					"accrued_interest": "0.00000001",
					"last_accrual_at": null,
					"next_accrual_after": null
				},
				{
					"currency": "eth",
					"principal": "1.00000000",
					"interest_rate": "0.00012",
					"interest_rate_apy": "1.86090",
					"accrued_interest": "0.00000001"
				}
			],
			"next_page_token": "following-token"
		}`))
	})
	defer server.Close()

	loans, err := client.MarginLoans(&MarginPagination{
		PageSize:  25,
		PageToken: "loan-token",
	})

	require.NoError(t, err, "MarginLoans should decode a root-level loan list")
	require.NotNil(t, loans, "MarginLoans should return a loan list")
	require.Len(t, loans.Loans, 3, "MarginLoans should return every loan entry")
	assert.Equal(t, "following-token", loans.NextPageToken, "next_page_token should decode")
	assert.Equal(t, Currency(USD), loans.Loans[0].Currency, "loan currency should decode")
	assert.Equal(t, Monetary("5000.00"), loans.Loans[0].Principal, "loan principal should decode")
	assert.Equal(t, Monetary("1.25"), loans.Loans[0].AccruedInterest, "accrued_interest should decode")
	require.NotNil(t, loans.Loans[0].LastAccrualAt, "last_accrual_at should decode when present")
	assert.Equal(t, time.Date(2025, 2, 27, 10, 0, 0, 0, time.UTC), loans.Loans[0].LastAccrualAt.Time(), "last_accrual_at should decode RFC3339 Z timestamps")
	require.NotNil(t, loans.Loans[0].NextAccrualAfter, "next_accrual_after should decode when present")
	assert.Equal(t, time.Date(2025, 2, 27, 11, 0, 0, 0, time.UTC), loans.Loans[0].NextAccrualAfter.Time(), "next_accrual_after should decode RFC3339 Z timestamps")
	assert.Nil(t, loans.Loans[1].LastAccrualAt, "null last_accrual_at should decode as nil")
	assert.Nil(t, loans.Loans[1].NextAccrualAfter, "null next_accrual_after should decode as nil")
	assert.Nil(t, loans.Loans[2].LastAccrualAt, "omitted last_accrual_at should decode as nil")
	assert.Nil(t, loans.Loans[2].NextAccrualAfter, "omitted next_accrual_after should decode as nil")
}

func TestMarginErrors(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginAccountSummary should GET before decoding errors")
		assert.Equal(t, "/margin_trading/v2alpha/account", r.URL.Path, "MarginAccountSummary should use the v2alpha account endpoint")

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"errors": [
				{"code": "not_eligible_user", "message": "User is not eligible for margin trading"},
				{"code": "missing_parameter", "message": "Missing required field"}
			]
		}`))
	})
	defer server.Close()

	summary, err := client.MarginAccountSummary()

	require.Error(t, err, "MarginAccountSummary should return margin root-level errors")
	assert.Nil(t, summary, "MarginAccountSummary should not return a summary when margin errors are present")
	var marginErrors MarginErrors
	require.True(t, errors.As(err, &marginErrors), "margin errors should support errors.As")
	require.Len(t, marginErrors, 2, "margin errors should preserve every error item")
	assert.Equal(t, "not_eligible_user", marginErrors[0].Code, "first margin error code should decode")
	assert.Equal(t, "User is not eligible for margin trading", marginErrors[0].Message, "first margin error message should decode")
	assert.Equal(t, "missing_parameter", marginErrors[1].Code, "second margin error code should decode")
	assert.Equal(t, "Missing required field", marginErrors[1].Message, "second margin error message should decode")
	assert.Contains(t, err.Error(), "not_eligible_user: User is not eligible for margin trading", "margin error text should include the first item")
	assert.Contains(t, err.Error(), "missing_parameter: Missing required field", "margin error text should include the second item")
}

func TestMarginNonErrorStatus(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "MarginCurrencies should GET before status validation")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"maintenance": true}`))
	})
	defer server.Close()

	currencies, err := client.MarginCurrencies()

	require.Error(t, err, "non-2xx margin responses without margin errors should fail")
	assert.Nil(t, currencies, "non-2xx margin responses should not return currencies")
	assert.Contains(t, err.Error(), "status 503", "status error should include the HTTP status")
}

func TestMarginStatusAndMovementTypeConstants(t *testing.T) {
	assert.Equal(t, MarginAccountStatus("ACTIVE"), MarginAccountStatusActive, "ACTIVE account status should be exported")
	assert.Equal(t, MarginAccountStatus("INACTIVE"), MarginAccountStatusInactive, "INACTIVE account status should be exported")
	assert.Equal(t, MarginAccountStatus("BLOCKED"), MarginAccountStatusBlocked, "BLOCKED account status should be exported")
	assert.Equal(t, MarginAccountStatus("LIQUIDATING"), MarginAccountStatusLiquidating, "LIQUIDATING account status should be exported")
	assert.Equal(t, MarginAccountStatus("ONBOARDING"), MarginAccountStatusOnboarding, "ONBOARDING account status should be exported")
	assert.Equal(t, MarginAccountStatus("ONBOARDED"), MarginAccountStatusOnboarded, "ONBOARDED account status should be exported")

	assert.Equal(t, MarginMovementType("DEPOSIT"), MarginMovementTypeDeposit, "DEPOSIT movement type should be exported")
	assert.Equal(t, MarginMovementType("WITHDRAWAL"), MarginMovementTypeWithdrawal, "WITHDRAWAL movement type should be exported")
	assert.Equal(t, MarginMovementType("LIQUIDATION_TRANSFER"), MarginMovementTypeLiquidationTransfer, "LIQUIDATION_TRANSFER movement type should be exported")
	assert.Equal(t, MarginMovementType("LIQUIDATION_DEBT_TRANSFER"), MarginMovementTypeLiquidationDebtTransfer, "LIQUIDATION_DEBT_TRANSFER movement type should be exported")
	assert.Equal(t, MarginMovementType("LIQUIDATION_REFUND"), MarginMovementTypeLiquidationRefund, "LIQUIDATION_REFUND movement type should be exported")
	assert.Equal(t, MarginMovementType("INTEREST_CHARGE"), MarginMovementTypeInterestCharge, "INTEREST_CHARGE movement type should be exported")
	assert.Equal(t, MarginMovementType("LENDING_POOL"), MarginMovementTypeLendingPool, "LENDING_POOL movement type should be exported")
}

func TestProcessMarginMovementValidation(t *testing.T) {
	validRequest := &MarginMovementRequest{
		Type:     MarginMovementTypeDeposit,
		Amount:   Monetary("1"),
		Currency: BTC,
	}

	tests := []struct {
		name           string
		request        *MarginMovementRequest
		idempotencyKey string
		wantErr        string
	}{
		{
			name:           "missing idempotency key",
			request:        validRequest,
			idempotencyKey: " ",
			wantErr:        "idempotency key is required",
		},
		{
			name:           "nil request",
			request:        nil,
			idempotencyKey: "idem-123",
			wantErr:        "margin movement request is required",
		},
		{
			name: "missing type",
			request: &MarginMovementRequest{
				Amount:   Monetary("1"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "type is required",
		},
		{
			name: "unsupported type",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeLiquidationTransfer,
				Amount:   Monetary("1"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "not supported",
		},
		{
			name: "missing currency",
			request: &MarginMovementRequest{
				Type:   MarginMovementTypeDeposit,
				Amount: Monetary("1"),
			},
			idempotencyKey: "idem-123",
			wantErr:        "currency is required",
		},
		{
			name: "missing amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "amount is required",
		},
		{
			name: "exponent amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("1e3"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "without signs or exponent notation",
		},
		{
			name: "signed amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("+1"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "without signs or exponent notation",
		},
		{
			name: "leading-dot amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary(".1"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "digits before the decimal point",
		},
		{
			name: "zero amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("0"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "amount must be positive",
		},
		{
			name: "zero fractional amount",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("0.000000000"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "amount must be positive",
		},
		{
			name: "non-zero ninth decimal",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("1.123456789"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "at most 8 non-zero fractional digits",
		},
		{
			name: "non-zero after trailing zero allowance",
			request: &MarginMovementRequest{
				Type:     MarginMovementTypeDeposit,
				Amount:   Monetary("1.1234567801"),
				Currency: BTC,
			},
			idempotencyKey: "idem-123",
			wantErr:        "at most 8 non-zero fractional digits",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := marginValidationClient(t)

			movement, err := client.ProcessMarginMovement(tc.idempotencyKey, tc.request)

			require.Error(t, err, "invalid margin movement request should fail")
			assert.Contains(t, err.Error(), tc.wantErr, "margin movement validation error should explain the invalid field")
			assert.Nil(t, movement, "invalid margin movement request should not return a movement")
		})
	}
}

func TestMarginMovementAmountValidationAllowsDocumentedFormat(t *testing.T) {
	validAmounts := []Monetary{
		Monetary("1"),
		Monetary("1.1"),
		Monetary("1.12345678"),
		Monetary("1.1234567800"),
		Monetary("0.00000001"),
		Monetary("0001.000000000"),
	}

	for _, amount := range validAmounts {
		amount := amount
		t.Run(string(amount), func(t *testing.T) {
			assert.NoError(t, validateMarginMovementAmount(amount), "valid margin movement amount should match the documented decimal rule")
		})
	}
}

func TestMarginPaginationAndIDValidation(t *testing.T) {
	client := marginValidationClient(t)

	movements, err := client.MarginMovements(&MarginPagination{PageSize: -1})
	require.Error(t, err, "negative movement page_size should fail before sending a request")
	assert.Contains(t, err.Error(), "page_size", "negative movement page_size error should name the invalid field")
	assert.Nil(t, movements, "invalid movement pagination should not return movements")

	loans, err := client.MarginLoans(&MarginPagination{PageSize: -1})
	require.Error(t, err, "negative loan page_size should fail before sending a request")
	assert.Contains(t, err.Error(), "page_size", "negative loan page_size error should name the invalid field")
	assert.Nil(t, loans, "invalid loan pagination should not return loans")

	movement, err := client.MarginMovement(" ")
	require.Error(t, err, "empty movement_id should fail before sending a request")
	assert.Contains(t, err.Error(), "movement_id is required", "empty movement_id error should name the invalid field")
	assert.Nil(t, movement, "invalid movement_id should not return a movement")
}

func assertMarginAccountSummary(t *testing.T, summary *MarginAccountSummary) {
	t.Helper()

	assert.Equal(t, MarginAccountStatusActive, summary.Status, "status should decode")
	assert.Equal(t, Monetary("2.5"), summary.MarginLevel.Current, "margin_level.current should decode")
	assert.Equal(t, Monetary("1.25"), summary.MarginLevel.Initial, "margin_level.initial should decode")
	assert.Equal(t, Monetary("1.3"), summary.MarginLevel.Safe, "margin_level.safe should decode")
	assert.Equal(t, Monetary("1.1"), summary.MarginLevel.Maintenance, "margin_level.maintenance should decode")
	assert.Equal(t, Monetary("1.8"), summary.MarginLevel.Realized, "margin_level.realized should decode")
	assert.Equal(t, Monetary("2.5"), summary.MarginLevel.Unrealized, "margin_level.unrealized should decode")
	assert.Equal(t, Currency(USD), summary.MarginFigure.Currency, "margin_figure.currency should decode")
	assert.Equal(t, Monetary("10000.00"), summary.MarginFigure.TotalAssetsAmount, "margin_figure.total_assets_amount should decode")
	assert.Equal(t, Monetary("10500.00"), summary.MarginFigure.GrossAssetsAmount, "margin_figure.gross_assets_amount should decode")
	assert.Equal(t, Monetary("4000.00"), summary.MarginFigure.TotalLiabilitiesAmount, "margin_figure.total_liabilities_amount should decode")
	assert.Equal(t, Monetary("5000.00"), summary.MarginFigure.TradingPowerAmount, "margin_figure.trading_power_amount should decode")
	require.Len(t, summary.Balances, 2, "balances should decode")
	assert.Equal(t, Currency(BTC), summary.Balances[0].Currency, "balance currency should decode")
	assert.Equal(t, Monetary("0.50000000"), summary.Balances[0].TotalAmount, "balance total_amount should decode")
	assert.Equal(t, Monetary("0.45000000"), summary.Balances[0].UnrealizedAmount, "balance unrealized_amount should decode")
	assert.Equal(t, Monetary("0.30000000"), summary.Balances[0].AvailableAmount, "balance available_amount should decode")
	assert.Equal(t, Currency(USD), summary.Balances[1].Currency, "second balance currency should decode")
	assert.Equal(t, Monetary("-1000.00"), summary.Balances[1].TotalAmount, "negative total_amount should decode")
	assert.Equal(t, Monetary("-1050.00"), summary.Balances[1].UnrealizedAmount, "negative unrealized_amount should decode")
	assert.Equal(t, Monetary("-1000.00"), summary.Balances[1].AvailableAmount, "negative available_amount should decode")
}

func marginValidationClient(t *testing.T) *Client {
	t.Helper()

	client := NewClient()
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			t.Fatalf("validation error should happen before sending request to %s", req.URL.String())
			return nil, nil
		}),
	})
	return client
}

func marginHTTPResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
