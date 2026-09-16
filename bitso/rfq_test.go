package bitso

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRFQEndpointURLs(t *testing.T) {
	client := NewClient()

	u, err := client.endpointURL("/balance")
	require.NoError(t, err, "default endpointURL should build v3 URLs")
	assert.Equal(t, "https://bitso.com/api/v3/balance", u.String(), "default endpointURL should keep the legacy API root")

	u, err = client.endpointURLForRoute(apiRouteV4, "/currency_conversions")
	require.NoError(t, err, "v4 endpointURL should build v4 URLs")
	assert.Equal(t, "https://bitso.com/api/v4/currency_conversions", u.String(), "v4 route should keep the legacy API root")

	u, err = client.endpointURLForRoute(apiRouteRFQ, "/pairs")
	require.NoError(t, err, "RFQ endpointURL should build product-root URLs")
	assert.Equal(t, "https://api.bitso.com/rfq/v1/pairs", u.String(), "RFQ route should use the RFQ product root by default")

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
}

func TestRFQPairs(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "RFQPairs should GET")
		assert.Equal(t, "/rfq/v1/pairs", r.URL.Path, "RFQPairs should use the RFQ product-root route")
		assert.Equal(t, "source=BTC&target=USDT", r.URL.RawQuery, "RFQPairs should send source and target filters")

		w.Write([]byte(`{
			"pairs": [
				{
					"source": "BTC",
					"target": "USDT",
					"source_precision": "8",
					"target_precision": "2"
				}
			]
		}`))
	})
	defer server.Close()

	pairs, err := client.RFQPairs(&RFQPairFilter{
		Source: "BTC",
		Target: "USDT",
	})

	require.NoError(t, err, "RFQPairs should decode a root-level pairs response")
	require.Len(t, pairs, 1, "RFQPairs should return all pairs from the response")
	assert.Equal(t, "BTC", pairs[0].Source, "source should decode")
	assert.Equal(t, "USDT", pairs[0].Target, "target should decode")
	assert.Equal(t, "8", pairs[0].SourcePrecision, "source precision should decode")
	assert.Equal(t, "2", pairs[0].TargetPrecision, "target precision should decode")
}

func TestRequestRFQQuoteDefaultBaseSignsBody(t *testing.T) {
	const nonce = "1731349200123456789"
	const key = "test-key"
	const secret = "tsecret"
	const expectedBody = `{"source":"BTC","target":"USDT","source_amount":"0.001"}`

	client := NewClient()
	client.SetAuth(key, secret)
	client.nonce = func() (string, error) { return nonce, nil }
	client.SetHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			assert.Equal(t, "POST", req.Method, "RequestRFQQuote should POST")
			assert.Equal(t, "api.bitso.com", req.URL.Host, "RequestRFQQuote should use the RFQ product host by default")
			assert.Equal(t, "/rfq/v1/quotes", req.URL.Path, "RequestRFQQuote should use the RFQ quote endpoint")
			assert.Empty(t, req.URL.RawQuery, "RequestRFQQuote should not send query parameters")
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"), "RequestRFQQuote should send JSON")

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err, "read RequestRFQQuote body")
			assert.Equal(t, expectedBody, string(body), "RequestRFQQuote should send the expected JSON body")

			expectedSignature := testSignature(secret, nonce+"POST"+"/rfq/v1/quotes"+expectedBody)
			assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, req.Header.Get("Authorization"), "RequestRFQQuote should sign the exact RFQ request body")

			return rfqHTTPResponse(req, `{
				"id": "quote-123",
				"source": "BTC",
				"target": "USDT",
				"source_amount": "0.001",
				"target_amount": "95.25",
				"rate": "95250.00",
				"status": "ACTIVE",
				"created_at": "2025-07-25T03:43:15Z",
				"expires_at": "2025-07-25T03:43:45.123456789Z",
				"can_confirm": false,
				"max_tradable_amount": {
					"value": "0.01",
					"currency": "BTC"
				}
			}`), nil
		}),
	})

	quote, err := client.RequestRFQQuote(&RFQQuoteRequest{
		Source:       "BTC",
		Target:       "USDT",
		SourceAmount: Monetary("0.001"),
	})

	require.NoError(t, err, "RequestRFQQuote should decode a root-level quote response")
	require.NotNil(t, quote, "RequestRFQQuote should return the quote")
	assert.Equal(t, "quote-123", quote.ID, "id should decode")
	assert.Equal(t, Monetary("0.001"), quote.SourceAmount, "source_amount should decode")
	assert.Equal(t, Monetary("95.25"), quote.TargetAmount, "target_amount should decode")
	assert.Equal(t, RFQQuoteStatusActive, quote.Status, "status should decode")
	assert.Equal(t, time.Date(2025, 7, 25, 3, 43, 15, 0, time.UTC), quote.CreatedAt.Time(), "created_at should decode RFC3339 Z timestamps")
	assert.Equal(t, time.Date(2025, 7, 25, 3, 43, 45, 123456789, time.UTC), quote.ExpiresAt.Time(), "expires_at should decode RFC3339Nano Z timestamps")
	require.NotNil(t, quote.MaxTradableAmount, "max_tradable_amount should decode")
	assert.Equal(t, Monetary("0.01"), quote.MaxTradableAmount.Value, "max_tradable_amount.value should decode")
	assert.Equal(t, "BTC", quote.MaxTradableAmount.Currency, "max_tradable_amount.currency should decode")
}

