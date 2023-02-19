package dns

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListRecordsResponse represents a response from the ListRecords method.
	ListRecordsResponse struct {
		Return []models.DNSRecord `json:"return"`
		models.APIResponse
	}
)
