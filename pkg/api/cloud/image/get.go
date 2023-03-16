package image

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get fetches a cloud image.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/image/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"code",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("code", request.Code)

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
