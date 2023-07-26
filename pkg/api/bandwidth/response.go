package bandwidth

import "github.com/sitehostnz/gosh/pkg/models"

type (
	// ListIPAddressesResponse represents a response from listing IP addresses with the `/bandwidth/get_ip_list.json` endpoint.
	ListIPAddressesResponse struct {
		models.APIResponse
		Return map[string]models.IPAddress `json:"return"`
	}
)
