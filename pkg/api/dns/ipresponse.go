package dns

import "github.com/sitehostnz/gosh/pkg/models"

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
	ListIPsResponse struct {
		Return map[string]IPInfo `json:"return"`
		models.APIResponse
	}
)
