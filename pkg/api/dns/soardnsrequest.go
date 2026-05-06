package dns

// UpdateSOARequest configures the new SOA values for a hosted
// zone. All fields are required by the API.
type UpdateSOARequest struct {
	// Domain is the zone whose SOA to replace (e.g. "example.nz").
	Domain string `json:"domain"`
	// NS is the primary nameserver (e.g. "ns1.sitehost.co.nz").
	NS string `json:"ns"`
	// Email is the SOA contact in @-separated form
	// (e.g. "support@sitehost.co.nz") — the API converts to the
	// dot-encoded BIND form internally.
	Email string `json:"email"`
	// Refresh / Retry / Expire / Minimum are TTL values in seconds.
	Refresh int `json:"refresh"`
	Retry   int `json:"retry"`
	Expire  int `json:"expire"`
	Minimum int `json:"minimum"`
}

// UpdateReverseDNSRequest sets the PTR record for an IP.
type UpdateReverseDNSRequest struct {
	IPAddr string `json:"ip_addr"`
	// RDNS is the desired reverse-DNS hostname
	// (e.g. "192-168-1-105.sitehost.co.nz").
	RDNS string `json:"rdns"`
}

// ResetReverseDNSRequest clears any custom PTR for an IP, returning
// it to the platform default.
type ResetReverseDNSRequest struct {
	IPAddr string `json:"ip_addr"`
}
