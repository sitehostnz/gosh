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
// stack is still present. Set DeleteRequest.Force to true to add
// `force_delete=1` to the request body, which tears down the
// infra stack and the server in one go.
//
// # Deletable states
//
// A server can only be deleted while it is **On**, **Off**, or in an
// **Unknown** state. Any other state is rejected with:
//
//	The specified server cannot be deleted while in the '<state>' state.
//
// Transitional states seen in practice include 'Provisioning' (after a
// build), 'Shutting Down' (after a reboot) and 'Upgrading' (after a plan
// change). The set is not documented upstream, so match on the shape of
// that message rather than on a list of state names — treating an
// unlisted state as a permanent failure leaks a server, which is the
// more expensive mistake.
//
// The rejection is a pre-flight check: nothing is queued when it fires,
// so the server still exists and a retry is required. Poll
// server.GetState until the state settles, then delete. Note that the
// reported state can lag reality — a server observed answering SSH was
// refused a moment later as 'Shutting Down'.
//
// A server may also be refused because it is locked, missing, in rescue
// mode, or has an active snapshot; those are not transitional and
// retrying will not help.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	u := "server/delete.json"

	keys := []string{
		"client_id",
		"name",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("name", request.Name)
	if request.Force {
		values.Add("force_delete", "1")
		keys = append(keys, "force_delete")
	}

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
