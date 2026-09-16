package bitso

import (
	"fmt"
	"net/url"
	"strings"
)

// OrderModification represents the modifiable properties for a v4 order.
type OrderModification struct {
	Major  Monetary `json:"major,omitempty"`
	Minor  Monetary `json:"minor,omitempty"`
	Price  Monetary `json:"price,omitempty"`
	Stop   Monetary `json:"stop,omitempty"`
	Cancel bool     `json:"cancel,omitempty"`
}

func (m *OrderModification) validate() error {
	if m == nil {
		return fmt.Errorf("order modification is required")
	}

	hasMajor := m.Major != ""
	hasMinor := m.Minor != ""
	hasPrice := m.Price != ""
	hasStop := m.Stop != ""

	if hasMajor && hasMinor {
		return fmt.Errorf("major and minor cannot both be set")
	}
	if !hasMajor && !hasMinor && !hasPrice && !hasStop {
		if m.Cancel {
			return fmt.Errorf("cancel cannot be sent alone")
		}
		return fmt.Errorf("order modification requires at least one of major, minor, price, or stop")
	}
	return nil
}

func validateOrderModificationLocator(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", name)
	}
	return nil
}

// ModifyOrder modifies an order identified by its Bitso-supplied order ID path parameter.
func (c *Client) ModifyOrder(oid string, modification *OrderModification) (string, error) {
	if err := validateOrderModificationLocator("oid", oid); err != nil {
		return "", err
	}
	return c.modifyOrder("/orders/"+url.PathEscape(oid), nil, modification)
}

// ModifyOrderByQueryOID modifies an order identified by its Bitso-supplied order ID query parameter.
func (c *Client) ModifyOrderByQueryOID(oid string, modification *OrderModification) (string, error) {
	if err := validateOrderModificationLocator("oid", oid); err != nil {
		return "", err
	}
	params := url.Values{"oid": {oid}}
	return c.modifyOrder("/orders", params, modification)
}

// ModifyOrderByOriginID modifies an order identified by its client-supplied origin ID query parameter.
func (c *Client) ModifyOrderByOriginID(originID string, modification *OrderModification) (string, error) {
	if err := validateOrderModificationLocator("origin_id", originID); err != nil {
		return "", err
	}
	params := url.Values{"origin_id": {originID}}
	return c.modifyOrder("/orders", params, modification)
}

func (c *Client) modifyOrder(endpoint string, params url.Values, modification *OrderModification) (string, error) {
	if err := modification.validate(); err != nil {
		return "", err
	}

	res := struct {
		Payload struct {
			OID string `json:"oid"`
		} `json:"payload"`
	}{}
	if err := c.patchResponseForRoute(apiRouteV4, endpoint, params, modification, &res); err != nil {
		return "", err
	}
	return res.Payload.OID, nil
}
