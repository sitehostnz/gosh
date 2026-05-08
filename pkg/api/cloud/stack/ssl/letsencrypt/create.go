package letsencrypt

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Create queues a Let's Encrypt cert issuance for the named stack.
// The HTTP-01 challenge runs against the stack's vhost via
// nginx-proxy; the stack must be reachable on port 80 at its label
// hostname for the challenge to succeed.
//
// Returns a scheduler job; consumers must poll job.Get until
// state="Completed" before assuming the cert is issued.
func (s *Client) Create(ctx context.Context, request CreateRequest) (response JobResponse, err error) {
	uri := "cloud/stack/ssl/lets_encrypt/create.json"
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
