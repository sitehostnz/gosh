package firewall

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Update function updates the server firewall.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := path.Join(apiPrefix, "update.json")

	keys := []string{
		"client_id",
		"server",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)

	for i, group := range request.SecurityGroups {
		keys = append(keys, fmt.Sprintf("groups[%d]", i))
		values.Add(fmt.Sprintf("groups[%d]", i), group)
	}

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
