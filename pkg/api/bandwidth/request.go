package bandwidth

type (
	// UsageOptions identifies an IP address (in CIDR form, e.g.
	// "203.0.113.10/32") for the bandwidth usage endpoints. Used by
	// GetUsageByDay, GetUsageByMonth, and GetUsageByYear.
	UsageOptions struct {
		IPAddr string `url:"ip_addr"`
	}
)
