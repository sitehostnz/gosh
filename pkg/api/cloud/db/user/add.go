package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new cloud database.
func (s *Client) Add(ctx context.Context, request AddRequest) (response AddResponse, err error) {
	uri := "cloud/db/user/add.json"
	keys := []string{
		"client_id",
		"server_name",
		"mysql_host",
		"username",
		"password",
		"database",
		"grants[]",
	}

	values := url.Values{}
	values.Add("client_id", s.client.ClientID)
	values.Add("server_name", request.ServerName)
	values.Add("mysql_host", request.MySQLHost)
	values.Add("username", request.Username)
	values.Add("password", request.Password)

	// grant and database are optional but both must be set
	if request.Grants != nil && request.Database != "" {
		values.Add("database", request.Database)
		for _, g := range request.Grants {
			values.Add("grants[]", g)
		}
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
