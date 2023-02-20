package environment

import (
	"context"
	"fmt"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

// Update applies updates to the stacks environment
func (s *Client) Update(ctx context.Context, request UpdateRequest) (*UpdateResponse, error) {

	u := "cloud/stack/environment/update.json"

	keys := []string{
		"client_id",
		"server",
		"project",
		"service",
	}

	values := url.Values{}
	// not sure as to the best way to do this, we are we likely to move records between clients?
	// let's just use the one specified in the config for now...
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("project", request.Project)
	values.Add("service", request.Service)

	args := make([]string, len(*request.EnvironmentVariables)*2)
	i := 0
	for x, v := range *request.EnvironmentVariables {
		args[i] = fmt.Sprintf("variables[%d][name]", x)
		values.Add(args[i], v.Name)

		args[i+1] = fmt.Sprintf("variables[%d][content]", x)
		values.Add(args[i+1], v.Content)
		i += 2
	}

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, append(keys, args...)))
	if err != nil {
		return nil, err
	}

	response := new(UpdateResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
