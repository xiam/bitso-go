package bitso

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const marginBaseURL = "https://api.bitso.com"

var (
	apiRouteMarginV1Alpha = apiRoute{
		path:               "margin_trading/v1alpha",
		productRootBaseURL: marginBaseURL,
	}
	apiRouteMarginV2Alpha = apiRoute{
		path:               "margin_trading/v2alpha",
		productRootBaseURL: marginBaseURL,
	}
)

// MarginAccountStatus represents the lifecycle state of a margin account.
type MarginAccountStatus string

const (
	MarginAccountStatusActive      MarginAccountStatus = "ACTIVE"
	MarginAccountStatusInactive    MarginAccountStatus = "INACTIVE"
	MarginAccountStatusBlocked     MarginAccountStatus = "BLOCKED"
	MarginAccountStatusLiquidating MarginAccountStatus = "LIQUIDATING"
	MarginAccountStatusOnboarding  MarginAccountStatus = "ONBOARDING"
	MarginAccountStatusOnboarded   MarginAccountStatus = "ONBOARDED"
)

// MarginLevels contains the account's current and threshold margin levels.
type MarginLevels struct {
	Current     Monetary `json:"current"`
	Initial     Monetary `json:"initial"`
	Safe        Monetary `json:"safe"`
	Maintenance Monetary `json:"maintenance"`
	Realized    Monetary `json:"realized"`
	Unrealized  Monetary `json:"unrealized"`
}

// MarginFigure contains margin account figures expressed in one currency.
type MarginFigure struct {
	Currency               Currency `json:"currency"`
	TotalAssetsAmount      Monetary `json:"total_assets_amount"`
	GrossAssetsAmount      Monetary `json:"gross_assets_amount"`
	TotalLiabilitiesAmount Monetary `json:"total_liabilities_amount"`
	TradingPowerAmount     Monetary `json:"trading_power_amount"`
}

// MarginBalance contains one margin balance entry.
type MarginBalance struct {
	Currency         Currency `json:"currency"`
	TotalAmount      Monetary `json:"total_amount"`
	UnrealizedAmount Monetary `json:"unrealized_amount"`
	AvailableAmount  Monetary `json:"available_amount"`
}

// MarginAccountSummary contains a margin account status, risk metrics, and balances.
type MarginAccountSummary struct {
	Status       MarginAccountStatus `json:"status"`
	MarginLevel  MarginLevels        `json:"margin_level"`
	MarginFigure MarginFigure        `json:"margin_figure"`
	Balances     []MarginBalance     `json:"balances"`
}

// MarginCurrency contains one currency available for margin trading.
type MarginCurrency struct {
	CurrencyCode    Currency `json:"currency_code"`
	InterestRate    Monetary `json:"interest_rate"`
	InterestRateAPY Monetary `json:"interest_rate_apy"`
	DiscountFactor  Monetary `json:"discount_factor"`
	LendingPoolSize Monetary `json:"lending_pool_size"`
}

// MarginPagination filters paginated margin endpoints.
type MarginPagination struct {
	PageSize  int
	PageToken string
}

// MarginMovementType represents a margin funds movement type.
type MarginMovementType string

const (
	MarginMovementTypeDeposit                 MarginMovementType = "DEPOSIT"
	MarginMovementTypeWithdrawal              MarginMovementType = "WITHDRAWAL"
	MarginMovementTypeLiquidationTransfer     MarginMovementType = "LIQUIDATION_TRANSFER"
	MarginMovementTypeLiquidationDebtTransfer MarginMovementType = "LIQUIDATION_DEBT_TRANSFER"
	MarginMovementTypeLiquidationRefund       MarginMovementType = "LIQUIDATION_REFUND"
	MarginMovementTypeInterestCharge          MarginMovementType = "INTEREST_CHARGE"
	MarginMovementTypeLendingPool             MarginMovementType = "LENDING_POOL"
)

// MarginMovement contains one margin funds movement.
type MarginMovement struct {
	ID        string             `json:"id"`
	Type      MarginMovementType `json:"type"`
	Amount    Monetary           `json:"amount"`
	Currency  Currency           `json:"currency"`
	CreatedAt Time               `json:"created_at"`
}

