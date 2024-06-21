package key

import (
	"context"
	"fmt"
)

// Get information about the SSH Key.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	u := fmt.Sprintf("ssh/key/get.json?key_id=%v", request.ID)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
