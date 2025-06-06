package security_groups

import (
	"context"
	"fmt"
)

// Get retrieves details of a security group including its rules and attached servers.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := fmt.Sprintf("server/firewall/security_groups/get.json?name=%v", request.Name)

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
