package bandwidth

import "context"

// GetUsageSummary retrieves the current-month bandwidth usage
// summary across all IPs allocated to the authenticated client via
// "bandwidth/get_usage_summary.json". Period keys in the response
// are "YYYY-MM" (current month).
func (s *Client) GetUsageSummary(ctx context.Context) (response UsageResponse, err error) {
	u := "bandwidth/get_usage_summary.json"

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
