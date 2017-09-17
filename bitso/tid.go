package bitso

import (
	"encoding/json"
	"strconv"
)

// TID respresents a transaction ID
type TID uint64

// UnmarshalJSON implements json.Unmarshaler
func (t *TID) UnmarshalJSON(in []byte) error {
	// This is necessary because Bitso sometimes sends string "tid" values and
	// some other times it feels like sending an integer.
	var v uint64

	// Try to decode the value into an integer.
	err := json.Unmarshal(in, &v)
	if err == nil {
		*t = TID(v)
		return nil
	}

	// Try to decode the value into an string.
	var z string
	if err = json.Unmarshal(in, &z); err != nil {
		return err
	}

	v, err = strconv.ParseUint(z, 10, 64)
	if err != nil {
		return err
	}

	*t = TID(v)
	return nil
}
