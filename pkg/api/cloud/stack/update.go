package stack

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update modifies an existing cloud stack via "cloud/stack/update.json".
// ServerName and Name are required. Label, EnableSSL, DockerCompose,
// and EnvironmentVariables are sent when non-zero / non-empty; the
// caller can leave any of them at zero to skip updating that field.
//
// Returns a scheduler job id; the operation is asynchronous.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response JobResponse, err error) {
	uri := "cloud/stack/update.json"
	keys := []string{
		"client_id",
		"server",
		"name",
		"label",
		"enable_ssl",
		"docker_compose",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server", request.ServerName)
	values.Add("name", request.Name)
	if request.Label != "" {
		values.Add("label", request.Label)
	}
	values.Add("enable_ssl", strconv.Itoa(request.EnableSSL))
	if request.DockerCompose != "" {
		values.Add("docker_compose", request.DockerCompose)
	}

	var vars string
	for _, envVar := range request.EnvironmentVariables {
		vars += fmt.Sprintf("  %s: %s\n", envVar.Name, envVar.Content)
	}
	if vars != "" {
		key := "environments[" + request.Name + ".env]"
		values.Add(key, fmt.Sprintf("vars: \n%s", vars))
		keys = append(keys, key)
	}

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}
	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}
	return response, nil
}
