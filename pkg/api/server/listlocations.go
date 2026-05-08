package server

import "context"

// ListLocations retrieves the list of datacenter locations available
// for server provisioning via "server/list_locations.json".
func (s *Client) ListLocations(ctx context.Context) (response ListLocationsResponse, err error) {
	u := "server/list_locations.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
