package models

type (
	//Subnet struct {
	//	IpAddr  bool   `json:"ip_addr"`
	//	Netmask string `json:"netmask"`
	//	Prefix  int    `json:"prefix"`
	//}

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
