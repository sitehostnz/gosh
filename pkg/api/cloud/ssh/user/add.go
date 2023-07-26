package user

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Add creates a new SSH user.
func (s *Client) Add(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/ssh/user/add.json"
	keys := []string{
		"client_id",
		"server_name",
		"username",
		"password",
		"containers[]",
		"ssh_keys[]",
		"read_only_config",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("username", request.Username)
	values.Add("password", request.Password)
	values.Add("read_only_config", strconv.Itoa(request.ReadOnlyConfig))

	for _, c := range request.Containers {
		values.Add("containers[]", c)
	}

	for _, k := range request.SSHKeys {
		values.Add("ssh_keys[]", k)
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
