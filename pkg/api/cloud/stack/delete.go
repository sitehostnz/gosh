package stack

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes a cloud stack via "cloud/stack/delete.json". Both
// ServerName and Name are required. Returns a scheduler job id; the
// operation is asynchronous.
//
// Destructive: removes the stack and its configuration. The
// containers themselves are stopped and removed as part of the job.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response JobResponse, err error) {
	uri := "cloud/stack/delete.json"
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
