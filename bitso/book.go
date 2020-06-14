package bitso

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

// Book represents an exchange book.
type Book struct {
	major Currency
	minor Currency
}

// NewBook returns a new exchange book from the given currencies.
func NewBook(major Currency, minor Currency) *Book {
	return &Book{major: major, minor: minor}
}

// Major returns the major currency in the book.
func (b Book) Major() Currency {
	return b.major
}

// Minor returns the minor currency in the book.
func (b Book) Minor() Currency {
	return b.minor
}

func (b Book) String() string {
	if currencyNames[b.major] == "" {
		panic("missing major")
	}
	if currencyNames[b.minor] == "" {
		panic("missing minor")
	}
	return currencyNames[b.major] + "_" + currencyNames[b.minor]
}

// MarshalJSON implements json.Marshaler.
func (b Book) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

func (b *Book) fromString(s string) error {
	z := strings.Split(s, "_")

	major, err := getCurrencyByName(z[0])
	if err != nil {
		return err
	}
	b.major = *major

	minor, err := getCurrencyByName(z[1])
	if err != nil {
		return err
	}
	b.minor = *minor

	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Book) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}

	return b.fromString(s)
}

func (b Book) Value() (driver.Value, error) {
	return b.String(), nil
}

func (b *Book) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	return b.fromString(value.(string))
}
