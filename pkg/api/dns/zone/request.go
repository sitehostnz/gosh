package zone

type (
	// GetRequest represents a request to get a DNSZone.
	GetRequest struct {
		DomainName string `json:"name"`
	}

	// CreateRequest represents a request to create a DNSZone (domain).
	CreateRequest struct {
		DomainName string `json:"name"`
	}

	// DeleteRequest represents a request to delete a DNSZone.
	DeleteRequest struct {
		DomainName string `json:"name"`
	}
)
