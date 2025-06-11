package job

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get information about a job.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	keys := []string{
		"id",
		"type",
	}

	values := url.Values{}
	values.Add("id", request.JobID)
	values.Add("type", request.Type)

	u := "job/get.json?" + utils.Encode(values, keys)

	req, err := s.client.NewRequest("GET", u, "")
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
