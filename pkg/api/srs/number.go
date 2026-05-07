package srs

import (
	"encoding/json"
	"strconv"
)

// Number tolerates the SRS API's mixed JSON-string / JSON-number
// serialisation of numeric fields. ContactSummary.DomainCount is
// the known case: established contacts return it as a JSON number
// (`0`), newly-created contacts return it as a JSON string (`"0"`),
// within the same shape. Mirrors the bandwidth.Number pattern from
// PR #43.
type Number int64

// UnmarshalJSON accepts either a JSON string or JSON number.
func (n *Number) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*n = Number(v)
		return nil
	}
	return json.Unmarshal(b, (*int64)(n))
}
