package record

type (
	// RecordRequest represents a request to get a DNSRecord.
	RecordRequest struct {
		ID         string `json:"id"`
		RRType     string `json:"rr_type"`
		DomainName string `json:"name"`
	}

	// ZoneRequest represents a request to get a DNSZone.
	ZoneRequest struct {
		DomainName string `json:"name"`
	}

	// CreateRequest represents a request to create a DNSRecord.
	CreateRequest struct {
		ClientID string `json:"client_id"`
		Domain   string `json:"domain"`
		Type     string `json:"type"`
		Name     string `json:"name"`
		Content  string `json:"content"`
		Priority string `json:"prio"`
	}

	// DeleteRequest represents a request to delete a DNSRecord.
	DeleteRequest struct {
		ClientID string `json:"client_id"`
		Domain   string `json:"domain"`
		RecordID string `json:"record_id"`
	}
)
