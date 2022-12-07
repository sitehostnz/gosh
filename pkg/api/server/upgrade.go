package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Upgrade a server's plan .
func (s *Client) Upgrade(ctx context.Context, opts UpgradeRequest) (response UpdateResponse, err error) {
	u := "server/upgrade_plan.json"

	keys := []string{
		"client_id",
		"name",
		"plan",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", opts.Name)
	values.Add("plan", opts.Plan)

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
