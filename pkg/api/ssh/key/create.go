package key

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

// Create an SSH Key.
func (s *Client) Create(ctx context.Context, opts CreateRequest) (response CreateResponse, err error) {
	u := "ssh/key/add.json"

	keys := []string{
		"label",
		"content",
		"params[custom_image_access]",
	}

	values := url.Values{}
	values.Add("label", opts.Label)
	values.Add("content", opts.Content)
	values.Add("params[custom_image_access]", fmt.Sprint(utils.BoolToInt(opts.CustomImageAccess)))

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
