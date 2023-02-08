package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// GetZone information about the DNSZone.
func (s *Client) GetZone(ctx context.Context, request GetZoneRequest) (response GetZoneResponse, err error) {
	u := "dns/search_domains.json"

	keys := []string{
		"client_id",
		"query[domain]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("query[domain]", request.DomainName)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	err = s.client.Do(ctx, req, &response)
	if err != nil {
		return response, err
	}

	// TODO add control for empty response

	// we're assuming that when we ask for a domain, the list here should be unique and we get one or none
	return response, nil
}
