package record

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ZoneResponse struct {
		DomainRecords *[]models.DNSRecord `json:"return"`
		Status        bool                `json:"status"`
		Message       string              `json:"msg"`
	}
)
