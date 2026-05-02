package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// UpdateMinimumTLSVersion sets the minimum TLS version the named
// CCS's nginx-proxy will negotiate via
// "cloud/server/update_minimum_tls_version.json". Returns a
// scheduler job.
//
// MinimumTLSVersion must be in the "TLSv1.x" format
// (e.g. "TLSv1.1", "TLSv1.2", "TLSv1.3"); other formats like "1.2"
// or "TLS_1_2" are rejected with "The minimum tls version is not
// valid."
//
// **There's no corresponding read endpoint.** Once set, the value
// is observable only via TLS handshake against a service running on
// the CCS, or by tracking the value in your own state. See
// examples/probe-tls-default for the handshake-observation
// approach.
//
// **Affects all stacks on the CCS.** This is a server-wide setting
// that controls TLS termination at nginx-proxy for every customer
// site running on the CCS. Tightening from the platform default
// (TLSv1.1) will reject older TLS clients across all customer sites.
func (s *Client) UpdateMinimumTLSVersion(ctx context.Context, request UpdateMinimumTLSVersionRequest) (response JobResponse, err error) {
	u := "cloud/server/update_minimum_tls_version.json"
	keys := []string{"client_id", "server_name", "minimum_tls_version"}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("minimum_tls_version", request.MinimumTLSVersion)

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
