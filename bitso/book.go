package bitso

import (
	"encoding/json"
	"strings"
)

type Book struct {
	major Currency
	minor Currency
}

func NewBook(major Currency, minor Currency) *Book {
	return &Book{major: major, minor: minor}
}

func (b Book) Major() Currency {
	return b.major
}

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

func (b Book) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.String())
}

func (b *Book) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
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
