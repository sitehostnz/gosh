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
)
