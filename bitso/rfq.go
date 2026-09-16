package bitso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const rfqBaseURL = "https://api.bitso.com"

var apiRouteRFQ = apiRoute{
	path:               "rfq/v1",
	productRootBaseURL: rfqBaseURL,
}

// RFQPairFilter filters available RFQ pairs.
type RFQPairFilter struct {
	Source string
	Target string
}

// RFQPair describes a source/target RFQ pair and its decimal precision.
type RFQPair struct {
	Source          string `json:"source"`
	Target          string `json:"target"`
	SourcePrecision string `json:"source_precision"`
	TargetPrecision string `json:"target_precision"`
}

// RFQQuoteRequest represents a request for an RFQ quote.
type RFQQuoteRequest struct {
	Source       string   `json:"source"`
	Target       string   `json:"target"`
	SourceAmount Monetary `json:"source_amount,omitempty"`
	TargetAmount Monetary `json:"target_amount,omitempty"`
}

// RFQQuoteStatus represents the lifecycle state of an RFQ quote.
type RFQQuoteStatus string

const (
	RFQQuoteStatusActive   RFQQuoteStatus = "ACTIVE"
	RFQQuoteStatusExpired  RFQQuoteStatus = "EXPIRED"
	RFQQuoteStatusAccepted RFQQuoteStatus = "ACCEPTED"
	RFQQuoteStatusRejected RFQQuoteStatus = "REJECTED"
)

// RFQTradableAmount describes the maximum tradable amount returned for a non-confirmable quote.
type RFQTradableAmount struct {
	Value    Monetary `json:"value"`
	Currency string   `json:"currency"`
}

// RFQQuote represents a Bitso RFQ quote.
type RFQQuote struct {
	ID                string             `json:"id"`
	Source            string             `json:"source"`
	Target            string             `json:"target"`
	SourceAmount      Monetary           `json:"source_amount"`
	TargetAmount      Monetary           `json:"target_amount"`
	Rate              Monetary           `json:"rate"`
	Status            RFQQuoteStatus     `json:"status"`
	CreatedAt         Time               `json:"created_at"`
	ExpiresAt         Time               `json:"expires_at"`
	CanConfirm        bool               `json:"can_confirm"`
	MaxTradableAmount *RFQTradableAmount `json:"max_tradable_amount,omitempty"`
}

// RFQConversionStatus represents the lifecycle state of an RFQ conversion.
type RFQConversionStatus string

const (
	RFQConversionStatusPending   RFQConversionStatus = "PENDING"
	RFQConversionStatusFailed    RFQConversionStatus = "FAILED"
	RFQConversionStatusCompleted RFQConversionStatus = "COMPLETED"
)

// RFQConversion represents an RFQ conversion transaction.
type RFQConversion struct {
	ID           string              `json:"id"`
	Source       string              `json:"source"`
	Target       string              `json:"target"`
	SourceAmount Monetary            `json:"source_amount"`
	TargetAmount Monetary            `json:"target_amount"`
	QuoteID      string              `json:"quote_id"`
	Rate         Monetary            `json:"rate"`
	Status       RFQConversionStatus `json:"status"`
	QuotedAt     Time                `json:"quoted_at"`
	CreatedAt    Time                `json:"created_at"`
	UpdatedAt    Time                `json:"updated_at"`
}

// RFQError is one error item returned by the RFQ API.
type RFQError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RFQErrors preserves all error items returned by the RFQ API.
type RFQErrors []RFQError

func (e RFQErrors) Error() string {
	switch len(e) {
	case 0:
		return "RFQ error"
	case 1:
		return fmt.Sprintf("RFQ error %s: %s", e[0].Code, e[0].Message)
	default:
		parts := make([]string, 0, len(e))
		for _, item := range e {
			parts = append(parts, fmt.Sprintf("%s: %s", item.Code, item.Message))
		}
		return "RFQ errors: " + strings.Join(parts, "; ")
	}
}

func (f *RFQPairFilter) values() url.Values {
	if f == nil {
		return nil
	}
	values := url.Values{}
	if f.Source != "" {
		values.Set("source", f.Source)
	}
	if f.Target != "" {
		values.Set("target", f.Target)
	}
	return values
}

