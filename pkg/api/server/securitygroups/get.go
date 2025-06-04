package securitygroups

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get retrieves details of a security group including its rules and attached servers.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "server/firewall/security_groups/get.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("GET", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
