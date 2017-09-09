package bitso

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Side uint8

const (
	Buy Side = iota
	Sell
)

func (s Side) MarshalJSON() ([]byte, error) {
	z := s.String()
	return []byte(fmt.Sprintf("%q", z)), errors.New("could not encode side")
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

func (s *Side) String() string {
	switch *s {
	case Buy:
		return "buy"
	case Sell:
		return "sell"
	}
	panic("reached")
}
