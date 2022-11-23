package job

import (
	"context"
	"fmt"
)

// Get gets the GetResponse with the provided ID.
func (s *Client) Get(ctx context.Context, request GetRequest) (*GetResponse, error) {
	u := fmt.Sprintf("job/get.json?job_id=%s&type=%s", request.JobID, request.Type)
	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return nil, err
	}

	response := new(GetResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
