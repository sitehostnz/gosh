package securitygroups

import (
	"context"
	"net/url"
	"path"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new security group.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := path.Join(apiPrefix, "add.json")

	keys := []string{
		"client_id",
		"label",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", request.Label)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	// Get the security group name by listing and filtering by label
	listResponse, err := s.List(ctx, ListAllRequest{
		Label: request.Label,
	})
	if err != nil {
		return response, err
	}

	// Find the security group with matching label
	for _, group := range listResponse.Return.Data {
		if group.Label == request.Label {
			response.Return.Name = group.Name
			break
		}
	}

	return response, nil
}
