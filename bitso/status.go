package bitso

import (
	"encoding/json"
	"fmt"
)

type Status uint8

const (
	StatusNone Status = iota

	StatusOpen
	StatusPartialFill
)

var statusNames = map[Status]string{
	StatusOpen:        "open",
	StatusPartialFill: "partial-fill",
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(in []byte) error {
	var z string
	if err := json.Unmarshal(in, &z); err != nil {
		return err
	}
	for k, v := range statusNames {
		if z == v {
			*s = k
			return nil
		}
	}
	return fmt.Errorf("unsupported status: %v", z)
}

func (s *Status) String() string {
	if z, ok := statusNames[*s]; ok {
		return z
	}
	panic("unsupported status")
}
