package job

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/net"
	"strconv"
)

// Get information about a job.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "job/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"type",
		"id",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	values := req.URL.Query()
	values.Add("type", request.Type)
	values.Add("id", strconv.Itoa(request.ID))

	req.URL.RawQuery = net.Encode(values, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
