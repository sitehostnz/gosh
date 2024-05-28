package user

import (
	"context"
	"net/url"
	"strconv"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Update updates the database's backup location.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/ssh/user/update.json"
	keys := []string{
		"client_id",
		"server_name",
		"username",
		"params[password]",
		"params[containers][]",
		"params[ssh_keys][]",
		"params[volumes][]",
		"params[read_only_config][]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("username", request.Username)
	values.Add("params[password]", request.Password)
	values.Add("params[read_only_config]", strconv.Itoa(utils.BoolToInt(request.ReadOnlyConfig)))

	for _, c := range request.Containers {
		values.Add("params[containers][]", c)
	}

	for _, k := range request.SSHKeys {
		values.Add("params[ssh_keys][]", k)
	}

	for _, k := range request.Volumes {
		values.Add("params[volumes][]", k)
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
