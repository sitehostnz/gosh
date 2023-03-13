package db

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/utils"
)

// Get fetches a cloud image.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/db/get.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"database",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	v := req.URL.Query()
	v.Add("client_id", s.client.ClientID)
	v.Add("server_name", request.ServerName)
	v.Add("mysql_host", request.MySQLHost)
	v.Add("database", request.Database)

	req.URL.RawQuery = utils.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
