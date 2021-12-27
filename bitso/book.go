package bitso

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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
	return fmt.Sprintf("%s_%s", b.major, b.minor)
}

// MarshalJSON implements json.Marshaler.
func (b Book) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

func (b *Book) fromString(s string) error {
	z := strings.Split(s, "_")
	if len(z) != 2 {
		return errors.New("unexpected book format")
	}

	b.major = ToCurrency(z[0])
	b.minor = ToCurrency(z[1])

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
