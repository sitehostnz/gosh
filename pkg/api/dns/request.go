package dns

type (
	// RecordRequest represents a request to get a DNSRecord.
	RecordRequest struct {
		ID         string `json:"id"`
		RRType     string `json:"rr_type"`
		DomainName string `json:"name"`
	}

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

	// AddRecordRequest represents a request to create a DNSRecord.
	AddRecordRequest struct {
		ClientID string `json:"client_id"`
		Domain   string `json:"domain"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Content  string `json:"content"`
		Priority string `json:"prio"`
	}

	// DeleteRecordRequest represents a request to delete a DNSRecord.
	DeleteRecordRequest struct {
		ClientID string `json:"client_id"`
		Domain   string `json:"domain"`
		RecordID string `json:"record_id"`
	}
)
