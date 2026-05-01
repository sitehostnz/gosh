package bandwidth

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListIPAddressesResponse represents a response from listing IP addresses with the `/bandwidth/get_ip_list.json` endpoint.
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
	UsageResponse struct {
		Return map[string]map[string]map[string]TrafficStats `json:"return"`
		models.APIResponse
	}

	// ResourceQuota is a single quota entry inside a resource group.
	// TotalUnits and UsedUnits are returned as strings; AvailableUnits
	// is returned as a number (and may be negative when over-quota).
	ResourceQuota struct {
		AttributeID    string   `json:"attribute_id"`
		AttributeName  string   `json:"attribute_name"`
		AttributeUnit  string   `json:"attribute_unit"`
		AttributeType  string   `json:"attribute_type"`
		TotalUnits     string   `json:"total_units"`
		UsedUnits      string   `json:"used_units"`
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
