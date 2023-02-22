package stack

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add creates a new cloud stack.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/stack/add.json"
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
	values.Add("enable_ssl", strconv.Itoa(request.EnableSSL))
	values.Add("docker_compose", request.DockerCompose)
	values.Add("label", request.Label)
	values.Add("name", request.Name)

	var vars string
	for _, envVar := range request.Environments {
		vars += fmt.Sprintf("  %s: %s\n", envVar.Name, envVar.Content)
	}

	if vars != "" {
		values.Add("environments["+request.Name+".env]", fmt.Sprintf("vars: \n%s", vars))
		keys = append(keys, "environments["+request.Name+".env]")
	}

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
