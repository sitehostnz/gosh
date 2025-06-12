package types //nolint:var-naming

import (
	"encoding/json"
	"strconv"
)

// MaybeBigInt is a simple wrapper unmarshalling json responses that could be either an int or an int in a string.
type MaybeBigInt int64

// UnmarshalJSON is a helper interface for dealing with numbers that may be returned in json responses as either a string with a number in it or numbers.
func (fi *MaybeBigInt) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*int64)(fi))
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*fi = MaybeBigInt(i)
	return nil
}
