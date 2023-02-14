package dns

type (
	// GetZoneRequest represents a request to get a DNSZone.
	GetZoneRequest struct {
		DomainName string `json:"name"`
	}

	// CreateZoneRequest represents a request to create a DNSZone (domain).
	CreateZoneRequest struct {
		DomainName string `json:"name"`
	}

	// DeleteZoneRequest represents a request to delete a DNSZone.
	DeleteZoneRequest struct {
		DomainName string `json:"name"`
	}

	// ListZoneRequest represents a request to list DNSZones.
	ListZoneOptions struct {
		Domain     string `url:"filters[domain],omitempty"`
		SortBy     string `url:"filters[sort_by],omitempty"`
		SortDir    string `url:"filters[sort_dir],omitempty"`
		PageSize   int    `url:"filters[page_size],omitempty"`
		PageNumber int    `url:"filters[page_number],omitempty"`
	}
)