func (r *RFQQuoteRequest) validate() error {
	if r == nil {
		return fmt.Errorf("RFQ quote request is required")
	}
	if strings.TrimSpace(r.Source) == "" {
		return fmt.Errorf("source is required")
	}
	if strings.TrimSpace(r.Target) == "" {
		return fmt.Errorf("target is required")
	}

	hasSourceAmount := r.SourceAmount != ""
	hasTargetAmount := r.TargetAmount != ""
	if hasSourceAmount == hasTargetAmount {
		return fmt.Errorf("exactly one of source_amount or target_amount is required")
	}
	if hasSourceAmount {
		return validatePositiveRFQAmount("source_amount", r.SourceAmount)
	}
	return validatePositiveRFQAmount("target_amount", r.TargetAmount)
}

func validatePositiveRFQAmount(name string, amount Monetary) error {
	value, err := amount.Decimal()
	if err != nil {
		return fmt.Errorf("%s must be a valid decimal: %w", name, err)
	}
	if value.Sign() <= 0 {
		return fmt.Errorf("%s must be positive", name)
	}
	return nil
}

func validateRFQID(name, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// RFQPairs returns available RFQ source/target pairs.
func (c *Client) RFQPairs(filter *RFQPairFilter) ([]RFQPair, error) {
	res := struct {
		Pairs []RFQPair `json:"pairs"`
	}{}
	if err := c.getRFQResponse("/pairs", filter.values(), &res); err != nil {
		return nil, err
	}
	return res.Pairs, nil
}

// RequestRFQQuote requests an RFQ quote.
func (c *Client) RequestRFQQuote(request *RFQQuoteRequest) (*RFQQuote, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}

	var quote RFQQuote
	if err := c.postRFQResponse("/quotes", request, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

// RFQQuote retrieves an RFQ quote by ID.
func (c *Client) RFQQuote(quoteID string) (*RFQQuote, error) {
	if err := validateRFQID("quote_id", quoteID); err != nil {
		return nil, err
	}

	var quote RFQQuote
	endpoint := "/quotes/" + url.PathEscape(quoteID)
	if err := c.getRFQResponse(endpoint, nil, &quote); err != nil {
		return nil, err
	}
	return &quote, nil
}

// ConvertRFQQuote converts a previously requested RFQ quote.
func (c *Client) ConvertRFQQuote(quoteID string) (*RFQConversion, error) {
	if err := validateRFQID("quote_id", quoteID); err != nil {
		return nil, err
	}

	request := struct {
		QuoteID string `json:"quote_id"`
	}{
		QuoteID: quoteID,
	}

	var conversion RFQConversion
	if err := c.postRFQResponse("/conversions", request, &conversion); err != nil {
		return nil, err
	}
	return &conversion, nil
}

// RFQConversion retrieves an RFQ conversion by ID.
func (c *Client) RFQConversion(conversionID string) (*RFQConversion, error) {
	if err := validateRFQID("conversion_id", conversionID); err != nil {
		return nil, err
	}

	var conversion RFQConversion
	endpoint := "/conversions/" + url.PathEscape(conversionID)
	if err := c.getRFQResponse(endpoint, nil, &conversion); err != nil {
		return nil, err
	}
	return &conversion, nil
}

func (c *Client) getRFQResponse(endpoint string, params url.Values, dest interface{}) error {
	return c.doDirectResponseForRoute("GET", apiRouteRFQ, endpoint, params, nil, dest, decodeRFQErrors)
}

func (c *Client) postRFQResponse(endpoint string, body interface{}, dest interface{}) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.doDirectResponseForRoute("POST", apiRouteRFQ, endpoint, nil, bytes.NewBuffer(buf), dest, decodeRFQErrors)
}

func decodeRFQErrors(buf []byte) error {
	res := struct {
		Errors RFQErrors `json:"errors"`
	}{}
	if err := json.Unmarshal(buf, &res); err != nil {
		return nil
	}
	if len(res.Errors) > 0 {
		return res.Errors
	}
	return nil
}
