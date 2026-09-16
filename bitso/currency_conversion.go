package bitso

import (
	"fmt"
	"net/url"
	"strings"
)

// CurrencyConversionQuoteRequest represents a request for a fixed-rate
// conversion quote. Exactly one amount field must be set.
type CurrencyConversionQuoteRequest struct {
	FromCurrency  Currency `json:"from_currency"`
	ToCurrency    Currency `json:"to_currency"`
	SpendAmount   Monetary `json:"spend_amount,omitempty"`
	ReceiveAmount Monetary `json:"receive_amount,omitempty"`
}

// CurrencyConversionEstimatedSlippage describes the expected execution slippage
// for a conversion quote.
type CurrencyConversionEstimatedSlippage struct {
	Value   Monetary `json:"value"`
	Level   string   `json:"level"`
	Message string   `json:"message"`
}

// CurrencyConversionQuote represents a Bitso currency conversion quote.
type CurrencyConversionQuote struct {
	ID                  string                              `json:"id"`
	FromAmount          Monetary                            `json:"from_amount"`
	FromCurrency        Currency                            `json:"from_currency"`
	ToAmount            Monetary                            `json:"to_amount"`
	ToCurrency          Currency                            `json:"to_currency"`
	EstimatedSlippage   CurrencyConversionEstimatedSlippage `json:"estimated_slippage"`
	Created             int64                               `json:"created"`
	Expires             int64                               `json:"expires"`
	Rate                Monetary                            `json:"rate"`
	PlainRate           Monetary                            `json:"plain_rate"`
	RateCurrency        Currency                            `json:"rate_currency"`
	Padding             Monetary                            `json:"padding"`
	Book                Book                                `json:"book"`
	NextRecurrentEvents map[string]int64                    `json:"next_recurrent_events"`
}

// CurrencyConversionStatus represents the lifecycle state of a conversion.
type CurrencyConversionStatus string

const (
	CurrencyConversionStatusOpen      CurrencyConversionStatus = "open"
	CurrencyConversionStatusQueued    CurrencyConversionStatus = "queued"
	CurrencyConversionStatusCompleted CurrencyConversionStatus = "completed"
	CurrencyConversionStatusFailed    CurrencyConversionStatus = "failed"
)

// CurrencyConversion represents the status and execution details for a
// submitted currency conversion.
type CurrencyConversion struct {
	ID           string                   `json:"id"`
	FromAmount   Monetary                 `json:"from_amount"`
	FromCurrency Currency                 `json:"from_currency"`
	ToAmount     Monetary                 `json:"to_amount"`
	ToCurrency   Currency                 `json:"to_currency"`
	Created      int64                    `json:"created"`
	Expires      int64                    `json:"expires"`
	Rate         Monetary                 `json:"rate"`
	PlainRate    Monetary                 `json:"plain_rate"`
	RateCurrency Currency                 `json:"rate_currency"`
	Book         Book                     `json:"book"`
	Status       CurrencyConversionStatus `json:"status"`
}

func (r *CurrencyConversionQuoteRequest) validate() error {
	if r == nil {
		return fmt.Errorf("currency conversion quote request is required")
	}
	if r.FromCurrency == CurrencyNone {
		return fmt.Errorf("from_currency is required")
	}
	if r.ToCurrency == CurrencyNone {
		return fmt.Errorf("to_currency is required")
	}

	hasSpendAmount := r.SpendAmount != ""
	hasReceiveAmount := r.ReceiveAmount != ""
	if hasSpendAmount == hasReceiveAmount {
		return fmt.Errorf("exactly one of spend_amount or receive_amount is required")
	}
	if hasSpendAmount {
		return validatePositiveCurrencyConversionAmount("spend_amount", r.SpendAmount)
	}
	return validatePositiveCurrencyConversionAmount("receive_amount", r.ReceiveAmount)
}

func validatePositiveCurrencyConversionAmount(name string, amount Monetary) error {
	value, err := amount.Decimal()
	if err != nil {
		return fmt.Errorf("%s must be a valid decimal: %w", name, err)
	}
	if value.Sign() <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
}

func validateCurrencyConversionID(name, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// RequestCurrencyConversionQuote requests a fixed-rate currency conversion quote.
func (c *Client) RequestCurrencyConversionQuote(request *CurrencyConversionQuoteRequest) (*CurrencyConversionQuote, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	res := struct {
		Payload CurrencyConversionQuote `json:"payload"`
	}{}
	if err := c.postResponseForRoute(apiRouteV4, "/currency_conversions", request, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}

// ExecuteCurrencyConversion executes a previously requested conversion quote and
// returns the conversion ID.
func (c *Client) ExecuteCurrencyConversion(quoteID string) (string, error) {
	if err := validateCurrencyConversionID("quote_id", quoteID); err != nil {
		return "", err
	}

	res := struct {
		Payload struct {
			OID string `json:"oid"`
		} `json:"payload"`
	}{}
	endpoint := "/currency_conversions/" + url.PathEscape(quoteID)
	if err := c.putResponseForRoute(apiRouteV4, endpoint, nil, &res); err != nil {
		return "", err
	}
	return res.Payload.OID, nil
}

// CurrencyConversionStatus returns the latest status for a submitted conversion.
func (c *Client) CurrencyConversionStatus(conversionID string) (*CurrencyConversion, error) {
	if err := validateCurrencyConversionID("conversion_id", conversionID); err != nil {
		return nil, err
	}

	res := struct {
		Payload CurrencyConversion `json:"payload"`
	}{}
	endpoint := "/currency_conversions/" + url.PathEscape(conversionID)
	if err := c.getResponseForRoute(apiRouteV4, endpoint, nil, &res); err != nil {
		return nil, err
	}
	return &res.Payload, nil
}
