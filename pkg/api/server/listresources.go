package server

import "context"

// ListResources retrieves the per-client resource quota groups via
// "server/list_resources.json". Each group contains one or more
// quotas (e.g. VPS Disk Space, VPS Memory) with total / used /
// available unit counts and the list of objects (servers) consuming
// each quota.
func (s *Client) ListResources(ctx context.Context) (response ListResourcesResponse, err error) {
	u := "server/list_resources.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
