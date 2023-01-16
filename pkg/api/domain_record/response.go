package domain_record

import "github.com/sitehostnz/gosh/pkg/models"

type (
	ZoneResponse struct {
		DomainRecords *[]models.DomainRecord `json:"return"`
		Status        bool                   `json:"status"`
		Message       string                 `json:"msg"`
	}
)
