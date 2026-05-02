package environment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Delete removes the named environment variables from a stack
// service via "cloud/stack/environment/delete.json". The variables
// to remove are passed by name; their values aren't required.
//
// Same param shape as Update — server, project, service — with
// `variables[N][name]` instead of `variables[N][name]` +
// `variables[N][content]` pairs.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	uri := "cloud/stack/environment/delete.json"
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

	args := make([]string, len(request.Names))
	for i, name := range request.Names {
		k := fmt.Sprintf("variables[%d][name]", i)
		args[i] = k
		values.Add(k, name)
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
