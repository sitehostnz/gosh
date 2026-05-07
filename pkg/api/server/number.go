package server

import (
	"encoding/json"
	"strconv"
)

// Number tolerates the server API's mixed JSON-string / JSON-number
// serialisation of numeric fields. ResourceQuota.UsedUnits and
// TotalUnits are the known cases: when usage is non-zero the value
// arrives as a JSON string (`"783"`), when usage is zero it arrives
// as a JSON number (`0`), within the same response. Mirrors
// bandwidth.Number from PR #43, which surfaced the same quirk on
// the parallel /bandwidth/list_resources endpoint.
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
