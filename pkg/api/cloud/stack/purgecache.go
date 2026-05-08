package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// PurgeCache clears cached content from a cloud stack via
// "cloud/stack/purge_cache.json". ServerName and Name are required.
// Returns a scheduler job id.
//
// Non-destructive: only the stack's edge cache is cleared; the
// stack itself, its configuration, and its data are untouched.
func (s *Client) PurgeCache(ctx context.Context, request PurgeCacheRequest) (response JobResponse, err error) {
	uri := "cloud/stack/purge_cache.json"
	keys := []string{"client_id", "server", "name"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
