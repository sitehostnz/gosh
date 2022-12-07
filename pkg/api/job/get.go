package job

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get information about a job.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	u := "job/get.json"

	keys := []string{
		"job_id",
		"type",
	}

	values := url.Values{}
	values.Add("job_id", request.JobID)
	values.Add("type", request.Type)

	req, err := s.client.NewRequest("GET", u, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
