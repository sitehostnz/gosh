package token

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add issues a new AI token. The secret is in the response's
// Return.Token and is not retrievable again.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	u := "ai/token/add.json"

	keys := []string{
		"label",
		"capabilities[]",
		"params[allowed_ips][]",
	}

	values := url.Values{}
	values.Add("label", request.Label)
	for _, c := range request.Capabilities {
		values.Add("capabilities[]", c)
	}
	for _, ip := range request.AllowedIPs {
		values.Add("params[allowed_ips][]", ip)
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
