package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete a server with the provided name via /server/delete.json.
//
// **Fresh CCSes need `force_delete=1` (verified live, May 2026).**
// Every fresh Cloud Container Server auto-deploys an `infra` stack
// (collectd, nginx-proxy, Let's Encrypt companion). Plain delete
// rejects with "the server has containers" because the infra
// stack is still present. The fix is to add `force_delete=1` to
// the request body, which tears down the infra stack and the
// server in one go.
//
// This base wrapper sends only `client_id` and `name`. Tearing
// down a fresh CCS therefore requires assembling the form body
// directly with `force_delete=1` until a Force field is added to
// DeleteRequest in a follow-up.
//
// **Cannot delete while in 'Upgrading' state.** If a recent
// server.Upgrade (plan upgrade) has just been issued, Delete is
// rejected with "The specified server cannot be deleted while in
// the 'Upgrading' state." Poll server.Get(name) until State is
// On or Off before issuing Delete.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	u := "server/delete.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