// MarginMovementList contains a page of margin funds movements.
type MarginMovementList struct {
	Movements     []MarginMovement `json:"movements"`
	NextPageToken string           `json:"next_page_token"`
}

// MarginMovementRequest moves funds between main and margin account balances.
type MarginMovementRequest struct {
	Type     MarginMovementType `json:"type"`
	Amount   Monetary           `json:"amount"`
	Currency Currency           `json:"currency"`
}

// MarginLoan contains one margin loan entry.
type MarginLoan struct {
	Currency         Currency `json:"currency"`
	Principal        Monetary `json:"principal"`
	InterestRate     Monetary `json:"interest_rate"`
	InterestRateAPY  Monetary `json:"interest_rate_apy"`
	AccruedInterest  Monetary `json:"accrued_interest"`
	LastAccrualAt    *Time    `json:"last_accrual_at"`
	NextAccrualAfter *Time    `json:"next_accrual_after"`
}

// MarginLoanList contains a page of margin loans.
type MarginLoanList struct {
	Loans         []MarginLoan `json:"loans"`
	NextPageToken string       `json:"next_page_token"`
}

// MarginError is one error item returned by the margin trading API.
type MarginError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MarginErrors preserves all error items returned by the margin trading API.
type MarginErrors []MarginError

func (e MarginErrors) Error() string {
	switch len(e) {
	case 0:
		return "margin error"
	case 1:
		return fmt.Sprintf("margin error %s: %s", e[0].Code, e[0].Message)
	default:
		parts := make([]string, 0, len(e))
		for _, item := range e {
			parts = append(parts, fmt.Sprintf("%s: %s", item.Code, item.Message))
		}
		return "margin errors: " + strings.Join(parts, "; ")
	}
}

func (p *MarginPagination) values() (url.Values, error) {
	if p == nil {
		return nil, nil
	}
	if p.PageSize < 0 {
		return nil, fmt.Errorf("page_size must be zero or positive")
	}

	values := url.Values{}
	if p.PageSize > 0 {
		values.Set("page_size", strconv.Itoa(p.PageSize))
	}
	if p.PageToken != "" {
		values.Set("page_token", p.PageToken)
	}
	return values, nil
}

func (r *MarginMovementRequest) validate() error {
	if r == nil {
		return fmt.Errorf("margin movement request is required")
	}
	switch r.Type {
	case "":
		return fmt.Errorf("type is required")
	case MarginMovementTypeDeposit, MarginMovementTypeWithdrawal:
	default:
		return fmt.Errorf("type %q is not supported for processing margin movements", r.Type)
	}
	if strings.TrimSpace(r.Currency.String()) == "" {
		return fmt.Errorf("currency is required")
	}
	if err := validateMarginMovementAmount(r.Amount); err != nil {
		return err
	}
	return nil
}

func validateMarginMovementAmount(amount Monetary) error {
	s := string(amount)
	if s == "" {
		return fmt.Errorf("amount is required")
	}

	intPart, fracPart, hasFraction := strings.Cut(s, ".")
	if intPart == "" {
		return fmt.Errorf("amount must be a positive decimal with digits before the decimal point")
	}
	if hasFraction && fracPart == "" {
		return fmt.Errorf("amount must be a positive decimal with at least one fractional digit after the decimal point")
	}

	hasNonZero := false
	for _, r := range intPart {
		if r < '0' || r > '9' {
			return fmt.Errorf("amount must be a positive decimal without signs or exponent notation")
		}
		if r != '0' {
			hasNonZero = true
		}
	}
	for _, r := range fracPart {
		if r < '0' || r > '9' {
			return fmt.Errorf("amount must be a positive decimal without signs or exponent notation")
		}
		if r != '0' {
			hasNonZero = true
		}
	}
	if !hasNonZero {
		return fmt.Errorf("amount must be positive")
	}
	if len(strings.TrimRight(fracPart, "0")) > 8 {
		return fmt.Errorf("amount must have at most 8 non-zero fractional digits")
	}
	return nil
}

