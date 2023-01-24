package zone

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListResponse represents a request to list all DNSZones (domains).
	ListResponse struct {
		Return struct {
			Domains *[]models.DNSZone `json:"data"`
		}
		models.APIResponse
	}

	// CreateResponse represents a request to create a DNSZone (domain).
	CreateResponse struct {
		Return struct {
			IsMigration bool   `json:"is_migration"`
			DomainName  string // add domain name in response
		} `json:"return"`
		models.APIResponse
	}
)
