package environment

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get returns the stack's environment variables.
func (s *Client) Get(ctx context.Context, request GetRequest) (*[]models.EnvironmentVariable, error) {
	req, err := s.client.NewRequest("GET", "cloud/stack/environment/get.json", "")
	if err != nil {
		return nil, err
	}
	keys := []string{
		"apikey",
		"client_id",
		"server",
		"project",
		"service",
	}

	v := req.URL.Query()
	v.Add("server", request.ServerName)
	v.Add("project", request.Project)
	v.Add("service", request.Service)

	req.URL.RawQuery = utils.Encode(v, keys)

	response := new(GetResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response.EnvironmentVariables, nil
}
