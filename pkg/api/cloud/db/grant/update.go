package grant

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Update updates the grants for a specific db/user/host.
func (s *Client) Update(ctx context.Context, request UpdateRequest) (response UpdateResponse, err error) {
	uri := "cloud/db/grant/update.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"database",
		"username",
		"grants[]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("database", request.Database)

	for _, grant := range request.Grants {
		values.Add("grants[]", grant)
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