func TestRFQQuote(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "RFQQuote should GET")
		assert.Equal(t, "/rfq/v1/quotes/quote%20123", r.URL.EscapedPath(), "RFQQuote should path-escape quote IDs")
		assert.Empty(t, r.URL.RawQuery, "RFQQuote should not send query parameters")

		w.Write([]byte(`{
			"id": "quote 123",
			"source": "BTC",
			"target": "USDT",
			"source_amount": "0.001",
			"target_amount": "95.25",
			"rate": "95250.00",
			"status": "EXPIRED",
			"created_at": "2025-07-25T03:43:15Z",
			"expires_at": "2025-07-25T03:43:45Z",
			"can_confirm": true
		}`))
	})
	defer server.Close()

	quote, err := client.RFQQuote("quote 123")

	require.NoError(t, err, "RFQQuote should decode a root-level quote response")
	require.NotNil(t, quote, "RFQQuote should return the quote")
	assert.Equal(t, "quote 123", quote.ID, "id should decode")
	assert.Equal(t, RFQQuoteStatusExpired, quote.Status, "status should decode")
	assert.True(t, quote.CanConfirm, "can_confirm should decode")
	assert.Nil(t, quote.MaxTradableAmount, "omitted max_tradable_amount should decode as nil")
}

func TestConvertRFQQuote(t *testing.T) {
	const nonce = "1731349200123456789"
	const key = "test-key"
	const secret = "tsecret"
	const expectedBody = `{"quote_id":"quote-123"}`

	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method, "ConvertRFQQuote should POST")
		assert.Equal(t, "/rfq/v1/conversions", r.URL.Path, "ConvertRFQQuote should use the RFQ conversion endpoint")
		assert.Empty(t, r.URL.RawQuery, "ConvertRFQQuote should not send query parameters")
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"), "ConvertRFQQuote should send JSON")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err, "read ConvertRFQQuote body")
		assert.Equal(t, expectedBody, string(body), "ConvertRFQQuote should send the quote id body")

		expectedSignature := testSignature(secret, nonce+"POST"+"/rfq/v1/conversions"+expectedBody)
		assert.Equal(t, "Bitso "+key+":"+nonce+":"+expectedSignature, r.Header.Get("Authorization"), "ConvertRFQQuote should sign the exact RFQ request body")

		w.Write([]byte(`{
			"id": "conversion-123",
			"source": "BTC",
			"target": "USDT",
			"source_amount": "0.001",
			"target_amount": "95.25",
			"quote_id": "quote-123",
			"rate": "95250.00",
			"status": "COMPLETED",
			"quoted_at": "2025-07-25T03:43:15Z",
			"created_at": "2025-07-25T03:43:16Z",
			"updated_at": "2025-07-25T03:43:17Z"
		}`))
	})
	defer server.Close()

	client.SetAuth(key, secret)
	client.nonce = func() (string, error) { return nonce, nil }

	conversion, err := client.ConvertRFQQuote("quote-123")

	require.NoError(t, err, "ConvertRFQQuote should decode a root-level conversion response")
	require.NotNil(t, conversion, "ConvertRFQQuote should return the conversion")
	assert.Equal(t, "conversion-123", conversion.ID, "id should decode")
	assert.Equal(t, "quote-123", conversion.QuoteID, "quote_id should decode")
	assert.Equal(t, RFQConversionStatusCompleted, conversion.Status, "status should decode")
	assert.Equal(t, time.Date(2025, 7, 25, 3, 43, 17, 0, time.UTC), conversion.UpdatedAt.Time(), "updated_at should decode")
}

