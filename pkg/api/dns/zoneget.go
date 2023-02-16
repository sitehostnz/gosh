package dns

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// GetZone searches for a domain in the DNS made easy service
// by sending a request with a query containing the domain name.
// If the response contains any matching domain, it returns it as part of the GetZoneResponse.
// If the response is empty, a control to handle this case is missing and should be added.
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

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	// TODO add control for empty response

	return response, nil
}
