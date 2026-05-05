package bandwidth

import (
	"encoding/json"

	"github.com/sitehostnz/gosh/pkg/models"
)

type (
	// ListIPAddressesResponse represents a response from listing IP addresses with the `/bandwidth/get_ip_list.json` endpoint.
	//
	// **Wire-shape quirk** (verified live): when the account has no
	// allocated IPs, "return" is the JSON array `[]`, not the empty
	// object `{}`. Custom UnmarshalJSON tolerates both forms.
	ListIPAddressesResponse struct {
		models.APIResponse
		Return map[string]models.IPAddress `json:"return"`
	}

	// TrafficStats is a single traffic class's peak/offpeak in/out
	// counters in MB. Used inside the usage endpoints' nested map.
	TrafficStats struct {
		PeakIn     float64 `json:"peak_in"`
		OffpeakIn  float64 `json:"offpeak_in"`
		PeakOut    float64 `json:"peak_out"`
		OffpeakOut float64 `json:"offpeak_out"`
	}

	// UsageResponse is the shared response shape for the usage
	// endpoints (get_usage_summary, get_usage_by_day, by_month,
	// by_year). The Return map nests three levels:
	//   ip-CIDR  →  period-key  →  traffic-class  →  TrafficStats
	// Where:
	//   - ip-CIDR is the address with prefix (e.g. "203.0.113.10/32")
	//   - period-key format depends on the endpoint:
	//       summary  → "YYYY-MM" (current month)
	//       by_day   → "YYYY-MM-DD"
	//       by_month → "YYYY-MM"
	//       by_year  → "YYYY"
	//   - traffic-class is "DOMESTIC" or "INTERNATIONAL"
	//
	// **Wire-shape quirk** (verified live): when the account has no
	// bandwidth history (or no rows in the queried window), "return"
	// is the JSON array `[]`, not the empty object `{}`. Custom
	// UnmarshalJSON tolerates both forms.
	UsageResponse struct {
		Return map[string]map[string]map[string]TrafficStats `json:"return"`
		models.APIResponse
	}

	// ResourceQuota is a single quota entry inside a resource group.
	// AvailableUnits is returned as a number (and may be negative
	// when over-quota).
	//
	// TotalUnits and UsedUnits use [Number] because the API mixes
	// JSON-string and JSON-number forms within a single response —
	// see the [Number] type documentation.
	ResourceQuota struct {
		AttributeID    string   `json:"attribute_id"`
		AttributeName  string   `json:"attribute_name"`
		AttributeUnit  string   `json:"attribute_unit"`
		AttributeType  string   `json:"attribute_type"`
		TotalUnits     Number   `json:"total_units"`
		UsedUnits      Number   `json:"used_units"`
		AvailableUnits int      `json:"available_units"`
		Objects        []string `json:"objects"`
	}

	// ResourceGroup represents a per-client resource quota group.
	ResourceGroup struct {
		ClientID  string          `json:"client_id"`
		GroupID   string          `json:"group_id"`
		GroupName string          `json:"group_name"`
		Quotas    []ResourceQuota `json:"quotas"`
	}

	// ListResourcesResponse represents the response from list_resources.
	ListResourcesResponse struct {
		Return []ResourceGroup `json:"return"`
		models.APIResponse
	}
)

// UnmarshalJSON tolerates the empty-array form the API returns when
// the account has no allocated IPs. See type comment.
func (r *ListIPAddressesResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse

	raw := envelope.Return
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		r.Return = map[string]models.IPAddress{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}

// UnmarshalJSON tolerates the empty-array form the API returns when
// the account has no bandwidth history in the queried window. See
// type comment.
func (r *UsageResponse) UnmarshalJSON(data []byte) error {
	var envelope struct {
		Return json.RawMessage `json:"return"`
		models.APIResponse
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}
	r.APIResponse = envelope.APIResponse

	raw := envelope.Return
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" {
		r.Return = map[string]map[string]map[string]TrafficStats{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}
