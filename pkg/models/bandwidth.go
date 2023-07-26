package models

type (
	// IPAddress does what it says on the can. It's an IP address.
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
