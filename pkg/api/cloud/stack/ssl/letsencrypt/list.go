package letsencrypt

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// List retrieves the LE certificates currently provisioned across
// the stacks on the given Cloud Container Server. Returns a map
// keyed by stack name. Stacks without an LE cert are absent from
// the map.
func (s *Client) List(ctx context.Context, request ListRequest) (response ListResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/list_all.json"
	keys := []string{"apikey", "client_id", "server"}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}

