package record

type (
	RecordRequest struct {
		Id         string `json:"id"`
		RRType     string `json:"rr_type"`
		DomainName string `json:"name"`
	}

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

	// DeleteRequest represents a request to delete a Server.
	DeleteRequest struct {
		ClientID string `json:"client_id"`
		Domain   string `json:"domain"`
		RecordID string `json:"record_id"`
	}
)
