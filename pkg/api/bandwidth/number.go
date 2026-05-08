package bandwidth

import (
	"encoding/json"
	"strconv"
)

// Number tolerates the bandwidth API's mixed JSON-string / JSON-number
// serialisation of numeric fields. Within a single list_resources
// response, used_units comes back as a string (`"783"`) when usage is
// non-zero and a JSON number (`0`) when usage is zero, in the same
// account. Decoding into a fixed Go type breaks for any account with
// at least one zero-used quota — a common shape, not an edge case.
//
// Number accepts either form and stores the value as float64.
type Number float64

// UnmarshalJSON accepts either a JSON string or JSON number.
func (n *Number) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*n = Number(v)
		return nil
	}
	return json.Unmarshal(b, (*float64)(n))
}
