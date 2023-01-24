package record

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ZoneResponse represents a result of a get DNSZone call.
	ZoneResponse struct {
		DomainRecords *[]models.DNSRecord `json:"return"`
		Status        bool                `json:"status"`
		Message       string              `json:"msg"`
	}
)