func validateMarginID(name, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

func validateMarginIdempotencyKey(idempotencyKey string) error {
	if strings.TrimSpace(idempotencyKey) == "" {
		return fmt.Errorf("idempotency key is required")
	}
	return nil
}

// CreateMarginAccount creates a margin account for the authenticated user.
func (c *Client) CreateMarginAccount() (*MarginAccountSummary, error) {
	var summary MarginAccountSummary
	if err := c.postMarginResponseForRoute(apiRouteMarginV2Alpha, "/account", nil, nil, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

// MarginAccountSummary returns the authenticated user's margin account summary.
func (c *Client) MarginAccountSummary() (*MarginAccountSummary, error) {
	var summary MarginAccountSummary
	if err := c.getMarginResponseForRoute(apiRouteMarginV2Alpha, "/account", nil, &summary); err != nil {
		return nil, err
	}
	return &summary, nil
}

// MarginCurrencies returns currencies available for margin trading.
func (c *Client) MarginCurrencies() ([]MarginCurrency, error) {
	res := struct {
		Currencies []MarginCurrency `json:"currencies"`
	}{}
	if err := c.getMarginResponseForRoute(apiRouteMarginV1Alpha, "/currencies", nil, &res); err != nil {
		return nil, err
	}
	return res.Currencies, nil
}

// MarginMovements returns a paginated list of margin funds movements.
func (c *Client) MarginMovements(pagination *MarginPagination) (*MarginMovementList, error) {
	params, err := pagination.values()
	if err != nil {
		return nil, err
	}

	var movements MarginMovementList
	if err := c.getMarginResponseForRoute(apiRouteMarginV1Alpha, "/movements", params, &movements); err != nil {
		return nil, err
	}
	return &movements, nil
}

// MarginMovement returns one margin funds movement by ID.
func (c *Client) MarginMovement(movementID string) (*MarginMovement, error) {
	if err := validateMarginID("movement_id", movementID); err != nil {
		return nil, err
	}

	var movement MarginMovement
	endpoint := "/movements/" + url.PathEscape(movementID)
	if err := c.getMarginResponseForRoute(apiRouteMarginV1Alpha, endpoint, nil, &movement); err != nil {
		return nil, err
	}
	return &movement, nil
}

// ProcessMarginMovement processes a deposit or withdrawal between main and margin balances.
func (c *Client) ProcessMarginMovement(idempotencyKey string, request *MarginMovementRequest) (*MarginMovement, error) {
	if err := validateMarginIdempotencyKey(idempotencyKey); err != nil {
		return nil, err
	}
	if err := request.validate(); err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("X-Idempotency-Key", idempotencyKey)

	var movement MarginMovement
	if err := c.postMarginResponseForRoute(apiRouteMarginV1Alpha, "/movements", request, headers, &movement); err != nil {
		return nil, err
	}
	return &movement, nil
}

// MarginLoans returns a paginated list of loans for the authenticated user's margin account.
func (c *Client) MarginLoans(pagination *MarginPagination) (*MarginLoanList, error) {
	params, err := pagination.values()
	if err != nil {
		return nil, err
	}

	var loans MarginLoanList
	if err := c.getMarginResponseForRoute(apiRouteMarginV1Alpha, "/loans", params, &loans); err != nil {
		return nil, err
	}
	return &loans, nil
}

func (c *Client) getMarginResponseForRoute(route apiRoute, endpoint string, params url.Values, dest interface{}) error {
	return c.doDirectResponseForRoute("GET", route, endpoint, params, nil, dest, decodeMarginErrors)
}

func (c *Client) postMarginResponseForRoute(route apiRoute, endpoint string, body interface{}, headers http.Header, dest interface{}) error {
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(buf)
	}
	return c.doDirectResponseForRouteWithHeaders("POST", route, endpoint, nil, reader, headers, dest, decodeMarginErrors)
}

func decodeMarginErrors(buf []byte) error {
	res := struct {
		Errors MarginErrors `json:"errors"`
	}{}
	if err := json.Unmarshal(buf, &res); err != nil {
		return nil
	}
	if len(res.Errors) > 0 {
		return res.Errors
	}
	return nil
}
