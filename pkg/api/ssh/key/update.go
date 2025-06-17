package key

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
	"github.com/sitehostnz/gosh/pkg/shtypes"
)

// Update an SSH Key.
func (s *Client) Update(ctx context.Context, opts UpdateRequest) (response UpdateResponse, err error) {
	u := "ssh/key/update.json"

	keys := []string{
		"key_id",
		"params[label]",
		"params[content]",
		"params[custom_image_access]",
	}

	values := url.Values{}
	values.Add("key_id", opts.ID)
	values.Add("params[label]", opts.Label)
	values.Add("params[content]", opts.Content)
	values.Add("params[custom_image_access]", fmt.Sprint(shtypes.BoolToInt(opts.CustomImageAccess)))

	req, err := s.client.NewRequest("POST", u, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
