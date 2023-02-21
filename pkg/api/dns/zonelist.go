package dns

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// ListZones retrieves a list of all DNS zones associated with the client's account,
// optionally filtering by specified parameters in ListZoneOptions.
// It returns a pointer to a slice of DNSZone objects and an error if any.
func (s *Client) ListZones(ctx context.Context, opt *ListZoneOptions) (response ListZoneResponse, err error) {
	u := "dns/list_domains.json"

	path, err := utils.AddOptions(u, opt)
	if err != nil {
		return response, err
	}

	req, err := s.client.NewRequest("GET", path, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
