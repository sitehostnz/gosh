package dns

import (
	"context"
	"fmt"
)

// ListRecords returns the records for a given domain.
func (s *Client) ListRecords(ctx context.Context, request ListRecordsRequest) (response ListRecordsResponse, err error) {
	u := fmt.Sprintf("dns/list_records.json?domain=%v", request.Domain)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
