package environment

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/models"
	"github.com/sitehostnz/gosh/pkg/utils"
)

func (s *Client) Get(ctx context.Context, request *GetRequest) (*[]models.EnvironmentVariable, error) {

	u := "cloud/stack/environment/get.json"
	req, err := s.client.NewRequest("GET", u, "")
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
