package models

type (
	// IPAddress does what it says on the can. It's an IP address.
	//
	// Caveat for IPv6: when this struct is populated from
	// /bandwidth/get_ip_list.json the IP field comes back with ':'
	// replaced by '.' (and '::' as '..'), e.g.
	// "2403.7000.8000.300..ce/128" instead of
	// "2403:7000:8000:300::ce/128". The mangled form is rejected by
	// any IP-input endpoint, so callers round-tripping IPv6 addresses
	// from this field need to canonicalise them first. Localised to
	// /bandwidth/get_ip_list.json — /bandwidth/get_usage_summary.json
	// and /server/list_servers.json both emit canonical IPv6 in the
	// same response.
	IPAddress struct {
		IP            string `json:"ip_addr"`
		Netmask       string `json:"netmask"`
		Prefix        string `json:"prefix"`
		Reserved      string `json:"reserved"`
		RDNS          string `json:"rdns"`
		Family        string `json:"addr_family"`
		DateAllocated string `json:"date_allocated"`
	}
)
