package environment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update applies updates to the stacks environment.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/stack/environment/update.json"
	keys := []string{
		"client_id",
		"server",
		"project",
		"service",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("project", request.Project)
	values.Add("service", request.Service)

	args := make([]string, len(request.EnvironmentVariables)*2)
	i := 0
	for x, v := range request.EnvironmentVariables {
		args[i] = fmt.Sprintf("variables[%d][name]", x)
		values.Add(args[i], v.Name)
		i++

		if v.Content != "" {
			args[i] = fmt.Sprintf("variables[%d][content]", x)
			values.Add(args[i], v.Content)
			i++
		}
	}

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, append(keys, args...)))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
