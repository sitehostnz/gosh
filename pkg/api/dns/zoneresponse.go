package dns

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListZoneResponse represents a request to list all DNSZones (domains).
	ListZoneResponse struct {
		Return struct {
			Domains *[]models.DNSZone `json:"data"`
		}
		models.APIResponse
	}

	// CreateZoneResponse represents a request to create a DNSZone (domain).
	CreateZoneResponse struct {
		Return struct {
			IsMigration bool   `json:"is_migration"`
			DomainName  string // add domain name in response
		} `json:"return"`
		models.APIResponse
	}

	// GetZoneResponse represents a result of a get Server call.
	GetZoneResponse struct {
		Return []struct {
			Name string `json:"name"`
		} `json:"return"`
		models.APIResponse
	}
)
