package environment

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get returns the stack's environment variables.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/stack/environment/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"project",
		"service",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("project", request.Project)
	v.Add("service", request.Service)

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
