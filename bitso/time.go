package bitso

import (
	"encoding/json"
	"time"
)

// Time represents a ISO8601 encoded time value.
type Time time.Time

const (
	iso8601Time     = "2006-01-02T15:04:05-0700"
	iso8601TimeNano = "2006-01-02T15:04:05.000-07:00"
)

var timeFormats = []string{
	iso8601Time,
	iso8601TimeNano,
}

// UnmarshalJSON implements json.Unmarshal
func (t *Time) UnmarshalJSON(in []byte) error {
	var s string
	if err := json.Unmarshal(in, &s); err != nil {
		return err
	}

	var err error
	var z time.Time
	for _, timeFormat := range timeFormats {
		z, err = time.Parse(timeFormat, s)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	*t = Time(z)
	return nil
}

// String implements fmt.Stringer
func (t Time) String() string {
	return time.Time(t).Format(iso8601Time)
}
