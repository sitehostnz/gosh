package dns

type (
	// RecordRequest represents a request to get a DNSRecord.
	RecordRequest struct {
		ID         string `json:"id"`
		RRType     string `json:"rr_type"`
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
