package stack

import (
	"context"
	"fmt"
)

// List returns cloud stacks, via "cloud/stack/list_all.json".
//
// ServerName is optional — omitted, this returns every stack on the
// account — but a name that does not resolve is rejected with "Please
// specify a valid server name filter.", not answered with an empty
// page.
//
// Note the two identifiers a stack carries. Name is what every other
// endpoint wants ("cc0123456789abcd", or a chosen one like "infra");
// Label is the display name and is frequently empty. Do not pass a
// label where a name is required.
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
