package db

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Delete deletes a cloud database.
func (s *Client) Delete(ctx context.Context, request DeleteRequest) (response DeleteResponse, err error) {
	uri := "cloud/db/delete.json"
	keys := []string{
		"server_name",
		"mysql_host",
		"database",
	}

	values := url.Values{}
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
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
