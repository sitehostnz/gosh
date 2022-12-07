package job

import (
	"context"
	"fmt"
)

// Get information about a job.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	u := fmt.Sprintf("job/get.json?job_id=%s&type=%s", request.JobID, request.Type)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
