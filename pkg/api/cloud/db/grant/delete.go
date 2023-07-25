package grant

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete deletes the grant from the specific database/host/user
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	uri := "cloud/db/grant/delete.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"database",
		"username",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("database", request.Database)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
