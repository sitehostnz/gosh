package db

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update updates the database's backup location.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/db/update.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"database",
		"params[container]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("database", request.Database)
	values.Add("params[container]", request.Container)

	req, err := s.client.NewRequest("POST", uri, net.Encode(values, keys))
	if err != nil {
		return response, err
	}

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
