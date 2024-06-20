package image

import (
	"context"
)

// ListImages returns a list of images.
func (s *Client) ListImages(ctx context.Context) (response GetResponse, err error) {
	uri := "cloud/stack/image/list_all.json"

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
