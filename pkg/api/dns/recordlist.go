package dns

import (
	"context"
	"fmt"
)

// ListRecords returns a list of DNS records for the specified domain.
// This function takes a context.Context and a ListRecordsRequest struct as input parameters.
// It returns a ListRecordsResponse struct and an error.
func (s *Client) ListRecords(ctx context.Context, request ListRecordsRequest) (response ListRecordsResponse, err error) {
	u := fmt.Sprintf("dns/list_records.json?domain=%v", request.Domain)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	// add back the domain name, since we have it in the model, but not in the response
	// or at least according to the api docs
	if response.Status {
		for k := range response.Return {
			response.Return[k].Domain = request.Domain
		}
	}

	return response, nil
}
