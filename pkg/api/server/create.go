package server

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Create a server.
func (s *Client) Create(ctx context.Context, opts CreateRequest) (response CreateResponse, err error) {
	u := "server/provision.json"

	keys := []string{
		"client_id",
		"label",
		"location",
		"product_code",
		"image",
		"params[ipv4]",
		"params[ssh_keys][]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("label", opts.Label)
	values.Add("location", opts.Location)
	values.Add("product_code", opts.ProductCode)
	values.Add("image", opts.Image)
	values.Add("params[ipv4]", "auto")

	if len(opts.Params.SSHKeys) > 0 {
		for _, key := range opts.Params.SSHKeys {
			values.Add("params[ssh_keys][]", key)
		}
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
