package letsencrypt

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Revoke invalidates the LE cert for the named stack at the CA.
// Distinct from Delete: revocation propagates across the public PKI
// and means the cert can never be reinstated — issue a new one if
// you need HTTPS to keep working. Use Delete to remove the cert
// from local config without revoking. Returns a scheduler job.
func (s *Client) Revoke(ctx context.Context, request RevokeRequest) (response JobResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/revoke.json"
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
