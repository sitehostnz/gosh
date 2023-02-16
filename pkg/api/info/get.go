package info

import (
	"context"
)

// Get retrieves information about the current user's account using the "api/get_info.json" API endpoint.
// It creates a new HTTP GET request with the client's API key and sends it using the HTTP client.
// The function returns a GetResponse struct containing the API response data and an error if one occurred during the request.
func (s *Client) Get(ctx context.Context) (response GetResponse, err error) {
	req, err := s.client.NewRequest("GET", "api/get_info.json", "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
