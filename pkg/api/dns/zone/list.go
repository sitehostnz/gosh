package zone

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
)

// Get information about the all DNSZone in the account.
func (s *Client) List(ctx context.Context) (*[]models.DNSZone, error) {
	u := "dns/list_domains.json"
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(ListResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	// if the request works, don't care about pagination right now...
	return response.Return.Domains, err
}
