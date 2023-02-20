package stack

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

// Create creates a new cloud stack.
func (s *Client) Create(ctx context.Context, request CreateRequest) (*CreateResponse, error) {

	u := "cloud/stack/add.json"

	keys := []string{
		"client_id",
		"server",
		"name",
		"label",
		"enable_ssl",
		"docker_compose",
	}

	values := url.Values{}
	// not sure as to the best way to do this, we are we likely to move records between clients?
	// let's just use the one specified in the config for now...
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	// values.Add("project", request.Project)
	// values.Add("service", request.Service)

	args := make([]string, len(*request.Environments))
	// i := 0
	// for x, v := range *request.EnvironmentVariables {
	// 	args[i] = fmt.Sprintf("variables[%d][name]", x)
	// 	values.Add(args[i], v.Name)
	//
	// 	args[i+1] = fmt.Sprintf("variables[%d][content]", x)
	// 	values.Add(args[i+1], v.Content)
	// 	i += 2
	// }

	req, err := s.client.NewRequest("POST", u, utils.Encode(values, append(keys, args...)))
	if err != nil {
		return nil, err
	}

	response := new(CreateResponse)
	err = s.client.Do(ctx, req, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
