package dns

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListZoneResponse represents a request to list all DNSZones (domains).
	ListZoneResponse struct {
		Return struct {
			models.Pagination
			Data []models.DNSZone `json:"data"`
		} `json:"return"`
		models.APIResponse
	}

	// CreateZoneResponse represents a request to create a DNSZone (domain).
	CreateZoneResponse struct {
		Return struct {
			IsMigration bool `json:"is_migration"`
		} `json:"return"`
		models.APIResponse
	}

	// GetZoneResponse represents a request to get a DNSZone (domain).
	GetZoneResponse struct {
		Return []struct {
			Name string `json:"name"`
		} `json:"return"`
		models.APIResponse
	}
)
