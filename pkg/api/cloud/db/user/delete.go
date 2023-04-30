package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete deletes a cloud user database.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	uri := "cloud/db/user/delete.json"
	keys := []string{
		"server_name",
		"mysql_host",
		"username",
	}

	values := url.Values{}
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)

	req, err := s.client.NewRequest("POST", uri, utils.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
