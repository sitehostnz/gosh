package stack

import (
	"context"
	"fmt"
)

// List fetches all cloud stacks on a specific server.
func (s *Client) List(ctx context.Context, request ListRequest) (response ListResponse, err error) {
	uri := "cloud/stack/list_all.json"

	if request.ServerName != "" {
		uri += fmt.Sprintf("?filters[server_name]=%s", request.ServerName)
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
