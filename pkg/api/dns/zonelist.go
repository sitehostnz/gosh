package dns

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// ListZones information about the all DNSZone in the account.
func (s *Client) ListZones(ctx context.Context, opt *ListZoneOptions) (*[]models.DNSZone, error) {
	u := "dns/list_domains.json"

	path, err := utils.AddOptions(u, opt)
	if err != nil {
		return nil, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return nil, err
	}

	response := new(ListZoneResponse)
	if err := s.client.Do(ctx, req, response); err != nil {
		return nil, err
	}

	return response.Return.Domains, err
}
