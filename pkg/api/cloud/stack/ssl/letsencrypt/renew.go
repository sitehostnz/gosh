package letsencrypt

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Renew forces an early renewal of the LE cert for the named stack.
// The companion auto-renews on schedule (typically when the cert
// has < 30 days remaining); use Renew only for out-of-band refresh.
// Returns a scheduler job.
func (s *Client) Renew(ctx context.Context, request RenewRequest) (response JobResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/renew.json"
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
