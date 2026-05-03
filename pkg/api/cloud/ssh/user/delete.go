package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete an existing SSH user via cloud/ssh/user/delete.json.
//
// **Two-phase delete required (validated live).** The first call
// queues a job that clears the user's container scoping; the
// second call queues a job that actually removes the user record.
// Calling Delete only once leaves the user record present on the
// CCS and `/cloud/ssh/user/list` keeps returning it.
//
// Calling Delete twice in quick succession is also wrong — the
// API rejects the second call with
// "Unable to remove SSH user, there is a job already running on
// this user." Sleep ~10s between phases (or poll the first
// phase's job to Completed) before issuing the second.
//
// See examples/build-a-site and examples/custom-image cleanup
// loops for the canonical pattern.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	uri := "cloud/ssh/user/delete.json"
	keys := []string{
		"client_id",
		"server_name",
		"username",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("username", request.Username)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
