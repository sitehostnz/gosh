package dns

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// AddRecordResponse is the return for adding a cloud database.
	AddRecordResponse struct {
		Return struct {
			ID string `json:"id"`
		} `json:"return"`
		models.APIResponse
	}

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
	// search_domains returns matching DNSZone summaries; only Name,
	// ClientID, and TemplateID are populated (Pending is returned by
	// list_domains but not by search_domains).
	GetZoneResponse struct {
		Return []models.DNSZone `json:"return"`
		models.APIResponse
	}

	// UpdateSOAResponse is returned by update_soa. Synchronous;
	// only models.APIResponse status fields populate.
	UpdateSOAResponse struct {
		models.APIResponse
	}

	// ReverseDNSResponse is the response shape for the
	// reset_reverse_dns and update_reverse_dns endpoints. Both
	// are synchronous.
	ReverseDNSResponse struct {
		models.APIResponse
	}
)
