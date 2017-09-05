package bitso

import (
	"encoding/json"
	"errors"
)

type Side uint8

const (
	Buy Side = iota
	Sell
)

func (s Side) MarshalJSON() ([]byte, error) {
	switch s {
	case Buy:
		return []byte(`"buy"`), nil
	case Sell:
		return []byte(`"sell"`), nil
	}
	return nil, errors.New("could not encode side")
}

func (s *Side) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	switch z {
	case "buy":
		*s = Buy
		return nil
	case "sell":
		*s = Sell
		return nil
	}
	return errors.New("could not decode side")
}
