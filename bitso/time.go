package bitso

import (
	"encoding/json"
	"time"
)

// Time represents a ISO8601 encoded time value.
type Time time.Time

const (
	iso8601Time = "2006-01-02T15:04:05-0700"
)

func (t *Time) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}
	z, err := time.Parse(iso8601Time, s)
	if err != nil {
		return err
	}
	*t = Time(z)
	return nil
}