func TestRFQConversion(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "RFQConversion should GET")
		assert.Equal(t, "/rfq/v1/conversions/conversion%20123", r.URL.EscapedPath(), "RFQConversion should path-escape conversion IDs")
		assert.Empty(t, r.URL.RawQuery, "RFQConversion should not send query parameters")

		w.Write([]byte(`{
			"id": "conversion 123",
			"source": "BTC",
			"target": "USDT",
			"source_amount": "0.001",
			"target_amount": "95.25",
			"quote_id": "quote-123",
			"rate": "95250.00",
			"status": "PENDING",
			"quoted_at": "2025-07-25T03:43:15Z",
			"created_at": "2025-07-25T03:43:16Z",
			"updated_at": "2025-07-25T03:43:17Z"
		}`))
	})
	defer server.Close()

	conversion, err := client.RFQConversion("conversion 123")

	require.NoError(t, err, "RFQConversion should decode a root-level conversion response")
	require.NotNil(t, conversion, "RFQConversion should return the conversion")
	assert.Equal(t, "conversion 123", conversion.ID, "id should decode")
	assert.Equal(t, RFQConversionStatusPending, conversion.Status, "status should decode")
	assert.Equal(t, Monetary("95250.00"), conversion.Rate, "rate should decode")
}

func TestRFQStatusConstants(t *testing.T) {
	assert.Equal(t, RFQQuoteStatus("ACTIVE"), RFQQuoteStatusActive, "ACTIVE quote status should be exported")
	assert.Equal(t, RFQQuoteStatus("EXPIRED"), RFQQuoteStatusExpired, "EXPIRED quote status should be exported")
	assert.Equal(t, RFQQuoteStatus("ACCEPTED"), RFQQuoteStatusAccepted, "ACCEPTED quote status should be exported")
	assert.Equal(t, RFQQuoteStatus("REJECTED"), RFQQuoteStatusRejected, "REJECTED quote status should be exported")
	assert.Equal(t, RFQConversionStatus("PENDING"), RFQConversionStatusPending, "PENDING conversion status should be exported")
	assert.Equal(t, RFQConversionStatus("FAILED"), RFQConversionStatusFailed, "FAILED conversion status should be exported")
	assert.Equal(t, RFQConversionStatus("COMPLETED"), RFQConversionStatusCompleted, "COMPLETED conversion status should be exported")
}

func TestRFQErrors(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "RFQQuote should GET before decoding errors")
		assert.Equal(t, "/rfq/v1/quotes/missing", r.URL.Path, "RFQQuote should use the RFQ quote endpoint")

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{
			"errors": [
				{"code": "quote_not_found", "message": "Quote not found"},
				{"code": "forbidden_request", "message": "Forbidden"}
			]
		}`))
	})
	defer server.Close()

	quote, err := client.RFQQuote("missing")

	require.Error(t, err, "RFQQuote should return RFQ root-level errors")
	assert.Nil(t, quote, "RFQQuote should not return a quote when RFQ errors are present")
	var rfqErrors RFQErrors
	require.True(t, errors.As(err, &rfqErrors), "RFQ errors should support errors.As")
	require.Len(t, rfqErrors, 2, "RFQ errors should preserve every error item")
	assert.Equal(t, "quote_not_found", rfqErrors[0].Code, "first RFQ error code should decode")
	assert.Equal(t, "Quote not found", rfqErrors[0].Message, "first RFQ error message should decode")
	assert.Equal(t, "forbidden_request", rfqErrors[1].Code, "second RFQ error code should decode")
	assert.Equal(t, "Forbidden", rfqErrors[1].Message, "second RFQ error message should decode")
	assert.Contains(t, err.Error(), "quote_not_found: Quote not found", "RFQ error text should include the first item")
	assert.Contains(t, err.Error(), "forbidden_request: Forbidden", "RFQ error text should include the second item")
}

func TestRFQNonErrorStatus(t *testing.T) {
	server, client := mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method, "RFQPairs should GET before status validation")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"maintenance": true}`))
	})
	defer server.Close()

	pairs, err := client.RFQPairs(nil)

	require.Error(t, err, "non-2xx RFQ responses without RFQ errors should fail")
	assert.Nil(t, pairs, "non-2xx RFQ responses should not return pairs")
	assert.Contains(t, err.Error(), "status 503", "status error should include the HTTP status")
}

func TestRFQQuoteValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *RFQQuoteRequest
		wantErr string
	}{
		{
			name:    "nil request",
			request: nil,
			wantErr: "request is required",
		},
		{
			name: "missing source",
			request: &RFQQuoteRequest{
				Target:       "USDT",
				SourceAmount: Monetary("0.001"),
			},
			wantErr: "source is required",
		},
		{
			name: "missing target",
			request: &RFQQuoteRequest{
				Source:       "BTC",
				SourceAmount: Monetary("0.001"),
			},
			wantErr: "target is required",
		},
		{
			name: "missing amount",
			request: &RFQQuoteRequest{
				Source: "BTC",
				Target: "USDT",
			},
			wantErr: "exactly one of source_amount or target_amount is required",
		},
		{
			name: "both amounts",
			request: &RFQQuoteRequest{
				Source:       "BTC",
				Target:       "USDT",
				SourceAmount: Monetary("0.001"),
				TargetAmount: Monetary("95.25"),
			},
			wantErr: "exactly one of source_amount or target_amount is required",
		},
		{
			name: "invalid source amount",
			request: &RFQQuoteRequest{
				Source:       "BTC",
				Target:       "USDT",
				SourceAmount: Monetary("not-a-number"),
			},
			wantErr: "source_amount must be a valid decimal",
		},
		{
			name: "zero source amount",
			request: &RFQQuoteRequest{
				Source:       "BTC",
				Target:       "USDT",
				SourceAmount: Monetary("0"),
			},
			wantErr: "source_amount must be positive",
		},
		{
			name: "negative target amount",
			request: &RFQQuoteRequest{
				Source:       "BTC",
				Target:       "USDT",
				TargetAmount: Monetary("-1"),
			},
			wantErr: "target_amount must be positive",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := rfqValidationClient(t)

			quote, err := client.RequestRFQQuote(tc.request)

			require.Error(t, err, "invalid RFQ quote request should fail")
			assert.Contains(t, err.Error(), tc.wantErr, "RFQ quote validation error should explain the invalid field")
			assert.Nil(t, quote, "invalid RFQ quote request should not return a quote")
		})
	}
}

func TestRFQIDValidation(t *testing.T) {
	client := rfqValidationClient(t)

	quote, err := client.RFQQuote(" ")
	require.Error(t, err, "empty quote id should fail")
	assert.Contains(t, err.Error(), "quote_id is required", "quote_id validation should explain the invalid field")
	assert.Nil(t, quote, "invalid quote id should not return a quote")

	conversion, err := client.ConvertRFQQuote("")
	require.Error(t, err, "empty quote id should fail before conversion")
	assert.Contains(t, err.Error(), "quote_id is required", "quote_id validation should explain the invalid field")
	assert.Nil(t, conversion, "invalid quote id should not return a conversion")

	conversion, err = client.RFQConversion(" ")
	require.Error(t, err, "empty conversion id should fail")
	assert.Contains(t, err.Error(), "conversion_id is required", "conversion_id validation should explain the invalid field")
	assert.Nil(t, conversion, "invalid conversion id should not return a conversion")
}

func TestRFQCustomAPIBaseURL(t *testing.T) {
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/api/v3/available_books":
			assert.Equal(t, "GET", r.Method, "AvailableBooks should keep using GET")
			w.Write(successResponse([]map[string]interface{}{}))
		case "/rfq/v1/pairs":
			assert.Equal(t, "GET", r.Method, "RFQPairs should use GET")
			w.Write([]byte(`{"pairs":[]}`))
		default:
			t.Fatalf("unexpected request path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewClient()
	client.SetAPIBaseURL(server.URL + "/api")

	books, err := client.AvailableBooks()
	require.NoError(t, err, "available books should still use the legacy API root")
	assert.Empty(t, books, "available books response should decode")

	pairs, err := client.RFQPairs(nil)
	require.NoError(t, err, "RFQPairs should use the custom product root")
	assert.Empty(t, pairs, "RFQPairs response should decode")
	assert.Equal(t, []string{"/api/v3/available_books", "/rfq/v1/pairs"}, paths, "custom base should keep legacy routes under /api and RFQ routes at the product root")
}

func rfqValidationClient(t *testing.T) *Client {
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

func rfqHTTPResponse(req *http.Request, body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
