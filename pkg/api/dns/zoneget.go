package dns

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/net"
	"net/url"
)

func (s *Client) GetZone(ctx context.Context, request GetZoneRequest) (response GetZoneResponse, err error) {

	resp, err := s.SearchZone(ctx, SearchZoneRequest{
		Query: request.DomainName,
		Limit: 1,
	})

	if err != nil {
		return response, err
	}

	response.APIResponse = resp.APIResponse

	// iterate over the domains to find the one we are looking for.
	// Since we've set a limit ths should be one...
	if len(resp.Return) > 0 {
		response.Return = resp.Return[0]
	}
	return response, nil
}

// SearchZone searches for a domain in the DNS service
// by sending a request with a query containing the domain name.
// Matching domains are included in the SearchZoneResponse.
func (s *Client) SearchZone(ctx context.Context, request SearchZoneRequest) (response SearchZoneResponse, err error) {
	u := "dns/search_domains.json"

	keys := []string{
		"client_id",
		"query[domain]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("query[domain]", request.Query)

	// would like to do this, but it's kinda not... working
	//if request.Offset > 0 {
	//	keys = append(keys, "offset[offset]")
	//	values.Add("offset[offset]", strconv.Itoa(request.Offset))
	//}
	//
	//if request.Limit > 0 {
	//	keys = append(keys, "offset[limit]")
	//	values.Add("offset[limit]", strconv.Itoa(request.Limit))
	//}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
