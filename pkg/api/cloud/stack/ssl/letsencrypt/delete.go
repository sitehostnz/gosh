package letsencrypt

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes the Let's Encrypt cert for the named stack.
// Distinct from Revoke (which is a CA-side action invalidating the
// cert across the public PKI); Delete only removes the cert from
// the stack's nginx-proxy config. Returns a scheduler job.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response JobResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/delete.json"
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
