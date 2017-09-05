package bitso

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Book struct {
	major Currency
	minor Currency
}

func (b Book) Major() Currency {
	return b.major
}

func (b Book) Minor() Currency {
	return b.minor
}

func (b Book) MarshalJSON() ([]byte, error) {
	if currencyNames[b.major] == "" {
		return nil, errors.New("fail to encode major")
	}
	if currencyNames[b.minor] == "" {
		return nil, errors.New("fail to encode minor")
	}
	return json.Marshal(fmt.Sprintf("%q", currencyNames[b.major]+"_"+currencyNames[b.minor]))
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
