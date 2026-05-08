package models

import (
	"encoding/json"
	"strings"
)

type (
	// IPAddress does what it says on the can. It's an IP address.
	//
	// **IPv6 wire-shape quirk + normalisation.**
	// /bandwidth/get_ip_list.json serialises IPv6 addresses with ':'
	// replaced by '.' (and '::' as '..'), e.g.
	// "2403.7000.8000.300..ce/128" rather than
	// "2403:7000:8000:300::ce/128". The mangled form is rejected by
	// every IP-input endpoint, so a naive read-then-write round-trip
	// breaks for IPv6 addresses retrieved from this struct.
	//
	// Custom UnmarshalJSON canonicalises the IP back to standard
	// IPv6 syntax when Family is "6" — '..' → '::' first, then
	// remaining '.' between hextets → ':'. IPv4 (family "4") is
	// passed through untouched. After unmarshalling, the IP field
	// is safe to feed back into any IP-input endpoint.
	//
	// The mangling is localised to /bandwidth/get_ip_list.json —
	// /bandwidth/get_usage_summary.json and /server/list_servers.json
	// both emit canonical IPv6 — but the normaliser is benign on
	// already-canonical strings, so it doesn't matter where the
	// IPAddress originated.
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

// UnmarshalJSON normalises mangled IPv6 IP fields — see type docs.
func (ip *IPAddress) UnmarshalJSON(data []byte) error {
	type alias IPAddress
	var raw alias
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Family == "6" {
		raw.IP = canonicaliseMangledIPv6(raw.IP)
	}
	*ip = IPAddress(raw)
	return nil
}

// canonicaliseMangledIPv6 reverses the dot-for-colon substitution the
// /bandwidth/get_ip_list.json endpoint applies to IPv6 addresses.
// Operates on the address part only; preserves any "/<prefix>"
// suffix verbatim. Already-canonical input (no '.') is returned
// unchanged.
func canonicaliseMangledIPv6(s string) string {
	if s == "" || !strings.Contains(s, ".") {
		return s
	}
	addr, suffix := s, ""
	if i := strings.LastIndex(s, "/"); i >= 0 {
		addr, suffix = s[:i], s[i:]
	}
	// '..' first so '::' isn't subsequently mangled to ':::'.
	addr = strings.ReplaceAll(addr, "..", "::")
	addr = strings.ReplaceAll(addr, ".", ":")
	return addr + suffix
}
