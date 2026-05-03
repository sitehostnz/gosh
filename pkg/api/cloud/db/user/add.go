package user

import (
	"context"
	"net/url"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Add creates a new database user via /cloud/db/user/add.json.
//
// **Username format gotcha** (verified live, May 2026): the API
// rejects names containing underscores or longer than ~16 chars
// with "Please specify a valid username." The exact rule isn't
// documented; the safe shape mirrors the cc<hex> stack-name
// pattern — alphanumeric, no separators, ≤16 chars.
//
// Examples that work: "g0a1b2c3d4e", "ccdeadbeef" (matches stack
// name).
// Examples that fail: "goshu_mariadb_8abd1e", "my_db_user".
//
// Grants takes the lowercased privilege names — typical app
// grant set is "select", "insert", "update", "delete", "create",
// "drop", "index", "alter", "create temporary tables",
// "lock tables", "create view", "show view".
//
// See examples/cloud-db-compare for the working pattern.
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
