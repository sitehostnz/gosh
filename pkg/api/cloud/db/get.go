package db

import (
	"context"

	"github.com/sitehostnz/gosh/pkg/net"
)

// Get fetches a single database by server, MySQL host and name, via
// "cloud/db/get.json".
//
// All three of ServerName, MySQLHost and Database are required. An
// unknown server is rejected with "Please specify a valid server name"
// rather than returning an empty result, so a successful call always
// carries a database.
//
// MySQLHost is the container name of the database host, as it appears
// in the mysql_host field of [Client.List] — "mariadb0", not a
// hostname you can resolve.
func (s *Client) Get(ctx context.Context, request GetRequest) (response GetResponse, err error) {
	uri := "cloud/db/get.json"
	keys := []string{
		"apikey",
		"client_id",
		"server_name",
		"mysql_host",
		"database",
	}

	req, err := s.client.NewRequest("GET", uri, "")
	if err != nil {
		return response, err
	}

	// apikey and client_id are already on the query from NewRequest;
	// re-adding them here sent client_id twice on every call.
	v := req.URL.Query()
	v.Add("server_name", request.ServerName)
	v.Add("mysql_host", request.MySQLHost)
	v.Add("database", request.Database)

	req.URL.RawQuery = net.Encode(v, keys)

	if err := s.client.Do(ctx, req, &response); err != nil {
		return response, err
	}

	return response, nil
}
