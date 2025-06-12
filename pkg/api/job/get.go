package job

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

// Get information about a job.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	keys := []string{
		"type",
		"id",
	}

	values := url.Values{}
	values.Add("id", request.ID.String())
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
