package user

import (
	"context"
	"github.com/sitehostnz/gosh/pkg/utils"
	"net/url"
)

// Update updates the database's backup location.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/db/user/update.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"username",
		"password",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("password", request.Password)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
