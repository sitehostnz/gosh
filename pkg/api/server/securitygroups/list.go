package securitygroups

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// List retrieves all security groups with optional filtering.
func (s *Client) List(ctx context.Context, request ListAllRequest) (response ListResponse, err error) {
	uri := "server/firewall/security_groups/list_all.json"

	path, err := utils.AddOptions(uri, request)
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
