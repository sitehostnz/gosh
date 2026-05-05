package dns

import (
	"encoding/json"

	"github.com/sitehostnz/gosh/pkg/models"
)

type (
	// IPInfo describes a single IP allocation as returned by
	// list_ips. String-typed numeric / boolean fields ("0"/"1",
	// "32", "128") reflect the API's actual response shape;
	// consumers should convert as needed.
	IPInfo struct {
		IPAddr        string `json:"ip_addr"`
		Netmask       string `json:"netmask"`
		Prefix        string `json:"prefix"`
		Reserved      string `json:"reserved"`
		RDNS          string `json:"rdns"`
		AddrFamily    string `json:"addr_family"`
		DateAllocated string `json:"date_allocated"`
		ServerID      string `json:"server_id"`
		Name          string `json:"name"`
		Label         string `json:"label"`
		IsPrimary     string `json:"is_primary"`
		IPType        string `json:"ip_type"`
	}

	// ListIPsResponse represents the response from list_ips. The
	// Return map is keyed by the IP address (as a string) — the
	// same string is also present as IPAddr inside each entry.
	//
	// **Wire-shape quirk** (verified live): when the account has no
	// allocated IPs, "return" is the JSON array `[]`, not the empty
	// object `{}`. Custom UnmarshalJSON tolerates both forms.
	ListIPsResponse struct {
		Return map[string]IPInfo `json:"return"`
		models.APIResponse
	}
)

// UnmarshalJSON tolerates the empty-array form the API returns when
// the account has no allocated IPs. See type comment.
func (r *ListIPsResponse) UnmarshalJSON(data []byte) error {
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
		r.Return = map[string]IPInfo{}
		return nil
	}
	return json.Unmarshal(raw, &r.Return)
}
