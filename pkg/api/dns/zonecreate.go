package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// CreateZone a DNSZone.
func (s *Client) CreateZone(ctx context.Context, opts CreateZoneRequest) (response CreateZoneResponse, err error) {
	u := "dns/create_domain.json"

	keys := []string{
		"client_id",
		"domain",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("domain", opts.DomainName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
